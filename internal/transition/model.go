package transition

import (
	"math/rand/v2"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/herbhall/cli-play/internal/rain"
)

type tickMsg time.Time

// CellPos is a pre-computed position of a text character on screen.
type CellPos struct {
	X, Y int
	Char rune
}

// Model drives the dissolve-rain-reveal transition between splash and menu.
type Model struct {
	width, height int
	phase         Phase
	grid          *rain.CellGrid
	columns       []*rain.Column
	charPool      *rain.CharPool

	// Pre-computed text positions.
	splashCells []CellPos
	menuCells   []CellPos

	// Dissolve tracking.
	dissolvedCount int
	splashLookup   map[[2]int]bool

	// Reveal tracking.
	depositedCount int
	depositOrder   []int // shuffled indices into menuCells

	// Timing.
	lastTick     time.Time
	phaseElapsed time.Duration

	// Drain control.
	draining bool

	done bool
}

// New creates a transition model that will dissolve splashText into rain
// and reveal menuText by depositing characters.
func New(width, height int, splashText, menuText string) Model {
	grid := rain.NewCellGrid(width, height)
	pool := rain.NewMatrixPool()

	splashCells := computePositions(splashText, width, height)
	menuCells := computePositions(menuText, width, height)

	// Build splash lookup for fast collision detection.
	lookup := make(map[[2]int]bool, len(splashCells))
	for _, cp := range splashCells {
		lookup[[2]int{cp.X, cp.Y}] = true
	}

	// Prepare shuffled deposit order (Fisher-Yates).
	order := make([]int, len(menuCells))
	for i := range order {
		order[i] = i
	}
	rand.Shuffle(len(order), func(i, j int) {
		order[i], order[j] = order[j], order[i]
	})

	// Populate grid with splash text characters.
	for _, cp := range splashCells {
		grid.Set(cp.X, cp.Y, rain.Cell{
			Char:  cp.Char,
			State: rain.CellSplash,
		})
	}

	return Model{
		width:        width,
		height:       height,
		phase:        PhaseDissolve,
		grid:         grid,
		columns:      nil,
		charPool:     pool,
		splashCells:  splashCells,
		menuCells:    menuCells,
		splashLookup: lookup,
		depositOrder: order,
		lastTick:     time.Now(),
	}
}

// computePositions parses text into screen coordinates, centered within
// the given terminal dimensions.
func computePositions(text string, width, height int) []CellPos {
	lines := strings.Split(text, "\n")

	// Find the maximum visual width across all lines.
	maxW := 0
	for _, line := range lines {
		w := visualWidth(line)
		if w > maxW {
			maxW = w
		}
	}

	// Center the block.
	offsetX := (width - maxW) / 2
	offsetY := (height - len(lines)) / 2

	var cells []CellPos
	for row, line := range lines {
		col := 0
		for _, ch := range line {
			if ch != ' ' {
				cells = append(cells, CellPos{
					X:    offsetX + col,
					Y:    offsetY + row,
					Char: ch,
				})
			}
			col += runeWidth(ch)
		}
	}
	return cells
}

// visualWidth returns the display width of a string, accounting for
// wide (CJK/block) characters that take 2 columns.
func visualWidth(s string) int {
	w := 0
	for _, ch := range s {
		w += runeWidth(ch)
	}
	return w
}

// runeWidth returns 2 for wide characters (block elements, CJK, katakana),
// 1 for everything else.
func runeWidth(ch rune) int {
	switch {
	case ch >= 0x2580 && ch <= 0x259F: // block elements (used in title art)
		return 1
	case ch >= 0x2500 && ch <= 0x257F: // box drawing
		return 1
	case ch >= 0xFF01 && ch <= 0xFF60: // fullwidth forms
		return 2
	case ch >= 0xFF61 && ch <= 0xFFDC: // halfwidth katakana/hangul
		return 1
	case ch >= 0x2E80 && ch <= 0x9FFF: // CJK
		return 2
	}
	return 1
}

// Init starts the tick loop.
func (m Model) Init() tea.Cmd {
	return doTick()
}

// Update advances the animation state machine.
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.grid.Resize(m.width, m.height)
		// Recompute positions would require the original text;
		// for now just update dimensions. A resize mid-transition
		// is uncommon and the animation is short.
		return m, nil

	case tea.KeyMsg:
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}
		// Any other key skips the animation.
		m.done = true
		m.phase = PhaseDone
		return m, nil

	case tickMsg:
		return m.tick()
	}

	return m, nil
}

// tick is the per-frame update, called ~30 times per second.
func (m Model) tick() (Model, tea.Cmd) {
	now := time.Now()
	dt := now.Sub(m.lastTick).Seconds()
	if dt > 0.1 {
		dt = 0.1 // cap to prevent jumps after lag
	}
	m.lastTick = now
	m.phaseElapsed += time.Duration(dt * float64(time.Second))

	m.updateColumns(dt)
	m.removeDeadColumns()
	if !m.draining {
		m.spawnColumns()
	}
	m.applyRainToGrid()

	switch m.phase {
	case PhaseDissolve:
		m.dissolve()
		if m.dissolvedCount >= len(m.splashCells) {
			m.phase = PhaseReveal
			m.phaseElapsed = 0
		}
	case PhaseReveal:
		m.reveal()
		if m.depositedCount >= len(m.menuCells) {
			m.phase = PhaseDrain
			m.draining = true
			m.phaseElapsed = 0
		}
	case PhaseDrain:
		if len(m.columns) == 0 {
			m.phase = PhaseDone
			m.done = true
			return m, nil
		}
	case PhaseDone:
		m.done = true
		return m, nil
	}

	return m, doTick()
}

