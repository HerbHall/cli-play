package fifteenpuzzle

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type phase int

const (
	phasePlaying  phase = iota
	phaseGameOver       // puzzle solved
)

type tickMsg struct{}

func tickCmd() tea.Cmd {
	return tea.Tick(time.Second, func(time.Time) tea.Msg {
		return tickMsg{}
	})
}

// Model is the Bubbletea model for the Fifteen Puzzle game.
type Model struct {
	game      *Game
	phase     phase
	width     int
	height    int
	done      bool
	elapsed   int
	ticking   bool
	started   bool
	HighScore int
}

// New creates a fresh Fifteen Puzzle model.
func New() Model {
	return Model{
		game:  NewGame(),
		phase: phasePlaying,
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

// FinalScore returns the number of moves, or -1 if not yet solved.
func (m Model) FinalScore() int {
	if m.game == nil || !m.game.Won {
		return -1
	}
	return m.game.Moves
}

// Update handles input and advances game state.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tickMsg:
		if m.phase == phasePlaying && m.ticking {
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
		case phasePlaying:
			return m.updatePlaying(key)
		case phaseGameOver:
			return m.updateGameOver(key)
		}
	}

	return m, nil
}

func (m Model) updatePlaying(key string) (tea.Model, tea.Cmd) {
	var moved bool

	switch key {
	case "up", "k":
		moved = m.game.MoveUp()
	case "down", "j":
		moved = m.game.MoveDown()
	case "left", "h":
		moved = m.game.MoveLeft()
	case "right", "l":
		moved = m.game.MoveRight()
	case "n":
		m.game.Reset()
		m.elapsed = 0
		m.ticking = false
		m.started = false
		m.phase = phasePlaying
		return m, nil
	case "q", "esc":
		m.done = true
		return m, nil
	}

	if moved && !m.started {
		m.started = true
		m.ticking = true
		return m, tickCmd()
	}

	if m.game.Won {
		m.ticking = false
		m.phase = phaseGameOver
	}

	return m, nil
}

func (m Model) updateGameOver(key string) (tea.Model, tea.Cmd) {
	switch key {
	case "n":
		m.game.Reset()
		m.phase = phasePlaying
		m.elapsed = 0
		m.ticking = false
		m.started = false
	case "q", "esc":
		m.done = true
	}
	return m, nil
}

// View renders the complete game screen.
func (m Model) View() string {
	sections := make([]string, 0, 10)

	sections = append(sections,
		titleStyle.Render("F I F T E E N   P U Z Z L E"),
		"",
	)

	// Status line: moves and time
	mins := m.elapsed / 60
	secs := m.elapsed % 60
	status := statusStyle.Render(fmt.Sprintf("Moves: %d  Time: %d:%02d", m.game.Moves, mins, secs))
	sections = append(sections, status, "", m.renderGrid(), "")

	// Game over message
	if m.phase == phaseGameOver {
		msg := fmt.Sprintf("Solved in %d moves! Time: %d:%02d", m.game.Moves, mins, secs)
		switch {
		case m.HighScore > 0 && m.game.Moves < m.HighScore:
			msg += " -- NEW BEST!"
		case m.HighScore > 0:
			msg += fmt.Sprintf(" (Best: %d moves)", m.HighScore)
		}
		sections = append(sections, winStyle.Render(msg), "")
	}

	// Footer
	var footer string
	switch m.phase {
	case phasePlaying:
		footer = "Arrow/HJKL Slide | N New | Q Quit"
	case phaseGameOver:
		footer = "N New Game | Q Quit"
	}
	sections = append(sections, footerStyle.Render(footer))

	content := lipgloss.JoinVertical(lipgloss.Center, sections...)
	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, content)
}

func (m Model) renderGrid() string {
	rows := make([]string, 0, boardSize*2+1)

	// Top border
	topBorder := gridBorderStyle.Render("+" + strings.Repeat("------+", boardSize))
	rows = append(rows, topBorder)

	for r := range boardSize {
		cells := make([]string, 0, boardSize)
		for c := range boardSize {
			val := m.game.Board[r][c]
			cells = append(cells, renderTile(val))
		}
		row := gridBorderStyle.Render("|") +
			strings.Join(cells, gridBorderStyle.Render("|")) +
			gridBorderStyle.Render("|")
		rows = append(rows, row)

		// Row separator
		sep := gridBorderStyle.Render("+" + strings.Repeat("------+", boardSize))
		rows = append(rows, sep)
	}

	return strings.Join(rows, "\n")
}

func renderTile(val int) string {
	if val == 0 {
		return emptyStyle.Render("      ")
	}
	label := fmt.Sprintf("%d", val)
	pad := 6 - len(label)
	left := pad / 2
	right := pad - left
	text := strings.Repeat(" ", left) + label + strings.Repeat(" ", right)
	return tileStyle.Render(text)
}

// --- Styles ---

var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#DCFFDC"))

	statusStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#DCFFDC"))

	tileStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FFD700"))

	emptyStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("242"))

	gridBorderStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("242"))

	winStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#00E632"))

	footerStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("240"))
)
