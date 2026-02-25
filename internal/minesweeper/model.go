package minesweeper

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type phase int

const (
	phaseDifficulty phase = iota
	phasePlaying
	phaseGameOver
)

type tickMsg struct{}

func tickCmd() tea.Cmd {
	return tea.Tick(time.Second, func(time.Time) tea.Msg {
		return tickMsg{}
	})
}

// Model is the Bubbletea model for the Minesweeper game.
type Model struct {
	game      *Game
	cursorRow int
	cursorCol int
	width     int
	height    int
	done      bool
	phase     phase
	elapsed   int
	ticking   bool
	diff      Difficulty
}

// New creates a fresh Minesweeper model at the difficulty selection screen.
func New() Model {
	return Model{
		phase: phaseDifficulty,
	}
}

// Init returns nil; no initial command needed.
func (m Model) Init() tea.Cmd {
	return nil
}

// Done returns true when the player wants to exit to the menu.
func (m Model) Done() bool {
	return m.done
}

// Update handles input and advances game state.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tickMsg:
		if m.phase == phasePlaying && m.ticking && m.game.State == Playing {
			m.elapsed++
			return m, tickCmd()
		}
		return m, nil

	case tea.KeyMsg:
		key := msg.String()

		if key == "ctrl+c" {
			return m, tea.Quit
		}

		switch m.phase {
		case phaseDifficulty:
			return m.updateDifficulty(key)
		case phasePlaying:
			return m.updatePlaying(key)
		case phaseGameOver:
			return m.updateGameOver(key)
		}
	}

	return m, nil
}

func (m Model) updateDifficulty(key string) (tea.Model, tea.Cmd) {
	switch key {
	case "1":
		return m.startGame(Beginner)
	case "2":
		return m.startGame(Intermediate)
	case "3":
		return m.startGame(Expert)
	case "q", "esc":
		m.done = true
	}
	return m, nil
}

func (m Model) startGame(diff Difficulty) (tea.Model, tea.Cmd) {
	m.diff = diff
	m.game = NewGame(diff)
	m.phase = phasePlaying
	m.cursorRow = 0
	m.cursorCol = 0
	m.elapsed = 0
	m.ticking = false
	return m, nil
}

func (m Model) updatePlaying(key string) (tea.Model, tea.Cmd) {
	switch key {
	case "up", "k":
		if m.cursorRow > 0 {
			m.cursorRow--
		}
	case "down", "j":
		if m.cursorRow < m.game.Rows-1 {
			m.cursorRow++
		}
	case "left", "h":
		if m.cursorCol > 0 {
			m.cursorCol--
		}
	case "right", "l":
		if m.cursorCol < m.game.Cols-1 {
			m.cursorCol++
		}
	case "enter", " ":
		if m.game.State != Playing {
			return m, nil
		}
		wasFirstClick := m.game.FirstClick
		m.game.Reveal(m.cursorRow, m.cursorCol)
		if wasFirstClick && !m.game.FirstClick {
			m.ticking = true
			if m.game.State == Playing {
				return m, tickCmd()
			}
		}
		if m.game.State != Playing {
			m.ticking = false
			m.phase = phaseGameOver
		}
	case "f":
		if m.game.State == Playing {
			m.game.ToggleFlag(m.cursorRow, m.cursorCol)
		}
	case "n":
		return m.startGame(m.diff)
	case "q", "esc":
		m.done = true
	}
	return m, nil
}

func (m Model) updateGameOver(key string) (tea.Model, tea.Cmd) {
	switch key {
	case "n":
		return m.startGame(m.diff)
	case "d":
		m.phase = phaseDifficulty
		m.game = nil
	case "q", "esc":
		m.done = true
	}
	return m, nil
}

// View renders the complete game screen.
func (m Model) View() string {
	switch m.phase {
	case phaseDifficulty:
		return m.viewDifficulty()
	case phasePlaying, phaseGameOver:
		return m.viewGame()
	}
	return ""
}

func (m Model) viewDifficulty() string {
	var sections []string

	sections = append(sections,
		titleStyle.Render("M I N E S W E E P E R"),
		"",
		headerStyle.Render("Select Difficulty"),
		"",
		optionStyle.Render("  [1]  Beginner      9 x 9    10 mines"),
		optionStyle.Render("  [2]  Intermediate  16 x 16  40 mines"),
		optionStyle.Render("  [3]  Expert        16 x 30  99 mines"),
		"",
		footerStyle.Render("Q Quit"),
	)

	content := lipgloss.JoinVertical(lipgloss.Center, sections...)
	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, content)
}

