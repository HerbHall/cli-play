package typingtest

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type tickMsg struct{}

func tickCmd() tea.Cmd {
	return tea.Tick(time.Second, func(time.Time) tea.Msg {
		return tickMsg{}
	})
}

// Model is the Bubbletea model for the Typing Test game.
type Model struct {
	game      *Game
	width     int
	height    int
	done      bool
	ticking   bool
	HighScore int
}

// New creates a fresh Typing Test model.
func New() Model {
	return Model{
		game: NewGame(),
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

// FinalScore returns the WPM as an integer, or 0 if incomplete.
func (m Model) FinalScore() int {
	if m.game == nil || m.game.State != Finished {
		return 0
	}
	return m.game.WPM()
}

// Update handles input and advances game state.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tickMsg:
		if m.game.State == Playing && m.ticking {
			expired := m.game.Tick()
			if expired {
				m.ticking = false
				return m, nil
			}
			return m, tickCmd()
		}
		return m, nil

	case tea.KeyMsg:
		key := msg.String()

		if key == "ctrl+c" {
			return m, tea.Quit
		}

		switch m.game.State {
		case Ready:
			return m.updateReady(key)
		case Playing:
			return m.updatePlaying(key)
		case Finished:
			return m.updateFinished(key)
		}
	}

	return m, nil
}

func (m Model) updateReady(key string) (tea.Model, tea.Cmd) {
	switch {
	case key == "q" || key == "esc":
		m.done = true
		return m, nil
	case len(key) == 1 && key[0] >= 'a' && key[0] <= 'z':
		m.game.TypeChar(rune(key[0]))
		m.ticking = true
		return m, tickCmd()
	case len(key) == 1 && key[0] >= 'A' && key[0] <= 'Z':
		m.game.TypeChar(rune(key[0]))
		m.ticking = true
		return m, tickCmd()
	}
	return m, nil
}

func (m Model) updatePlaying(key string) (tea.Model, tea.Cmd) {
	switch {
	case key == "esc":
		m.done = true
	case key == "backspace":
		m.game.Backspace()
	case key == " ":
		m.game.AdvanceWord()
		if m.game.State == Finished {
			m.ticking = false
		}
	case len(key) == 1:
		ch := rune(key[0])
		if (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') {
			m.game.TypeChar(ch)
		}
	}
	return m, nil
}

func (m Model) updateFinished(key string) (tea.Model, tea.Cmd) {
	switch key {
	case "n":
		m.game = NewGame()
		m.ticking = false
	case "q", "esc":
		m.done = true
	}
	return m, nil
}

// View renders the complete game screen.
func (m Model) View() string {
	var sections []string

	sections = append(sections,
		titleStyle.Render("T Y P I N G  T E S T"),
		"",
	)

	switch m.game.State {
	case Ready:
		sections = append(sections, m.viewReady()...)
	case Playing:
		sections = append(sections, m.viewPlaying()...)
	case Finished:
		sections = append(sections, m.viewFinished()...)
	}

	content := lipgloss.JoinVertical(lipgloss.Center, sections...)
	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, content)
}

func (m Model) viewReady() []string {
	var sections []string

	sections = append(sections,
		m.renderTextBlock(),
		"",
		statsStyle.Render(fmt.Sprintf("Time: %ds  |  Words: %d", m.game.TimeLimit, len(m.game.Words))),
		"",
		footerStyle.Render("Start typing to begin | Q Quit"),
	)
	return sections
}

func (m Model) viewPlaying() []string {
	var sections []string

	sections = append(sections,
		m.renderTextBlock(),
		"",
		m.renderStats(),
		"",
		footerStyle.Render("Type the words | Space next word | Backspace correct | Esc Quit"),
	)
	return sections
}

