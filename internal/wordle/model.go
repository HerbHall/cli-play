package wordle

import (
	"fmt"
	"strings"
	"unicode"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type phase int

const (
	phaseTyping phase = iota
	phaseGameOver
)

// Model is the Bubbletea model for the Wordle game.
type Model struct {
	game         *Game
	currentGuess string
	phase        phase
	width        int
	height       int
	done         bool
	message      string
	HighScore    int
}

// New creates a fresh Wordle model.
func New() Model {
	return Model{
		game:  NewGame(),
		phase: phaseTyping,
	}
}

// Init returns nil; no initial command needed.
func (m Model) Init() tea.Cmd {
	return nil
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
		case phaseTyping:
			return m.updateTyping(key)
		case phaseGameOver:
			return m.updateGameOver(key)
		}
	}

	return m, nil
}

func (m Model) updateTyping(key string) (tea.Model, tea.Cmd) {
	switch key {
	case "backspace":
		if m.currentGuess != "" {
			m.currentGuess = m.currentGuess[:len(m.currentGuess)-1]
			m.message = ""
		}
	case "enter":
		if len(m.currentGuess) < 5 {
			m.message = "Not enough letters"
			return m, nil
		}
		_, err := m.game.Guess(m.currentGuess)
		if err != nil {
			m.message = err.Error()
			return m, nil
		}
		m.currentGuess = ""
		m.message = ""
		if m.game.IsOver() {
			m.phase = phaseGameOver
			if m.game.Won {
				guesses := len(m.game.Guesses)
				if m.HighScore > 0 && guesses < m.HighScore {
					m.message = fmt.Sprintf("Correct in %d/6! %s — NEW HIGH SCORE!", guesses, m.game.Target)
				} else if m.HighScore > 0 {
					m.message = fmt.Sprintf("Correct in %d/6! %s (Best: %d/6)", guesses, m.game.Target, m.HighScore)
				} else {
					m.message = fmt.Sprintf("Correct! The word was %s", m.game.Target)
				}
			} else {
				m.message = fmt.Sprintf("The word was %s", m.game.Target)
			}
		}
	case "q", "esc":
		m.done = true
	default:
		if len(key) == 1 && unicode.IsLetter(rune(key[0])) && len(m.currentGuess) < 5 {
			m.currentGuess += strings.ToUpper(key)
			m.message = ""
		}
	}
	return m, nil
}

func (m Model) updateGameOver(key string) (tea.Model, tea.Cmd) {
	switch key {
	case "n":
		m.game = NewGame()
		m.phase = phaseTyping
		m.currentGuess = ""
		m.message = ""
	case "q", "esc":
		m.done = true
	}
	return m, nil
}

// Done returns true when the player wants to exit.
func (m Model) Done() bool {
	return m.done
}

// FinalScore returns the number of guesses used, or -1 if the game wasn't won.
func (m Model) FinalScore() int {
	if m.game == nil || !m.game.Won {
		return -1
	}
	return len(m.game.Guesses)
}

// View renders the complete game screen.
func (m Model) View() string {
	var sections []string

	sections = append(sections,
		titleStyle.Render("W O R D L E"),
		"",
		m.renderGrid(),
		"",
		m.renderKeyboard(),
		"",
	)

	if m.message != "" {
		var msgStyled string
		switch {
		case m.game.Won:
			msgStyled = winStyle.Render(m.message)
		case m.game.Over:
			msgStyled = lossStyle.Render(m.message)
		default:
			msgStyled = messageStyle.Render(m.message)
		}
		sections = append(sections, msgStyled)
	} else {
		sections = append(sections, "")
	}

	var footer string
	switch m.phase {
	case phaseTyping:
		footer = "Type a word | Enter Submit | Q Quit"
	case phaseGameOver:
		footer = "N New Game | Q Quit"
	}
	sections = append(sections, footerStyle.Render(footer))

	content := lipgloss.JoinVertical(lipgloss.Center, sections...)

	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, content)
}

