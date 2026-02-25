package tictactoe

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type phase int

const (
	phasePlaying phase = iota
	phaseGameOver
)

// Model is the Bubbletea model for the Tic-Tac-Toe game.
type Model struct {
	game      *Game
	cursorRow int
	cursorCol int
	phase     phase
	message   string
	width     int
	height    int
	done      bool
	HighScore int
}

// New creates a fresh Tic-Tac-Toe model.
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

// FinalScore returns 1 for win, 0 for draw, -1 for loss.
func (m Model) FinalScore() int {
	if m.game == nil {
		return 0
	}
	switch m.game.GameState() {
	case Won:
		return 1
	case Lost:
		return -1
	default:
		return 0
	}
}

// Update handles input and advances game state.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
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
	switch key {
	case "up", "k":
		if m.cursorRow > 0 {
			m.cursorRow--
		}
	case "down", "j":
		if m.cursorRow < 2 {
			m.cursorRow++
		}
	case "left", "h":
		if m.cursorCol > 0 {
			m.cursorCol--
		}
	case "right", "l":
		if m.cursorCol < 2 {
			m.cursorCol++
		}
	case "enter", " ":
		err := m.game.Move(m.cursorRow, m.cursorCol)
		if err != nil {
			m.message = err.Error()
			return m, nil
		}
		if m.game.IsOver() {
			m.phase = phaseGameOver
			m.message = m.resultMessage()
			return m, nil
		}
		m.game.AIMove()
		if m.game.IsOver() {
			m.phase = phaseGameOver
			m.message = m.resultMessage()
			return m, nil
		}
		m.message = "Your turn"
	case "n":
		m.game = NewGame()
		m.phase = phasePlaying
		m.cursorRow = 0
		m.cursorCol = 0
		m.message = ""
	case "q", "esc":
		m.done = true
	}
	return m, nil
}

func (m Model) updateGameOver(key string) (tea.Model, tea.Cmd) {
	switch key {
	case "n", "enter":
		m.game = NewGame()
		m.phase = phasePlaying
		m.cursorRow = 0
		m.cursorCol = 0
		m.message = ""
	case "q", "esc":
		m.done = true
	}
	return m, nil
}

func (m Model) resultMessage() string {
	switch m.game.GameState() {
	case Won:
		return "You win!"
	case Lost:
		return "AI wins!"
	case Draw:
		return "It's a draw!"
	default:
		return ""
	}
}

// View renders the complete game screen.
func (m Model) View() string {
	var sections []string

	// Title and board
	sections = append(sections, titleStyle.Render("T I C - T A C - T O E"), "", m.renderBoard(), "")

	// Status message
	switch {
	case m.phase == phaseGameOver && m.game.GameState() == Won:
		sections = append(sections, winStyle.Render(m.message))
	case m.phase == phaseGameOver && m.game.GameState() == Lost:
		sections = append(sections, loseStyle.Render(m.message))
	case m.phase == phaseGameOver && m.game.GameState() == Draw:
		sections = append(sections, drawStyle.Render(m.message))
	case m.phase == phaseGameOver:
		sections = append(sections, messageStyle.Render(m.message))
	case m.message != "":
		sections = append(sections, messageStyle.Render(m.message))
	default:
		sections = append(sections, messageStyle.Render("Your turn — place X"))
	}
	sections = append(sections, "")

	// Footer controls
	var footer string
	if m.phase == phaseGameOver {
		footer = "N New Game  |  Q Quit"
	} else {
		footer = "Arrows Move  |  Enter Place  |  N New  |  Q Quit"
	}
	sections = append(sections, footerStyle.Render(footer))

	content := lipgloss.JoinVertical(lipgloss.Center, sections...)
	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, content)
}

func (m Model) renderBoard() string {
	board := m.game.Board()
	var rows []string

	for r := 0; r < 3; r++ {
		var cells []string
		for c := 0; c < 3; c++ {
			cell := board[r][c]
			isCursor := r == m.cursorRow && c == m.cursorCol && m.phase == phasePlaying
			cells = append(cells, m.renderCell(cell, isCursor))
		}
		rows = append(rows, strings.Join(cells, gridStyle.Render(" │ ")))

		if r < 2 {
			rows = append(rows, gridStyle.Render("─── ┼ ─── ┼ ───"))
		}
	}

	return strings.Join(rows, "\n")
}

func (m Model) renderCell(cell Cell, isCursor bool) string {
	var text string
	var style lipgloss.Style

	switch cell {
	case X:
		text = " X "
		style = xStyle
	case O:
		text = " O "
		style = oStyle
	default:
		if isCursor {
			text = " · "
		} else {
			text = "   "
		}
		style = emptyStyle
	}

	if isCursor {
		style = style.Background(lipgloss.Color("#333333"))
	}

	return style.Render(text)
}

// --- Styles ---

var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#DCFFDC"))

	xStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#FFD700"))

	oStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#FF0000"))

	emptyStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("242"))

	gridStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("242"))

	messageStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#DCFFDC"))

	winStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#00E632"))

	loseStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FF0000"))

	drawStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("242"))

	footerStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("240"))
)
