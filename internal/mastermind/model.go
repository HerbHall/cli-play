package mastermind

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type phase int

const (
	phasePlaying phase = iota
	phaseGameOver
)

// Color names for the legend display.
var colorNames = [NumColors + 1]string{
	0: "",
	1: "Red",
	2: "Green",
	3: "Blue",
	4: "Yellow",
	5: "Purple",
	6: "Orange",
}

// Model is the Bubbletea model for the Mastermind game.
type Model struct {
	game      *Game
	input     [CodeLength]int // current guess being built (0 = empty)
	cursor    int             // which slot the cursor is on (0-3)
	phase     phase
	width     int
	height    int
	done      bool
	message   string
	HighScore int
}

// New creates a fresh Mastermind model.
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
	switch key {
	case "1", "2", "3", "4", "5", "6":
		color := int(key[0] - '0')
		m.input[m.cursor] = color
		if m.cursor < CodeLength-1 {
			m.cursor++
		}
		m.message = ""

	case "left", "h":
		if m.cursor > 0 {
			m.cursor--
		}

	case "right", "l":
		if m.cursor < CodeLength-1 {
			m.cursor++
		}

	case "backspace", "delete":
		m.input[m.cursor] = 0
		if m.cursor > 0 {
			m.cursor--
		}
		m.message = ""

	case "enter":
		if !m.inputComplete() {
			m.message = "Fill all 4 slots before submitting"
			return m, nil
		}
		fb := m.game.Guess(m.input)
		m.input = [CodeLength]int{}
		m.cursor = 0

		if m.game.Won() {
			m.phase = phaseGameOver
			count := m.game.GuessCount()
			msg := fmt.Sprintf("You cracked the code in %d guess", count)
			if count != 1 {
				msg += "es"
			}
			msg += "!"
			if m.HighScore > 0 && count < m.HighScore {
				msg += " NEW BEST!"
			}
			m.message = msg
		} else if m.game.IsOver() {
			m.phase = phaseGameOver
			secret := m.game.Secret()
			m.message = fmt.Sprintf("Out of guesses! The code was %s",
				formatCode(secret))
		} else {
			m.message = fmt.Sprintf("%d exact, %d misplaced",
				fb.Exact, fb.Misplaced)
		}

	case "q", "esc":
		m.done = true
	}

	return m, nil
}

func (m Model) updateGameOver(key string) (tea.Model, tea.Cmd) {
	switch key {
	case "n":
		m.game = NewGame()
		m.phase = phasePlaying
		m.input = [CodeLength]int{}
		m.cursor = 0
		m.message = ""
	case "q", "esc":
		m.done = true
	}
	return m, nil
}

func (m Model) inputComplete() bool {
	for _, v := range m.input {
		if v == 0 {
			return false
		}
	}
	return true
}

// Done returns true when the player wants to exit.
func (m Model) Done() bool {
	return m.done
}

// FinalScore returns the number of guesses used (lower is better).
// Returns -1 if the player lost.
func (m Model) FinalScore() int {
	if m.game == nil || !m.game.IsOver() {
		return -1
	}
	if !m.game.Won() {
		return -1
	}
	return m.game.GuessCount()
}

// View renders the complete game screen.
func (m Model) View() string {
	var sections []string

	// Title
	sections = append(sections, titleStyle.Render("M A S T E R M I N D"))

	// Color legend
	sections = append(sections, m.renderLegend())
	sections = append(sections, "")

	// Guess rows
	sections = append(sections, m.renderBoard()...)
	sections = append(sections, "")

	// Message
	if m.message != "" {
		if m.phase == phaseGameOver && m.game.Won() {
			sections = append(sections, winStyle.Render(m.message))
		} else if m.phase == phaseGameOver {
			sections = append(sections, loseStyle.Render(m.message))
		} else {
			sections = append(sections, messageStyle.Render(m.message))
		}
	} else {
		sections = append(sections, "")
	}

	// Footer
	var footer string
	switch m.phase {
	case phasePlaying:
		footer = "1-6 Set color  |  ←→ Move  |  Backspace Clear  |  Enter Submit  |  Q Quit"
	case phaseGameOver:
		footer = "N New Game  |  Q Quit"
	}
	sections = append(sections, footerStyle.Render(footer))

	content := lipgloss.JoinVertical(lipgloss.Center, sections...)

	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, content)
}

