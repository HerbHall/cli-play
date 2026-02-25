package settings

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// option represents a single setting with its possible values.
type option struct {
	label   string
	values  []string
	current int
}

// Model is the settings screen tea.Model.
type Model struct {
	store   *Store
	options []option
	cursor  int
	width   int
	height  int
	done    bool
	saved   bool
}

// NewModel creates a settings screen from the current config.
func NewModel(store *Store) Model {
	cfg := store.Config

	opts := []option{
		{
			label:   "Animation Speed",
			values:  []string{"slow", "normal", "fast", "off"},
			current: indexOf(string(cfg.AnimationSpeed), []string{"slow", "normal", "fast", "off"}),
		},
		{
			label:   "Color Theme",
			values:  []string{"matrix", "amber", "blue", "red"},
			current: indexOf(string(cfg.Theme), []string{"matrix", "amber", "blue", "red"}),
		},
		{
			label:   "Minesweeper Default",
			values:  []string{"beginner", "intermediate", "expert"},
			current: indexOf(cfg.MinesweeperDefault, []string{"beginner", "intermediate", "expert"}),
		},
		{
			label:   "Sudoku Default",
			values:  []string{"easy", "medium", "hard"},
			current: indexOf(cfg.SudokuDefault, []string{"easy", "medium", "hard"}),
		},
	}

	return Model{
		store:   store,
		options: opts,
	}
}

func indexOf(val string, list []string) int {
	for i, v := range list {
		if v == val {
			return i
		}
	}
	return 0
}

// Init returns nil; no initial command needed.
func (m Model) Init() tea.Cmd {
	return nil
}

// Update handles input for the settings screen.
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
				m.cursor = len(m.options) - 1
			}
		case "down", "j":
			m.cursor++
			if m.cursor >= len(m.options) {
				m.cursor = 0
			}
		case "left", "h":
			opt := &m.options[m.cursor]
			opt.current--
			if opt.current < 0 {
				opt.current = len(opt.values) - 1
			}
			m.applyToStore()
		case "right", "l":
			opt := &m.options[m.cursor]
			opt.current++
			if opt.current >= len(opt.values) {
				opt.current = 0
			}
			m.applyToStore()
		case "enter", "s":
			_ = m.store.Save()
			m.saved = true
			m.done = true
		case "q", "esc":
			m.done = true
		}
	}

	return m, nil
}

// applyToStore writes the current option selections back to the store config.
func (m *Model) applyToStore() {
	m.store.Config.AnimationSpeed = AnimationSpeed(m.options[0].values[m.options[0].current])
	m.store.Config.Theme = Theme(m.options[1].values[m.options[1].current])
	m.store.Config.MinesweeperDefault = m.options[2].values[m.options[2].current]
	m.store.Config.SudokuDefault = m.options[3].values[m.options[3].current]
}

// Styles.
var (
	settingsTitleStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("#DCFFDC"))

	settingsLabelStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("15")).
				Width(22)

	settingsActiveLabel = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("#FFD700")).
				Width(22)

	settingsValueStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("242"))

	settingsSelectedValue = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("#FFD700"))

	settingsCursorStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#FFD700"))

	settingsFooterStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("240"))
)

// View renders the settings screen.
func (m Model) View() string {
	var b strings.Builder

	b.WriteString(settingsTitleStyle.Render("Settings"))
	b.WriteString("\n\n")

	for i, opt := range m.options {
		indicator := "  "
		labelStyle := settingsLabelStyle
		if i == m.cursor {
			indicator = "> "
			labelStyle = settingsActiveLabel
		}

		b.WriteString(settingsCursorStyle.Render(indicator))
		b.WriteString(labelStyle.Render(opt.label))
		b.WriteString("  ")

		// Render value selector: < value1  value2  value3 >
		var vals []string
		for j, v := range opt.values {
			if j == opt.current {
				vals = append(vals, settingsSelectedValue.Render(fmt.Sprintf("[%s]", v)))
			} else {
				vals = append(vals, settingsValueStyle.Render(v))
			}
		}
		b.WriteString(strings.Join(vals, "  "))
		b.WriteString("\n")
	}

	b.WriteString("\n")
	b.WriteString(settingsFooterStyle.Render("  ↑↓ Navigate | ←→ Change | S Save | Q Back"))

	content := lipgloss.NewStyle().
		Align(lipgloss.Center).
		Render(b.String())

	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, content)
}

// Done returns true when the user exits the settings screen.
func (m Model) Done() bool {
	return m.done
}

// Saved returns true if settings were saved before exiting.
func (m Model) Saved() bool {
	return m.saved
}
