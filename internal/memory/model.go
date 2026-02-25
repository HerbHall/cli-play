package memory

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type phase int

const (
	phasePlaying  phase = iota
	phaseRevealed       // two non-matching cards shown briefly
	phaseGameOver
)

type flipBackMsg struct{}

func flipBackCmd() tea.Cmd {
	return tea.Tick(800*time.Millisecond, func(time.Time) tea.Msg {
		return flipBackMsg{}
	})
}

// Model is the Bubbletea model for the Memory game.
type Model struct {
	game      *Game
	cursorRow int
	cursorCol int
	width     int
	height    int
	done      bool
	phase     phase
	HighScore int
}

// New creates a fresh Memory model.
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

// Done returns true when the player wants to exit to the menu.
func (m Model) Done() bool {
	return m.done
}

// FinalScore returns the number of moves taken, or -1 if the game is not complete.
func (m Model) FinalScore() int {
	if m.game == nil || !m.game.GameOver {
		return -1
	}
	return m.game.Moves
}

// Update handles input and advances game state.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case flipBackMsg:
		if m.phase == phaseRevealed {
			m.game.ResolveNoMatch()
			m.phase = phasePlaying
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
		case phaseRevealed:
			return m.updateRevealed(key)
		case phaseGameOver:
			return m.updateGameOver(key)
		}
	}

	return m, nil
}

func (m Model) updatePlaying(key string) (tea.Model, tea.Cmd) {
	switch key {
	case "up", "k":
		if m.cursorRow > 0 {
			m.cursorRow--
		}
	case "down", "j":
		if m.cursorRow < rows-1 {
			m.cursorRow++
		}
	case "left", "h":
		if m.cursorCol > 0 {
			m.cursorCol--
		}
	case "right", "l":
		if m.cursorCol < cols-1 {
			m.cursorCol++
		}
	case "enter", " ":
		if m.game.FlipCard(m.cursorRow, m.cursorCol) && m.game.HasSecond {
			if m.game.CheckMatch() {
				m.game.ResolveMatch()
				if m.game.GameOver {
					m.phase = phaseGameOver
				}
			} else {
				m.phase = phaseRevealed
				return m, flipBackCmd()
			}
		}
	case "q", "esc":
		m.done = true
	}
	return m, nil
}

func (m Model) updateRevealed(key string) (tea.Model, tea.Cmd) {
	switch key {
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
		m.cursorRow = 0
		m.cursorCol = 0
	case "q", "esc":
		m.done = true
	}
	return m, nil
}

// View renders the complete game screen.
func (m Model) View() string {
	var sections []string

	sections = append(sections,
		titleStyle.Render("M E M O R Y"),
		"",
		m.renderStatus(),
		"",
		m.renderGrid(),
		"",
	)

	if m.phase == phaseGameOver {
		sections = append(sections, m.renderWinMessage(), "")
	}

	sections = append(sections, m.renderFooter())

	content := lipgloss.JoinVertical(lipgloss.Center, sections...)
	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, content)
}

func (m Model) renderStatus() string {
	return statusStyle.Render(fmt.Sprintf("Pairs: %d/%d  |  Moves: %d",
		m.game.PairsFound, numPairs, m.game.Moves))
}

func (m Model) renderGrid() string {
	gridRows := make([]string, 0, rows)

	for r := 0; r < rows; r++ {
		cells := make([]string, 0, cols)
		for c := 0; c < cols; c++ {
			card := m.game.Board[r][c]
			isCursor := r == m.cursorRow && c == m.cursorCol
			cells = append(cells, m.renderCard(card, isCursor))
		}
		gridRows = append(gridRows, strings.Join(cells, " "))
	}

	return strings.Join(gridRows, "\n")
}

func (m Model) renderCard(card Card, isCursor bool) string {
	var text string
	var style lipgloss.Style

	base := cardBaseStyle

	switch card.State {
	case FaceDown:
		text = " ? "
		if isCursor {
			style = base.Background(cursorBg).Foreground(faceDownColor)
		} else {
			style = base.Foreground(faceDownColor)
		}
	case FaceUp:
		text = fmt.Sprintf(" %c ", card.Symbol)
		if isCursor {
			style = base.Background(cursorBg).Foreground(faceUpColor).Bold(true)
		} else {
			style = base.Foreground(faceUpColor).Bold(true)
		}
	case Matched:
		text = fmt.Sprintf(" %c ", card.Symbol)
		if isCursor {
			style = base.Background(cursorBg).Foreground(matchedColor).Bold(true)
		} else {
			style = base.Foreground(matchedColor).Bold(true)
		}
	}

	return style.Render(text)
}

func (m Model) renderWinMessage() string {
	msg := fmt.Sprintf("All pairs found in %d moves!", m.game.Moves)
	switch {
	case m.HighScore > 0 && m.game.Moves < m.HighScore:
		msg = fmt.Sprintf("All pairs found in %d moves -- NEW BEST!", m.game.Moves)
	case m.HighScore > 0:
		msg = fmt.Sprintf("All pairs found in %d moves (Best: %d)", m.game.Moves, m.HighScore)
	}
	return winStyle.Render(msg)
}

func (m Model) renderFooter() string {
	switch m.phase {
	case phasePlaying:
		return footerStyle.Render("Arrows Move | Enter Flip | Q Quit")
	case phaseRevealed:
		return footerStyle.Render("No match -- flipping back...")
	case phaseGameOver:
		return footerStyle.Render("N New Game | Q Quit")
	}
	return ""
}

// --- Styles ---

var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#DCFFDC"))

	statusStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("242"))

	cardBaseStyle = lipgloss.NewStyle().
			Width(3).
			Align(lipgloss.Center)

	faceDownColor = lipgloss.Color("242")
	faceUpColor   = lipgloss.Color("#FFD700")
	matchedColor  = lipgloss.Color("#00E632")
	cursorBg      = lipgloss.Color("#333333")

	winStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#00E632"))

	footerStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("240"))
)
