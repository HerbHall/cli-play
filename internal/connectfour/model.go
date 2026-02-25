package connectfour

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type phase int

const (
	phasePlaying  phase = iota
	phaseGameOver
)

// Model is the Bubbletea model for the Connect Four game.
type Model struct {
	game      *Game
	cursor    int
	phase     phase
	message   string
	width     int
	height    int
	done      bool
	HighScore int
}

// New creates a fresh Connect Four model.
func New() Model {
	return Model{
		game:   NewGame(),
		cursor: 3, // start at center column
		phase:  phasePlaying,
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

// FinalScore returns 1 for win, 0 for draw, -1 for loss/incomplete.
func (m Model) FinalScore() int {
	if m.game == nil {
		return -1
	}
	switch m.game.GameState() {
	case Won:
		return 1
	case Draw:
		return 0
	default:
		return -1
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
	case "left", "h":
		if m.cursor > 0 {
			m.cursor--
		}
	case "right", "l":
		if m.cursor < Cols-1 {
			m.cursor++
		}
	case "enter", " ":
		_, err := m.game.DropDisc(m.cursor)
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
		m.cursor = 3
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
		m.cursor = 3
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

	sections = append(sections,
		titleStyle.Render("C O N N E C T   F O U R"), "",
		m.renderSelector(),
		m.renderBoard(),
		m.renderColumnNumbers(), "",
	)

	// Status message
	switch {
	case m.phase == phaseGameOver && m.game.GameState() == Won:
		sections = append(sections, winStyle.Render(m.message))
	case m.phase == phaseGameOver && m.game.GameState() == Lost:
		sections = append(sections, loseStyle.Render(m.message))
	case m.phase == phaseGameOver && m.game.GameState() == Draw:
		sections = append(sections, drawStyle.Render(m.message))
	case m.message != "":
		sections = append(sections, messageStyle.Render(m.message))
	default:
		sections = append(sections, messageStyle.Render("Your turn — drop a red disc"))
	}
	sections = append(sections, "")

	// Footer controls
	var footer string
	if m.phase == phaseGameOver {
		footer = "N New Game  |  Q Quit"
	} else {
		footer = "←→ Move  |  Enter Drop  |  N New  |  Q Quit"
	}
	sections = append(sections, footerStyle.Render(footer))

	content := lipgloss.JoinVertical(lipgloss.Center, sections...)
	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, content)
}

func (m Model) renderSelector() string {
	parts := make([]string, 0, Cols)
	for c := 0; c < Cols; c++ {
		if c == m.cursor && m.phase == phasePlaying {
			parts = append(parts, redDiscStyle.Render(" ▼ "))
		} else {
			parts = append(parts, "   ")
		}
	}
	return strings.Join(parts, "")
}

func (m Model) renderBoard() string {
	board := m.game.Board()
	rows := make([]string, 0, Rows*2+2)

	// Top border
	top := "╔"
	for c := 0; c < Cols; c++ {
		top += "═══"
		if c < Cols-1 {
			top += "╤"
		}
	}
	top += "╗"
	rows = append(rows, gridStyle.Render(top))

	// Board rows
	for r := 0; r < Rows; r++ {
		row := gridStyle.Render("║")
		for c := 0; c < Cols; c++ {
			cell := board[r][c]
			row += m.renderCell(cell)
			if c < Cols-1 {
				row += gridStyle.Render("│")
			}
		}
		row += gridStyle.Render("║")
		rows = append(rows, row)

		if r < Rows-1 {
			sep := "╟"
			for c := 0; c < Cols; c++ {
				sep += "───"
				if c < Cols-1 {
					sep += "┼"
				}
			}
			sep += "╢"
			rows = append(rows, gridStyle.Render(sep))
		}
	}

	// Bottom border
	bot := "╚"
	for c := 0; c < Cols; c++ {
		bot += "═══"
		if c < Cols-1 {
			bot += "╧"
		}
	}
	bot += "╝"
	rows = append(rows, gridStyle.Render(bot))

	return strings.Join(rows, "\n")
}

func (m Model) renderCell(cell Cell) string {
	switch cell {
	case Red:
		return redDiscStyle.Render(" ● ")
	case Yellow:
		return yellowDiscStyle.Render(" ● ")
	default:
		return emptyStyle.Render(" ○ ")
	}
}

func (m Model) renderColumnNumbers() string {
	parts := make([]string, 0, Cols)
	for c := 0; c < Cols; c++ {
		parts = append(parts, fmt.Sprintf(" %d ", c+1))
	}
	return dimStyle.Render(" " + strings.Join(parts, " "))
}

// --- Styles ---

var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#DCFFDC"))

	redDiscStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FF0000"))

	yellowDiscStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FFD700"))

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

	dimStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("242"))

	footerStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("240"))
)
