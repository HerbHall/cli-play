package app

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/herbhall/cli-play/internal/menu"
	"github.com/herbhall/cli-play/internal/splash"
	"github.com/herbhall/cli-play/internal/transition"
)

// screen identifies the active screen.
type screen int

const (
	screenSplash screen = iota
	screenTransition
	screenMenu
)

// Model is the top-level container that routes between screens.
type Model struct {
	active     screen
	width      int
	height     int
	splash     splash.Model
	transition transition.Model
	menu       menu.Model
}

// New creates the top-level app model starting at the splash screen.
func New() Model {
	return Model{
		active: screenSplash,
		splash: splash.New(),
		menu:   menu.New(),
	}
}

// Init delegates to the active sub-model's Init.
func (m Model) Init() tea.Cmd {
	return m.splash.Init()
}

// Update handles messages and routes them to the active sub-model.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		// Forward to all sub-models so they have dimensions when activated.
		m.splash, _ = m.splash.Update(msg)
		m.transition, _ = m.transition.Update(msg)
		m.menu, _ = m.menu.Update(msg)
		return m, nil

	case tea.KeyMsg:
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}

		if m.active == screenSplash {
			// Build the text that the transition will dissolve and reveal.
			splashText := splash.TitleArt + "\n\n" + splash.Credits
			menuText := menu.MenuText(m.width, m.height)
			m.transition = transition.New(m.width, m.height, splashText, menuText)
			m.active = screenTransition
			return m, m.transition.Init()
		}
	}

	// Forward to active sub-model.
	switch m.active {
	case screenSplash:
		var cmd tea.Cmd
		m.splash, cmd = m.splash.Update(msg)
		return m, cmd

	case screenTransition:
		var cmd tea.Cmd
		m.transition, cmd = m.transition.Update(msg)
		if m.transition.Done() {
			m.active = screenMenu
			return m, m.menu.Init()
		}
		return m, cmd

	case screenMenu:
		var cmd tea.Cmd
		m.menu, cmd = m.menu.Update(msg)
		if m.menu.Quitting() {
			return m, tea.Quit
		}
		if m.menu.Selected() >= 0 {
			// Games not implemented yet.
			return m, tea.Quit
		}
		return m, cmd
	}

	return m, nil
}

// View renders the active sub-model.
func (m Model) View() string {
	switch m.active {
	case screenSplash:
		return m.splash.View()
	case screenTransition:
		return m.transition.View()
	case screenMenu:
		return m.menu.View()
	}
	return ""
}
