package tetris

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

type tickMsg struct{}

func tickCmd(interval int) tea.Cmd {
	return tea.Tick(time.Duration(interval)*time.Millisecond, func(time.Time) tea.Msg {
		return tickMsg{}
	})
}

// Model is the Bubbletea model for the Tetris game.
type Model struct {
	game      *Game
	phase     phase
	width     int
	height    int
	done      bool
	HighScore int
}

// New creates a fresh Tetris model.
func New() Model {
	return Model{
		game:  NewGame(),
		phase: phasePlaying,
	}
}

// Init starts the gravity tick.
func (m Model) Init() tea.Cmd {
	return tickCmd(m.game.TickInterval())
}

// Done returns true when the player wants to return to the menu.
func (m Model) Done() bool {
	return m.done
}

// FinalScore returns the game score, or 0 if the game is still in progress.
func (m Model) FinalScore() int {
	if m.game == nil || !m.game.Over {
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
		if m.phase == phasePlaying && !m.game.Over {
			m.game.MoveDown()
			if m.game.Over {
				m.phase = phaseGameOver
				return m, nil
			}
			return m, tickCmd(m.game.TickInterval())
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
	case "left", "h":
		m.game.MoveLeft()
	case "right", "l":
		m.game.MoveRight()
	case "down", "j":
		m.game.MoveDown()
		if m.game.Over {
			m.phase = phaseGameOver
			return m, nil
		}
	case "up", "k":
		m.game.Rotate()
	case " ":
		m.game.HardDrop()
		if m.game.Over {
			m.phase = phaseGameOver
			return m, nil
		}
		return m, tickCmd(m.game.TickInterval())
	case "p":
		m.phase = phasePaused
		return m, nil
	case "n":
		m.game.Reset()
		m.phase = phasePlaying
		return m, tickCmd(m.game.TickInterval())
	case "q", "esc":
		m.done = true
	}
	return m, nil
}

func (m Model) updatePaused(key string) (tea.Model, tea.Cmd) {
	switch key {
	case "p":
		m.phase = phasePlaying
		return m, tickCmd(m.game.TickInterval())
	case "q", "esc":
		m.done = true
	}
	return m, nil
}

func (m Model) updateGameOver(key string) (tea.Model, tea.Cmd) {
	switch key {
	case "n":
		m.game.Reset()
		m.phase = phasePlaying
		return m, tickCmd(m.game.TickInterval())
	case "q", "esc":
		m.done = true
	}
	return m, nil
}

// View renders the complete game screen.
func (m Model) View() string {
	var sections []string

	sections = append(sections, titleStyle.Render("T E T R I S"))

	// Score, level, lines
	scoreText := fmt.Sprintf("Score: %d", m.game.Score)
	if m.HighScore > 0 {
		scoreText += fmt.Sprintf("  (Best: %d)", m.HighScore)
	}
	sections = append(sections,
		infoStyle.Render(scoreText),
		infoStyle.Render(fmt.Sprintf("Level: %d  Lines: %d", m.game.Level, m.game.Lines)),
		"",
	)

	// Board and next piece side by side
	boardView := m.renderBoard()
	nextView := m.renderNextPiece()
	sideBySide := lipgloss.JoinHorizontal(lipgloss.Top, boardView, "  ", nextView)
	sections = append(sections, sideBySide, "")

	// Phase messages
	switch m.phase {
	case phasePaused:
		sections = append(sections, pauseStyle.Render("PAUSED"), "")
	case phaseGameOver:
		if m.HighScore > 0 && m.game.Score > m.HighScore {
			sections = append(sections, gameOverStyle.Render(
				fmt.Sprintf("GAME OVER! Score: %d -- NEW HIGH SCORE!", m.game.Score),
			), "")
		} else {
			sections = append(sections, gameOverStyle.Render(
				fmt.Sprintf("GAME OVER! Score: %d", m.game.Score),
			), "")
		}
	default:
		sections = append(sections, "")
	}

	// Footer
	var footer string
	switch m.phase {
	case phasePlaying:
		footer = "Arrow/HJKL Move | Space Drop | P Pause | N New | Q Quit"
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
	// Build a composite view: board cells + current piece + ghost piece.
	ghostRow := m.game.GhostRow()
	ghostPiece := m.game.Current
	ghostPiece.Row = ghostRow
	ghostCells := ghostPiece.Cells()
	currentCells := m.game.Current.Cells()

	// Maps for fast lookup.
	type cellInfo struct {
		pieceType PieceType
		isGhost   bool
	}
	overlay := make(map[Point]cellInfo, 8)
	if !m.game.Over {
		for _, c := range ghostCells {
			if c.Row >= 0 && c.Row < BoardHeight && c.Col >= 0 && c.Col < BoardWidth {
				overlay[c] = cellInfo{m.game.Current.Type, true}
			}
		}
		// Current piece overwrites ghost where they overlap.
		for _, c := range currentCells {
			if c.Row >= 0 && c.Row < BoardHeight && c.Col >= 0 && c.Col < BoardWidth {
				overlay[c] = cellInfo{m.game.Current.Type, false}
			}
		}
	}

	var rows strings.Builder
	topBorder := borderStyle.Render("+" + strings.Repeat("--", BoardWidth) + "+")
	rows.WriteString(topBorder)
	rows.WriteString("\n")

	for r := range BoardHeight {
		rows.WriteString(borderStyle.Render("|"))
		for c := range BoardWidth {
			pt := Point{r, c}
			if info, ok := overlay[pt]; ok {
				if info.isGhost {
					rows.WriteString(ghostStyle.Render("[]"))
				} else {
					rows.WriteString(pieceStyle(info.pieceType).Render("[]"))
				}
			} else if m.game.Board[r][c] != PieceNone {
				rows.WriteString(pieceStyle(m.game.Board[r][c]).Render("[]"))
			} else {
				rows.WriteString(emptyStyle.Render(" ."))
			}
		}
		rows.WriteString(borderStyle.Render("|"))
		rows.WriteString("\n")
	}

	bottomBorder := borderStyle.Render("+" + strings.Repeat("--", BoardWidth) + "+")
	rows.WriteString(bottomBorder)

	return rows.String()
}

func (m Model) renderNextPiece() string {
	var b strings.Builder
	b.WriteString(infoStyle.Render("Next:"))
	b.WriteString("\n")

	// Create a small 4x4 preview grid.
	preview := Piece{
		Type:     m.game.Next,
		Rotation: 0,
		Row:      0,
		Col:      0,
	}
	previewCells := preview.Cells()

	cellSet := make(map[Point]bool, 4)
	for _, c := range previewCells {
		cellSet[c] = true
	}

	for r := range 4 {
		for c := range 4 {
			if cellSet[Point{r, c}] {
				b.WriteString(pieceStyle(m.game.Next).Render("[]"))
			} else {
				b.WriteString("  ")
			}
		}
		b.WriteString("\n")
	}

	return b.String()
}

func pieceStyle(pt PieceType) lipgloss.Style {
	base := lipgloss.NewStyle()
	switch pt {
	case PieceI:
		return base.Foreground(lipgloss.Color("#00FFFF"))
	case PieceO:
		return base.Foreground(lipgloss.Color("#FFD700"))
	case PieceT:
		return base.Foreground(lipgloss.Color("#840084"))
	case PieceS:
		return base.Foreground(lipgloss.Color("#00E632"))
	case PieceZ:
		return base.Foreground(lipgloss.Color("#FF0000"))
	case PieceJ:
		return base.Foreground(lipgloss.Color("#0000FF"))
	case PieceL:
		return base.Foreground(lipgloss.Color("#FF8C00"))
	default:
		return base.Foreground(lipgloss.Color("240"))
	}
}

// Styles.
var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#DCFFDC"))

	infoStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#DCFFDC"))

	borderStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("242"))

	emptyStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("238"))

	ghostStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("240"))

	pauseStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FFD700"))

	gameOverStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FF0000"))

	footerStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("240"))
)