// renderGrid draws the 6-row guess grid with colored cells.
func (m Model) renderGrid() string {
	var rows []string

	for i := 0; i < m.game.MaxGuesses; i++ {
		switch {
		case i < len(m.game.Guesses):
			rows = append(rows, m.renderGuessRow(m.game.Guesses[i], m.game.Results[i]))
		case i == len(m.game.Guesses) && m.phase == phaseTyping:
			rows = append(rows, m.renderCurrentRow())
		default:
			rows = append(rows, m.renderEmptyRow())
		}
	}

	return strings.Join(rows, "\n")
}

// renderGuessRow renders a completed guess with colored backgrounds.
func (m Model) renderGuessRow(guess string, result GuessResult) string {
	cells := make([]string, 0, len(guess))
	for i, ch := range guess {
		letter := string(ch)
		var cell string
		switch result[i] {
		case Correct:
			cell = correctCellStyle.Render(" " + letter + " ")
		case Present:
			cell = presentCellStyle.Render(" " + letter + " ")
		default:
			cell = absentCellStyle.Render(" " + letter + " ")
		}
		cells = append(cells, cell)
	}
	return strings.Join(cells, " ")
}

// renderCurrentRow renders the row being typed with cursor indicators.
func (m Model) renderCurrentRow() string {
	var cells []string
	for i := 0; i < 5; i++ {
		if i < len(m.currentGuess) {
			letter := string(m.currentGuess[i])
			cells = append(cells, inputCellStyle.Render(" "+letter+" "))
		} else {
			cells = append(cells, emptyCellStyle.Render("   "))
		}
	}
	return strings.Join(cells, " ")
}

// renderEmptyRow renders a blank row placeholder.
func (m Model) renderEmptyRow() string {
	cells := make([]string, 0, 5)
	for range 5 {
		cells = append(cells, emptyCellStyle.Render("   "))
	}
	return strings.Join(cells, " ")
}

// renderKeyboard draws the on-screen keyboard with per-key coloring.
func (m Model) renderKeyboard() string {
	rows := []string{"QWERTYUIOP", "ASDFGHJKL", "ZXCVBNM"}
	lines := make([]string, 0, len(rows))

	for i, row := range rows {
		var keys []string
		for _, ch := range row {
			letter := string(ch)
			state, exists := m.game.KeyboardState[ch]
			var key string
			if !exists {
				key = untestedKeyStyle.Render(" " + letter + " ")
			} else {
				switch state {
				case Correct:
					key = correctKeyStyle.Render(" " + letter + " ")
				case Present:
					key = presentKeyStyle.Render(" " + letter + " ")
				default:
					key = absentKeyStyle.Render(" " + letter + " ")
				}
			}
			keys = append(keys, key)
		}
		line := strings.Join(keys, "")
		// Indent the second and third rows for staggered keyboard look.
		if i == 1 {
			line = " " + line
		} else if i == 2 {
			line = "  " + line
		}
		lines = append(lines, line)
	}

	return strings.Join(lines, "\n")
}

// --- Styles ---

var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("15"))

	correctCellStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("15")).
				Background(lipgloss.Color("#538D4E"))

	presentCellStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("15")).
				Background(lipgloss.Color("#B59F3B"))

	absentCellStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("15")).
			Background(lipgloss.Color("#3A3A3C"))

	inputCellStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("15")).
			Border(lipgloss.NormalBorder()).
			BorderForeground(lipgloss.Color("242"))

	emptyCellStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("240")).
			Border(lipgloss.NormalBorder()).
			BorderForeground(lipgloss.Color("238"))

	correctKeyStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("15")).
			Background(lipgloss.Color("#538D4E"))

	presentKeyStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("15")).
			Background(lipgloss.Color("#B59F3B"))

	absentKeyStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("240")).
			Background(lipgloss.Color("#3A3A3C"))

	untestedKeyStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("15")).
				Background(lipgloss.Color("#818384"))

	winStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#538D4E"))

	lossStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FF4444"))

	messageStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#DCFFDC"))

	footerStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("240"))
)