func (m Model) viewFinished() []string {
	var sections []string

	wpm := m.game.WPM()
	accuracy := m.game.Accuracy()
	wordsCompleted := m.game.CurrentWord

	var resultStyle lipgloss.Style
	switch {
	case wpm > 50:
		resultStyle = goodResultStyle
	case wpm >= 30:
		resultStyle = mediumResultStyle
	default:
		resultStyle = slowResultStyle
	}

	sections = append(sections,
		resultStyle.Render("Test Complete!"),
		"",
		resultStyle.Render(fmt.Sprintf("WPM: %d", wpm)),
		resultStyle.Render(fmt.Sprintf("Accuracy: %.1f%%", accuracy)),
		resultStyle.Render(fmt.Sprintf("Words: %d / %d", wordsCompleted, len(m.game.Words))),
		resultStyle.Render(fmt.Sprintf("Time: %ds", int(m.game.Elapsed.Seconds()))),
	)

	if m.HighScore > 0 && wpm > m.HighScore {
		sections = append(sections, "", goodResultStyle.Render("NEW HIGH SCORE!"))
	} else if m.HighScore > 0 {
		sections = append(sections, "", statsStyle.Render(fmt.Sprintf("Best: %d WPM", m.HighScore)))
	}

	sections = append(sections,
		"",
		footerStyle.Render("N New Test | Q Quit"),
	)
	return sections
}

func (m Model) renderStats() string {
	wpm := m.game.WPM()
	accuracy := m.game.Accuracy()
	remaining := m.game.TimeRemaining()

	return statsStyle.Render(fmt.Sprintf(
		"WPM: %d  |  Accuracy: %.1f%%  |  Time: %ds",
		wpm, accuracy, remaining,
	))
}

// renderTextBlock renders the target text with word-by-word highlighting.
// Uses a wrapping approach: renders words inline, wrapping at ~60 chars.
func (m Model) renderTextBlock() string {
	const maxWidth = 60

	var lines []string
	var currentLine strings.Builder
	lineLen := 0

	for i, word := range m.game.Words {
		rendered := m.renderWord(i)
		wordLen := len(word)

		if lineLen > 0 && lineLen+1+wordLen > maxWidth {
			lines = append(lines, currentLine.String())
			currentLine.Reset()
			lineLen = 0
		}

		if lineLen > 0 {
			currentLine.WriteString(" ")
			lineLen++
		}
		currentLine.WriteString(rendered)
		lineLen += wordLen
	}

	if currentLine.Len() > 0 {
		lines = append(lines, currentLine.String())
	}

	return strings.Join(lines, "\n")
}

func (m Model) renderWord(wordIdx int) string {
	word := m.game.Words[wordIdx]

	switch {
	case wordIdx < m.game.CurrentWord:
		// Completed word: show green for correct, red for errors.
		return m.renderCompletedWord(wordIdx)
	case wordIdx == m.game.CurrentWord:
		// Current word: show typed chars with cursor.
		return m.renderCurrentWord()
	default:
		// Upcoming word: dim.
		return upcomingStyle.Render(word)
	}
}

func (m Model) renderCompletedWord(wordIdx int) string {
	word := m.game.Words[wordIdx]
	var b strings.Builder
	for i, ch := range word {
		if i < len(m.game.WordErrors[wordIdx]) && m.game.WordErrors[wordIdx][i] {
			b.WriteString(errorStyle.Render(string(ch)))
		} else {
			b.WriteString(correctStyle.Render(string(ch)))
		}
	}
	return b.String()
}

func (m Model) renderCurrentWord() string {
	word := m.game.Words[m.game.CurrentWord]
	var b strings.Builder

	for i, ch := range word {
		switch {
		case i < m.game.CharPos:
			// Already typed.
			if m.game.CurrentErrors[i] {
				b.WriteString(errorStyle.Render(string(ch)))
			} else {
				b.WriteString(correctStyle.Render(string(ch)))
			}
		case i == m.game.CharPos:
			// Cursor position.
			b.WriteString(cursorStyle.Render(string(ch)))
		default:
			// Not yet reached.
			b.WriteString(upcomingStyle.Render(string(ch)))
		}
	}

	return b.String()
}

// --- Styles ---

var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#DCFFDC"))

	correctStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#00E632"))

	cursorStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FFD700")).
			Background(lipgloss.Color("#444444"))

	errorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF0000")).
			Strikethrough(true)

	upcomingStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("242"))

	statsStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#DCFFDC"))

	goodResultStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#00E632"))

	mediumResultStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("#FFD700"))

	slowResultStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FF0000"))

	footerStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("240"))
)
