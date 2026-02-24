package splash

import (
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// blinkMsg triggers the prompt blink toggle.
type blinkMsg struct{}

// Model is the splash screen tea.Model.
type Model struct {
	width      int
	height     int
	ready      bool // wait for WindowSizeMsg
	blinkShow  bool // whether prompt is visible
	blinkTicks int  // counter for blink toggle
}

// New creates a splash screen model.
func New() Model {
	return Model{
		blinkShow: true,
	}
}

// Init starts the blink timer.
func (m Model) Init() tea.Cmd {
	return tea.Tick(500*time.Millisecond, func(time.Time) tea.Msg {
		return blinkMsg{}
	})
}

// Update handles messages for the splash screen.
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.ready = true
		return m, nil

	case blinkMsg:
		m.blinkShow = !m.blinkShow
		m.blinkTicks++
		return m, tea.Tick(500*time.Millisecond, func(time.Time) tea.Msg {
			return blinkMsg{}
		})
	}

	return m, nil
}

// View renders the splash screen.
func (m Model) View() string {
	if !m.ready {
		return ""
	}

	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("15")) // bright white

	creditsStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("242")) // dim gray

	promptStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("34")). // green
		Italic(true)

	// Build the content block.
	var b strings.Builder
	b.WriteString(titleStyle.Render(TitleArt))
	b.WriteString("\n\n")
	b.WriteString(creditsStyle.Render(Credits))
	b.WriteString("\n\n\n")
	if m.blinkShow {
		b.WriteString(promptStyle.Render(Prompt))
	}

	content := lipgloss.NewStyle().
		Align(lipgloss.Center).
		Render(b.String())

	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, content)
}