func (m Model) viewGame() string {
	if m.game == nil {
		return ""
	}

	var sections []string

	// Title with difficulty
	diffNames := map[Difficulty]string{
		Beginner:     "Beginner",
		Intermediate: "Intermediate",
		Expert:       "Expert",
	}
	title := titleStyle.Render(fmt.Sprintf("Minesweeper - %s", diffNames[m.diff]))
	sections = append(sections, title, "")

	// Status bar
	remaining := m.game.TotalMines - m.game.FlagsUsed
	status := statusStyle.Render(fmt.Sprintf("Mines: %d  Flags: %d  Time: %d", remaining, m.game.FlagsUsed, m.elapsed))
	sections = append(sections, status, "",
		m.renderGrid(), "",
	)

	// Game over message
	if m.phase == phaseGameOver {
		switch m.game.State {
		case Won:
			sections = append(sections, winStyle.Render("YOU WIN!"))
		case Lost:
			sections = append(sections, loseStyle.Render("GAME OVER - Mine hit!"))
		}
		sections = append(sections, "")
	}

	// Footer
	var footer string
	if m.phase == phaseGameOver {
		footer = "N New Game | D Difficulty | Q Quit"
	} else {
		footer = "Arrows Move | Enter Reveal | F Flag | N New | Q Quit"
	}
	sections = append(sections, footerStyle.Render(footer))

	content := lipgloss.JoinVertical(lipgloss.Center, sections...)
	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, content)
}

func (m Model) renderGrid() string {
	var rows []string

	for r := 0; r < m.game.Rows; r++ {
		var cells []string
		for c := 0; c < m.game.Cols; c++ {
			cell := m.game.Grid[r][c]
			isCursor := r == m.cursorRow && c == m.cursorCol

			text := m.renderCell(cell)
			style := m.cellStyle(cell, isCursor)
			cells = append(cells, style.Render(text))
		}
		rows = append(rows, strings.Join(cells, ""))
	}

	return strings.Join(rows, "\n")
}

func (m Model) renderCell(cell Cell) string {
	switch cell.State {
	case Hidden:
		return "##"
	case Flagged:
		return "FF"
	case Revealed:
		if cell.Mine {
			return "* "
		}
		if cell.Adjacent == 0 {
			return "  "
		}
		return fmt.Sprintf("%d ", cell.Adjacent)
	}
	return "##"
}

func (m Model) cellStyle(cell Cell, isCursor bool) lipgloss.Style {
	base := lipgloss.NewStyle().Width(2)

	if isCursor && m.phase == phasePlaying {
		return base.
			Background(lipgloss.Color("#444444")).
			Bold(true).
			Foreground(m.cellForeground(cell))
	}

	return base.Foreground(m.cellForeground(cell))
}

func (m Model) cellForeground(cell Cell) lipgloss.Color {
	switch cell.State {
	case Hidden:
		return lipgloss.Color("#808080")
	case Flagged:
		return lipgloss.Color("#FF0000")
	case Revealed:
		if cell.Mine {
			return lipgloss.Color("#FF0000")
		}
		return numberColor(cell.Adjacent)
	}
	return lipgloss.Color("#808080")
}

func numberColor(n int) lipgloss.Color {
	switch n {
	case 1:
		return lipgloss.Color("#0000FF")
	case 2:
		return lipgloss.Color("#008200")
	case 3:
		return lipgloss.Color("#FF0000")
	case 4:
		return lipgloss.Color("#000084")
	case 5:
		return lipgloss.Color("#840000")
	case 6:
		return lipgloss.Color("#008284")
	case 7:
		return lipgloss.Color("#840084")
	case 8:
		return lipgloss.Color("#808080")
	default:
		return lipgloss.Color("#FFFFFF")
	}
}

// --- Styles ---

var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("15"))

	headerStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("15")).
			Underline(true)

	statusStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("242"))

	optionStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#00E632"))

	footerStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("240"))

	winStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#00E632"))

	loseStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FF0000"))
)
