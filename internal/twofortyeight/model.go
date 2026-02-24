package twofortyeight

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type phase int

const (
	phasePlaying  phase = iota
	phaseWon                    // reached 2048, offer continue or new game
	phaseGameOver               // no moves left
)

// Model is the Bubbletea model for the 2048 game.
type Model struct {
	game      *Game
	phase     phase
	width     int
	height    int
	done      bool
	HighScore int
}

// New creates a fresh 2048 model.
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
		case phaseWon:
			return m.updateWon(key)
		case phaseGameOver:
			return m.updateGameOver(key)
		}
	}

	return m, nil
}

func (m Model) updatePlaying(key string) (tea.Model, tea.Cmd) {
	switch key {
	case "up", "k":
		m.game.Move(Up)
	case "down", "j":
		m.game.Move(Down)
	case "left", "h":
		m.game.Move(Left)
	case "right", "l":
		m.game.Move(Right)
	case "n":
		m.game.Reset()
		m.phase = phasePlaying
	case "q", "esc":
		m.done = true
	}

	if m.game.Won && !m.game.Continued {
		m.phase = phaseWon
	} else if m.game.Over {
		m.phase = phaseGameOver
	}

	return m, nil
}

func (m Model) updateWon(key string) (tea.Model, tea.Cmd) {
	switch key {
	case "c":
		m.game.ContinueAfterWin()
		m.phase = phasePlaying
	case "n":
		m.game.Reset()
		m.phase = phasePlaying
	case "q", "esc":
		m.done = true
	}
	return m, nil
}

func (m Model) updateGameOver(key string) (tea.Model, tea.Cmd) {
	switch key {
	case "n", "enter":
		m.game.Reset()
		m.phase = phasePlaying
	case "q", "esc":
		m.done = true
	}
	return m, nil
}

// Done returns true when the player wants to return to the menu.
func (m Model) Done() bool {
	return m.done
}

// FinalScore returns the game score.
func (m Model) FinalScore() int {
	if m.game == nil {
		return 0
	}
	return m.game.Score
}

// View renders the complete game screen.
func (m Model) View() string {
	var sections []string

	// Title
	sections = append(sections, titleStyle.Render("2 0 4 8"))

	// Score
	scoreText := fmt.Sprintf("Score: %d", m.game.Score)
	if m.HighScore > 0 {
		scoreText += fmt.Sprintf("  (Best: %d)", m.HighScore)
	}
	sections = append(sections, scoreStyle.Render(scoreText))

	sections = append(sections, "")

	// Board
	sections = append(sections, m.renderBoard())

	sections = append(sections, "")

	// Phase message
	switch m.phase {
	case phaseWon:
		sections = append(sections, wonStyle.Render("You reached 2048!"))
	case phaseGameOver:
		if m.HighScore > 0 && m.game.Score > m.HighScore {
			sections = append(sections, wonStyle.Render(fmt.Sprintf("Game Over! Score: %d — NEW HIGH SCORE!", m.game.Score)))
		} else {
			sections = append(sections, gameOverStyle.Render(fmt.Sprintf("Game Over! Score: %d", m.game.Score)))
		}
	default:
		sections = append(sections, "")
	}

	// Help
	var footer string
	switch m.phase {
	case phasePlaying:
		footer = "Arrow/HJKL Move | N New | Q Quit"
	case phaseWon:
		footer = "C Continue | N New Game | Q Quit"
	case phaseGameOver:
		footer = "N New Game | Q Quit"
	}
	sections = append(sections, helpStyle.Render(footer))

	content := lipgloss.JoinVertical(lipgloss.Center, sections...)

	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, content)
}

func (m Model) renderBoard() string {
	var rows []string
	for r := range boardSize {
		var cells []string
		for c := range boardSize {
			cells = append(cells, renderTile(m.game.Board[r][c]))
		}
		rows = append(rows, lipgloss.JoinHorizontal(lipgloss.Center, cells...))
	}
	board := lipgloss.JoinVertical(lipgloss.Left, rows...)
	return boardStyle.Render(board)
}

const tileWidth = 7

func renderTile(value int) string {
	style := tileStyle(value)
	if value == 0 {
		return style.Render("  .  ")
	}
	label := fmt.Sprintf("%d", value)
	pad := tileWidth - len(label)
	left := pad / 2
	right := pad - left
	text := strings.Repeat(" ", left) + label + strings.Repeat(" ", right)
	return style.Render(text)
}

func tileStyle(value int) lipgloss.Style {
	base := lipgloss.NewStyle().
		Width(tileWidth).
		Align(lipgloss.Center)

	switch value {
	case 0:
		return base.Foreground(lipgloss.Color("240"))
	case 2:
		return base.Foreground(lipgloss.Color("#EEE4DA")).Bold(true)
	case 4:
		return base.Foreground(lipgloss.Color("#EDE0C8")).Bold(true)
	case 8:
		return base.Foreground(lipgloss.Color("#F2B179")).Bold(true)
	case 16:
		return base.Foreground(lipgloss.Color("#F59563")).Bold(true)
	case 32:
		return base.Foreground(lipgloss.Color("#F67C5F")).Bold(true)
	case 64:
		return base.Foreground(lipgloss.Color("#F65E3B")).Bold(true)
	case 128:
		return base.Foreground(lipgloss.Color("#EDCF72")).Bold(true)
	case 256:
		return base.Foreground(lipgloss.Color("#EDCC61")).Bold(true)
	case 512:
		return base.Foreground(lipgloss.Color("#EDC850")).Bold(true)
	case 1024:
		return base.Foreground(lipgloss.Color("#EDC53F")).Bold(true)
	case 2048:
		return base.Foreground(lipgloss.Color("#EDC22E")).Bold(true)
	default:
		return base.Foreground(lipgloss.Color("15")).
			Background(lipgloss.Color("#3C3A32")).Bold(true)
	}
}

// Styles.
var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#EDC22E"))

	scoreStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("242"))

	boardStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("240")).
			Padding(0, 1)

	wonStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#EDC22E"))

	gameOverStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FF4444"))

	helpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("240"))
)
