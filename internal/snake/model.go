package snake

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type phase int

const (
	phasePlaying phase = iota
	phasePaused
	phaseGameOver
)

const tickInterval = 150 * time.Millisecond

type tickMsg struct{}

func tickCmd() tea.Cmd {
	return tea.Tick(tickInterval, func(time.Time) tea.Msg {
		return tickMsg{}
	})
}

// Model is the Bubbletea model for the Snake game.
type Model struct {
	game      *Game
	phase     phase
	width     int
	height    int
	done      bool
	HighScore int
}

// New creates a fresh Snake model.
func New() Model {
	return Model{
		game:  NewGame(),
		phase: phasePlaying,
	}
}

// Init starts the tick loop.
func (m Model) Init() tea.Cmd {
	return tickCmd()
}

// Done returns true when the player wants to exit to the menu.
func (m Model) Done() bool {
	return m.done
}

// FinalScore returns the score (food eaten), or 0 if game incomplete.
func (m Model) FinalScore() int {
	if m.game == nil {
		return 0
	}
	return m.game.Score
}

// Update handles input and advances game state.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tickMsg:
		if m.phase == phasePlaying {
			m.game.Tick()
			if m.game.State == StateGameOver {
				m.phase = phaseGameOver
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

		switch m.phase {
		case phasePlaying:
			return m.updatePlaying(key)
		case phasePaused:
			return m.updatePaused(key)
		case phaseGameOver:
			return m.updateGameOver(key)
		}
	}

	return m, nil
}

func (m Model) updatePlaying(key string) (tea.Model, tea.Cmd) {
	switch key {
	case "up", "k":
		m.game.SetDirection(DirUp)
	case "down", "j":
		m.game.SetDirection(DirDown)
	case "left", "h":
		m.game.SetDirection(DirLeft)
	case "right", "l":
		m.game.SetDirection(DirRight)
	case "p":
		m.phase = phasePaused
		return m, nil
	case "q", "esc":
		m.done = true
	}
	return m, nil
}

func (m Model) updatePaused(key string) (tea.Model, tea.Cmd) {
	switch key {
	case "p":
		m.phase = phasePlaying
		return m, tickCmd()
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
		return m, tickCmd()
	case "q", "esc":
		m.done = true
	}
	return m, nil
}

// View renders the complete game screen.
func (m Model) View() string {
	var sections []string

	// Title.
	sections = append(sections, titleStyle.Render("S N A K E"))

	// Score line.
	scoreText := fmt.Sprintf("Score: %d", m.game.Score)
	if m.HighScore > 0 {
		scoreText += fmt.Sprintf("  (Best: %d)", m.HighScore)
	}
	sections = append(sections, scoreStyle.Render(scoreText), "", m.renderBoard(), "")

	// Status message.
	switch m.phase {
	case phasePaused:
		sections = append(sections, scoreStyle.Render("PAUSED"))
	case phaseGameOver:
		if m.HighScore > 0 && m.game.Score > m.HighScore {
			sections = append(sections, winStyle.Render(fmt.Sprintf("Game Over! Score: %d -- NEW HIGH SCORE!", m.game.Score)))
		} else {
			sections = append(sections, gameOverStyle.Render(fmt.Sprintf("Game Over! Score: %d", m.game.Score)))
		}
	default:
		sections = append(sections, "")
	}

	// Footer.
	var footer string
	switch m.phase {
	case phasePlaying:
		footer = "Arrow/HJKL Move | P Pause | Q Quit"
	case phasePaused:
		footer = "P Resume | Q Quit"
	case phaseGameOver:
		footer = "N New Game | Q Quit"
	}
	sections = append(sections, footerStyle.Render(footer))

	content := lipgloss.JoinVertical(lipgloss.Center, sections...)
	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, content)
}

func (m Model) renderBoard() string {
	// Top border.
	topBorder := borderStyle.Render("┌" + strings.Repeat("──", m.game.Width) + "┐")

	rows := make([]string, 0, m.game.Height+2)
	rows = append(rows, topBorder)

	headPos := m.game.Snake[0]

	for y := 0; y < m.game.Height; y++ {
		var row strings.Builder
		row.WriteString(borderStyle.Render("│"))
		for x := 0; x < m.game.Width; x++ {
			p := Point{X: x, Y: y}
			switch {
			case p == headPos:
				row.WriteString(headStyle.Render("██"))
			case m.game.IsOccupied(p):
				row.WriteString(bodyStyle.Render("██"))
			case p == m.game.Food:
				row.WriteString(foodStyle.Render("●·"))
			default:
				row.WriteString("  ")
			}
		}
		row.WriteString(borderStyle.Render("│"))
		rows = append(rows, row.String())
	}

	// Bottom border.
	bottomBorder := borderStyle.Render("└" + strings.Repeat("──", m.game.Width) + "┘")
	rows = append(rows, bottomBorder)

	return strings.Join(rows, "\n")
}

// --- Styles ---

var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#DCFFDC"))

	scoreStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#DCFFDC"))

	bodyStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#00E632"))

	headStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FFD700"))

	foodStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF0000"))

	borderStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("242"))

	winStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#00E632"))

	gameOverStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FF0000"))

	footerStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("240"))
)