// updateColumns advances every rain column by dt.
func (m *Model) updateColumns(dt float64) {
	for _, col := range m.columns {
		col.Update(dt, m.height, m.charPool)
	}
}

// removeDeadColumns filters out fully-drained columns.
func (m *Model) removeDeadColumns() {
	alive := m.columns[:0]
	for _, col := range m.columns {
		if !col.IsDead() {
			alive = append(alive, col)
		}
	}
	m.columns = alive
}

// spawnColumns probabilistically creates new rain columns.
func (m *Model) spawnColumns() {
	occupied := make(map[int]bool, len(m.columns))
	for _, col := range m.columns {
		occupied[col.X] = true
	}
	for x := 0; x < m.width; x++ {
		if !occupied[x] && rand.Float64() < SpawnRate {
			m.columns = append(m.columns, rain.SpawnColumn(x, m.height))
		}
	}
}

// applyRainToGrid writes rain trail characters to the grid, preserving
// deposited cells. Also stores trail position in FadeStep for gradient
// coloring during View.
func (m *Model) applyRainToGrid() {
	// Clear previous rain cells.
	for y := 0; y < m.grid.Height; y++ {
		for x := 0; x < m.grid.Width; x++ {
			cell := m.grid.Get(x, y)
			if cell.State == rain.CellRain || cell.State == rain.CellDissolving {
				m.grid.Set(x, y, rain.Cell{})
			}
		}
	}

	// Write current trail chars.
	for _, col := range m.columns {
		trailLen := len(col.Trail)
		for i, tc := range col.Trail {
			existing := m.grid.Get(col.X, tc.Y)
			if existing.State == rain.CellDeposited {
				continue // never overwrite deposited
			}

			// Compute trail position: 0 = tail (dimmest), trailLen-1 = head (brightest).
			// Store as 0..100 in FadeStep where 0 = head, 100 = tail.
			posFromHead := trailLen - 1 - i
			var fadePct int
			if trailLen > 1 {
				fadePct = (posFromHead * 100) / (trailLen - 1)
			}

			state := rain.CellRain
			if existing.State == rain.CellSplash {
				state = rain.CellDissolving
			}

			m.grid.Set(col.X, tc.Y, rain.Cell{
				Char:     tc.Char,
				State:    state,
				FadeStep: fadePct,
			})
		}
	}
}

// dissolve marks splash cells as dissolved when rain passes over them.
func (m *Model) dissolve() {
	for _, col := range m.columns {
		for _, tc := range col.Trail {
			key := [2]int{col.X, tc.Y}
			if m.splashLookup[key] {
				delete(m.splashLookup, key)
				m.dissolvedCount++
			}
		}
	}
}

// reveal deposits menu characters at a steady rate.
func (m *Model) reveal() {
	if len(m.menuCells) == 0 {
		return
	}

	// Deposit enough chars to finish in ~2 seconds at 30fps.
	depositsPerTick := len(m.menuCells) / 60
	if depositsPerTick < 1 {
		depositsPerTick = 1
	}

	for i := 0; i < depositsPerTick && m.depositedCount < len(m.menuCells); i++ {
		idx := m.depositOrder[m.depositedCount]
		cp := m.menuCells[idx]
		m.grid.Set(cp.X, cp.Y, rain.Cell{
			Char:       cp.Char,
			State:      rain.CellDeposited,
			TargetChar: cp.Char,
		})
		m.depositedCount++
	}
}

// Styles for rendering.
var (
	splashStyle   = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("15"))
	depositStyle  = lipgloss.NewStyle().Bold(true).Foreground(rain.ColorDeposited)
)

// rainStyle returns a styled string for a rain cell using its FadeStep
// as the trail position (0 = head, 100 = tail).
func rainCellStyle(fadePct int) lipgloss.Style {
	pos := float64(fadePct) / 100.0
	return lipgloss.NewStyle().Foreground(rain.TrailColor(pos))
}

// View renders the full grid to a string.
func (m Model) View() string {
	if m.grid == nil {
		return ""
	}

	var buf strings.Builder
	buf.Grow(m.width * m.height * 4) // rough estimate for UTF-8 + ANSI

	for y := 0; y < m.height; y++ {
		for x := 0; x < m.width; x++ {
			cell := m.grid.Get(x, y)
			switch cell.State {
			case rain.CellEmpty:
				buf.WriteRune(' ')
			case rain.CellSplash:
				buf.WriteString(splashStyle.Render(string(cell.Char)))
			case rain.CellRain, rain.CellDissolving:
				buf.WriteString(rainCellStyle(cell.FadeStep).Render(string(cell.Char)))
			case rain.CellDeposited:
				buf.WriteString(depositStyle.Render(string(cell.Char)))
			default:
				buf.WriteRune(' ')
			}
		}
		if y < m.height-1 {
			buf.WriteRune('\n')
		}
	}
	return buf.String()
}

// Done returns true when the transition animation has completed
// or been skipped.
func (m Model) Done() bool {
	return m.done
}

// doTick returns a tea.Cmd that fires a tickMsg after TickInterval.
func doTick() tea.Cmd {
	return tea.Tick(TickInterval, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}