func (m Model) renderLegend() string {
	var parts []string
	for i := 1; i <= NumColors; i++ {
		style := colorStyle(i)
		parts = append(parts, style.Render(fmt.Sprintf("%d=%s", i, colorNames[i])))
	}
	return legendStyle.Render(strings.Join(parts, "  "))
}

func (m Model) renderBoard() []string {
	rows := make([]string, 0, MaxGuesses)
	guesses := m.game.Guesses()

	// Previous guesses with feedback
	for _, entry := range guesses {
		rows = append(rows, m.renderGuessRow(entry))
	}

	// Current input row (only when still playing)
	if m.phase == phasePlaying {
		rows = append(rows, m.renderInputRow())
	}

	// Empty rows for remaining guesses
	remaining := MaxGuesses - len(guesses)
	if m.phase == phasePlaying {
		remaining-- // one slot used by input row
	}
	for range remaining {
		rows = append(rows, m.renderEmptyRow())
	}

	return rows
}

func (m Model) renderGuessRow(entry GuessEntry) string {
	var pegs []string
	for _, v := range entry.Code {
		pegs = append(pegs, colorStyle(v).Render("●"))
	}
	codeStr := strings.Join(pegs, " ")

	// Feedback pegs
	fbStr := renderFeedback(entry.Feedback)

	return fmt.Sprintf("  %s  │  %s", codeStr, fbStr)
}

func (m Model) renderInputRow() string {
	var pegs []string
	for i := range CodeLength {
		var peg string
		if m.input[i] != 0 {
			peg = colorStyle(m.input[i]).Render("●")
		} else {
			peg = emptyStyle.Render("○")
		}
		if i == m.cursor {
			peg = cursorStyle.Render("[") + peg + cursorStyle.Render("]")
		} else {
			peg = " " + peg + " "
		}
		pegs = append(pegs, peg)
	}
	return strings.Join(pegs, "")
}

func (m Model) renderEmptyRow() string {
	var pegs []string
	for range CodeLength {
		pegs = append(pegs, emptyStyle.Render("○"))
	}
	codeStr := strings.Join(pegs, " ")
	fbStr := strings.Repeat(emptyStyle.Render("·")+" ", CodeLength)
	return fmt.Sprintf("  %s  │  %s", codeStr, strings.TrimRight(fbStr, " "))
}

func renderFeedback(fb Feedback) string {
	var parts []string
	for range fb.Exact {
		parts = append(parts, exactStyle.Render("●"))
	}
	for range fb.Misplaced {
		parts = append(parts, misplacedStyle.Render("○"))
	}
	for range CodeLength - fb.Exact - fb.Misplaced {
		parts = append(parts, emptyStyle.Render("·"))
	}
	return strings.Join(parts, " ")
}

func formatCode(code [CodeLength]int) string {
	var parts []string
	for _, v := range code {
		parts = append(parts, colorStyle(v).Render("●"))
	}
	return strings.Join(parts, " ")
}

func colorStyle(color int) lipgloss.Style {
	switch color {
	case 1:
		return lipgloss.NewStyle().Foreground(lipgloss.Color("#FF0000"))
	case 2:
		return lipgloss.NewStyle().Foreground(lipgloss.Color("#00E632"))
	case 3:
		return lipgloss.NewStyle().Foreground(lipgloss.Color("#0000FF"))
	case 4:
		return lipgloss.NewStyle().Foreground(lipgloss.Color("#FFD700"))
	case 5:
		return lipgloss.NewStyle().Foreground(lipgloss.Color("#840084"))
	case 6:
		return lipgloss.NewStyle().Foreground(lipgloss.Color("#FF8C00"))
	default:
		return lipgloss.NewStyle().Foreground(lipgloss.Color("242"))
	}
}

// --- Styles ---

var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#DCFFDC"))

	legendStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("242"))

	exactStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFD700"))

	misplacedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("15"))

	emptyStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("242"))

	cursorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFD700")).
			Bold(true)

	winStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#00E632"))

	loseStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FF0000"))

	messageStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#DCFFDC"))

	footerStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("240"))
)
