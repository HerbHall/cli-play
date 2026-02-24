package menu

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
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
}

// Model is the game selection menu.
type Model struct {
	choices  []GameChoice
	cursor   int
	width    int
	height   int
	selected int
	quitting bool
}

// New creates a menu model.
func New() Model {
	return Model{
		choices:  Games,
		cursor:   0,
		selected: -1,
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
)

// View renders the menu.
func (m Model) View() string {
	var b strings.Builder

	b.WriteString(titleStyle.Render("CLI Play"))
	b.WriteString("\n\n")

	for i, choice := range m.choices {
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

	for _, choice := range Games {
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

// Quitting returns true if the user pressed quit.
func (m Model) Quitting() bool {
	return m.quitting
}
