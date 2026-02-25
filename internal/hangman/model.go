package hangman

import (
	"fmt"
	"strings"
	"unicode"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type phase int

const (
	phasePlaying phase = iota
	phaseGameOver
)

// Model is the Bubbletea model for the Hangman game.
type Model struct {
	game      *Game
	phase     phase
	width     int
	height    int
	done      bool
	message   string
	HighScore int
}

// New creates a fresh Hangman model.
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
	switch {
	case key == "q" || key == "esc":
		m.done = true
	case len(key) == 1 && unicode.IsLetter(rune(key[0])):
		letter := rune(key[0])
		err := m.game.Guess(letter)
		if err != nil {
			m.message = err.Error()
			return m, nil
		}
		upper := unicode.ToUpper(letter)
		if strings.ContainsRune(m.game.Target(), upper) {
			m.message = fmt.Sprintf("'%c' is in the word!", upper)
		} else {
			m.message = fmt.Sprintf("'%c' is not in the word", upper)
		}
		if m.game.IsOver() {
			m.phase = phaseGameOver
			if m.game.Won() {
				wrong := m.game.WrongGuesses()
				switch {
				case m.HighScore > 0 && wrong < m.HighScore:
					m.message = fmt.Sprintf("You won! The word was %s (%d wrong) — NEW BEST!", m.game.Target(), wrong)
				case m.HighScore > 0:
					m.message = fmt.Sprintf("You won! The word was %s (%d wrong, Best: %d)", m.game.Target(), wrong, m.HighScore)
				default:
					m.message = fmt.Sprintf("You won! The word was %s", m.game.Target())
				}
			} else {
				m.message = fmt.Sprintf("Game over! The word was %s", m.game.Target())
			}
		}
	}
	return m, nil
}

func (m Model) updateGameOver(key string) (tea.Model, tea.Cmd) {
	switch key {
	case "n":
		m.game = NewGame()
		m.phase = phasePlaying
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

// FinalScore returns the number of wrong guesses (lower is better), or -1 if lost.
func (m Model) FinalScore() int {
	if m.game == nil || !m.game.Won() {
		return -1
	}
	return m.game.WrongGuesses()
}

// View renders the complete game screen.
func (m Model) View() string {
	var sections []string

	sections = append(sections,
		titleStyle.Render("H A N G M A N"),
		"",
		renderHangman(m.game.WrongGuesses()),
		"",
		wordStyle.Render(m.game.RevealedWord()),
		"",
		renderUsedLetters(m.game),
		"",
	)

	if m.message != "" {
		switch {
		case m.phase == phaseGameOver && m.game.Won():
			sections = append(sections, winStyle.Render(m.message))
		case m.phase == phaseGameOver:
			sections = append(sections, loseStyle.Render(m.message))
		default:
			sections = append(sections, messageStyle.Render(m.message))
		}
	} else {
		sections = append(sections, "")
	}

	var footer string
	switch m.phase {
	case phasePlaying:
		wrong := m.game.WrongGuesses()
		remaining := m.game.MaxWrong() - wrong
		footer = fmt.Sprintf("A-Z Guess | %d/%d wrong | Q Quit", wrong, m.game.MaxWrong())
		if remaining <= 2 {
			footer = fmt.Sprintf("A-Z Guess | %d/%d wrong (careful!) | Q Quit", wrong, m.game.MaxWrong())
		}
	case phaseGameOver:
		footer = "N New Game | Q Quit"
	}
	sections = append(sections, footerStyle.Render(footer))

	content := lipgloss.JoinVertical(lipgloss.Center, sections...)

	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, content)
}

// renderHangman returns the ASCII hangman figure for the given number of wrong guesses.
func renderHangman(wrong int) string {
	head := " "
	body := " "
	leftArm := " "
	rightArm := " "
	leftLeg := " "
	rightLeg := " "

	if wrong >= 1 {
		head = "O"
	}
	if wrong >= 2 {
		body = "|"
	}
	if wrong >= 3 {
		leftArm = "/"
	}
	if wrong >= 4 {
		rightArm = "\\"
	}
	if wrong >= 5 {
		leftLeg = "/"
	}
	if wrong >= 6 {
		rightLeg = "\\"
	}

	lines := []string{
		"  -----",
		"  |   |",
		"  |   " + head,
		"  |  " + leftArm + body + rightArm,
		"  |  " + leftLeg + " " + rightLeg,
		"  |",
	}

	style := gallowsStyle
	if wrong >= maxWrongGuesses {
		style = gallowsDangerStyle
	}

	return style.Render(strings.Join(lines, "\n"))
}

// renderUsedLetters shows all guessed letters with correct ones highlighted.
func renderUsedLetters(g *Game) string {
	used := g.UsedLetters()
	if len(used) == 0 {
		return dimStyle.Render("No guesses yet")
	}

	parts := make([]string, 0, len(used))
	for _, ch := range used {
		letter := string(ch)
		if strings.ContainsRune(g.Target(), ch) {
			parts = append(parts, correctLetterStyle.Render(letter))
		} else {
			parts = append(parts, wrongLetterStyle.Render(letter))
		}
	}

	return "Used: " + strings.Join(parts, " ")
}

// --- Styles ---

var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#DCFFDC"))

	wordStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("15"))

	gallowsStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("242"))

	gallowsDangerStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#FF0000"))

	correctLetterStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("#00E632"))

	wrongLetterStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("242")).
				Strikethrough(true)

	winStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#00E632"))

	loseStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FF0000"))

	messageStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#DCFFDC"))

	dimStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("242"))

	footerStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("240"))
)
