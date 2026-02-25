package menu

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/herbhall/cli-play/internal/scores"
)

// GameChoice represents a selectable game entry.
type GameChoice struct {
	Name        string
	Description string
}

// Games is the list of available games.
var Games = []GameChoice{
	{"Yahtzee", "Roll dice, fill your scorecard"},
	{"Blackjack", "Beat the dealer to 21"},
	{"Wordle", "Guess the 5-letter word"},
	{"Minesweeper", "Clear the field, dodge the mines"},
	{"Sudoku", "Fill the grid with logic"},
	{"2048", "Slide and merge to 2048"},
	{"Hangman", "Guess the word letter by letter"},
	{"Tic-Tac-Toe", "Beat the AI on a 3x3 grid"},
	{"Mastermind", "Break the secret color code"},
}

// SettingsIndex is the menu index for the Settings entry.
const SettingsIndex = 9

// Model is the game selection menu.
type Model struct {
	choices  []GameChoice
	cursor   int
	width    int
	height   int
	selected int
	quitting bool
	scores   *scores.Store
}

// allChoices returns Games plus the Settings entry.
var allChoices = append(Games, GameChoice{"Settings", "Preferences and configuration"})

// New creates a menu model with optional score display.
func New(s *scores.Store) Model {
	return Model{
		choices:  allChoices,
		cursor:   0,
		selected: -1,
		scores:   s,
	}
}

// Init returns nil; no initial command needed.
func (m Model) Init() tea.Cmd {
	return nil
}

// Update handles key navigation.
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			m.cursor--
			if m.cursor < 0 {
				m.cursor = len(m.choices) - 1
			}
		case "down", "j":
			m.cursor++
			if m.cursor >= len(m.choices) {
				m.cursor = 0
			}
		case "enter":
			m.selected = m.cursor
		case "q", "esc":
			m.quitting = true
		}
	}

	return m, nil
}

// Styles.
var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#DCFFDC"))

	cursorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFD700"))

	nameStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("15"))

	nameSelectedStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("#FFD700"))

	descStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("242"))

	footerStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("240"))

	highScoreStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFD700"))
)

// View renders the menu.
func (m Model) View() string {
	var b strings.Builder

	b.WriteString(titleStyle.Render("CLI Play"))
	b.WriteString("\n\n")

	for i, choice := range m.choices {
		// Separator before Settings.
		if i == SettingsIndex {
			b.WriteString("\n")
		}

		indicator := "  "
		ns := nameStyle
		if i == m.cursor {
			indicator = "> "
			ns = nameSelectedStyle
		}

		b.WriteString(cursorStyle.Render(indicator))
		b.WriteString(ns.Render(choice.Name))
		b.WriteString("  ")
		b.WriteString(descStyle.Render(choice.Description))

		if hs := m.highScoreLabel(i); hs != "" {
			b.WriteString("  ")
			b.WriteString(highScoreStyle.Render(hs))
		}
		b.WriteString("\n")
	}

	b.WriteString("\n")
	b.WriteString(footerStyle.Render("  ↑↓ Navigate | Enter Select | Q Quit"))

	content := lipgloss.NewStyle().
		Align(lipgloss.Center).
		Render(b.String())

	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, content)
}

// MenuText returns the menu layout as plain text (no ANSI styling).
// The transition model uses this to pre-compute character positions
// for the reveal animation.
func MenuText(width, height int) string {
	var b strings.Builder

	b.WriteString("CLI Play")
	b.WriteString("\n\n")

	for i, choice := range allChoices {
		if i == SettingsIndex {
			b.WriteString("\n")
		}
		b.WriteString(fmt.Sprintf("  %s  %s\n", choice.Name, choice.Description))
	}

	b.WriteString("\n")
	b.WriteString("  ↑↓ Navigate | Enter Select | Q Quit")

	return b.String()
}

// Selected returns the index of the selected game, or -1 if none.
func (m Model) Selected() int {
	return m.selected
}

// ResetSelection clears the selected state so the menu can be reused
// after returning from a game.
func (m *Model) ResetSelection() {
	m.selected = -1
}

// highScoreLabel returns a formatted high score string for the given game index.
func (m Model) highScoreLabel(index int) string {
	if m.scores == nil {
		return ""
	}
	switch index {
	case 0: // Yahtzee
		if e := m.scores.Get("yahtzee"); e != nil {
			return fmt.Sprintf("[Best: %d]", e.Value)
		}
	case 1: // Blackjack
		if e := m.scores.Get("blackjack"); e != nil {
			return fmt.Sprintf("[Best: $%d]", e.Value)
		}
	case 2: // Wordle
		if e := m.scores.Get("wordle"); e != nil {
			return fmt.Sprintf("[Best: %d/6]", e.Value)
		}
	case 3: // Minesweeper
		return m.bestTimeLabel("minesweeper")
	case 4: // Sudoku
		return m.bestTimeLabel("sudoku")
	case 5: // 2048
		if e := m.scores.Get("2048"); e != nil {
			return fmt.Sprintf("[Best: %d]", e.Value)
		}
	case 6: // Hangman
		if e := m.scores.Get("hangman"); e != nil {
			return fmt.Sprintf("[Best: %d wrong]", e.Value)
		}
	case 7: // Tic-Tac-Toe
		if e := m.scores.Get("tictactoe"); e != nil {
			return fmt.Sprintf("[Wins: %d]", e.Value)
		}
	case 8: // Mastermind
		if e := m.scores.Get("mastermind"); e != nil {
			return fmt.Sprintf("[Best: %d guesses]", e.Value)
		}
	}
	return ""
}

// bestTimeLabel returns the best time across all difficulties for a timed game.
func (m Model) bestTimeLabel(game string) string {
	var best *scores.Entry
	for _, diff := range []string{"beginner", "intermediate", "expert", "easy", "medium", "hard"} {
		e := m.scores.GetDifficulty(game, diff)
		if e != nil && (best == nil || e.Value < best.Value) {
			best = e
		}
	}
	if best == nil {
		return ""
	}
	mins := best.Value / 60
	secs := best.Value % 60
	return fmt.Sprintf("[Best: %d:%02d]", mins, secs)
}

// Quitting returns true if the user pressed quit.
func (m Model) Quitting() bool {
	return m.quitting
}
