package app

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/herbhall/cli-play/internal/blackjack"
	"github.com/herbhall/cli-play/internal/menu"
	"github.com/herbhall/cli-play/internal/minesweeper"
	"github.com/herbhall/cli-play/internal/splash"
	"github.com/herbhall/cli-play/internal/sudoku"
	"github.com/herbhall/cli-play/internal/transition"
	"github.com/herbhall/cli-play/internal/twofortyeight"
	"github.com/herbhall/cli-play/internal/wordle"
	"github.com/herbhall/cli-play/internal/yahtzee"
)

// gameModel is implemented by every playable game.
type gameModel interface {
	tea.Model
	Done() bool
}

// screen identifies the active screen.
type screen int

const (
	screenSplash screen = iota
	screenTransition
	screenMenu
	screenGame
)

// Model is the top-level container that routes between screens.
type Model struct {
	active     screen
	width      int
	height     int
	splash     splash.Model
	transition transition.Model
	menu       menu.Model
	game       gameModel
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
		if m.game != nil {
			var updated tea.Model
			updated, _ = m.game.Update(msg)
			m.game = updated.(gameModel)
		}
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
		if sel := m.menu.Selected(); sel >= 0 {
			return m.launchGame(sel)
		}
		return m, cmd

	case screenGame:
		var cmd tea.Cmd
		var updated tea.Model
		updated, cmd = m.game.Update(msg)
		m.game = updated.(gameModel)
		if m.game.Done() {
			m.game = nil
			m.active = screenMenu
			m.menu.ResetSelection()
			return m, nil
		}
		return m, cmd
	}

	return m, nil
}

// launchGame creates the appropriate game model for the given menu index.
func (m Model) launchGame(index int) (tea.Model, tea.Cmd) {
	switch index {
	case 0: // Yahtzee
		g := yahtzee.New()
		m.game = &g
	case 1: // Blackjack
		g := blackjack.New()
		m.game = &g
	case 2: // Wordle
		g := wordle.New()
		m.game = &g
	case 3: // Minesweeper
		g := minesweeper.New()
		m.game = &g
	case 4: // Sudoku
		g := sudoku.New()
		m.game = &g
	case 5: // 2048
		g := twofortyeight.New()
		m.game = &g
	default:
		m.menu.ResetSelection()
		return m, nil
	}
	m.active = screenGame
	cmd := m.game.Init()
	// Forward current dimensions to the game.
	sizeMsg := tea.WindowSizeMsg{Width: m.width, Height: m.height}
	var updated tea.Model
	updated, _ = m.game.Update(sizeMsg)
	m.game = updated.(gameModel)
	return m, cmd
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
	case screenGame:
		if m.game != nil {
			return m.game.View()
		}
	}
	return ""
}
