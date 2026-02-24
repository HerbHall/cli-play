package sudoku

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const (
	phaseDifficulty = "difficulty"
	phasePlaying    = "playing"
	phaseGameOver   = "gameover"
)

// Model is the Bubbletea model for the Sudoku game.
type Model struct {
	game       *Game
	cursorRow  int
	cursorCol  int
	pencilMode bool
	width      int
	height     int
	done       bool
	phase      string
	message    string
}

// New creates a fresh Sudoku model starting at difficulty selection.
func New() Model {
	return Model{
		phase: phaseDifficulty,
	}
}

// Init returns nil; no initial command needed.
func (m Model) Init() tea.Cmd {
	return nil
}

// Done returns true when the player wants to return to the menu.
func (m Model) Done() bool {
	return m.done
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
		case phaseDifficulty:
			return m.updateDifficulty(key)
		case phasePlaying:
			return m.updatePlaying(key)
		case phaseGameOver:
			return m.updateGameOver(key)
		}
	}

	return m, nil
}

func (m Model) updateDifficulty(key string) (tea.Model, tea.Cmd) {
	switch key {
	case "1":
		m.game = NewGame(Easy)
		m.phase = phasePlaying
		m.message = ""
	case "2":
		m.game = NewGame(Medium)
		m.phase = phasePlaying
		m.message = ""
	case "3":
		m.game = NewGame(Hard)
		m.phase = phasePlaying
		m.message = ""
	case "q", "esc":
		m.done = true
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
		if m.cursorRow < 8 {
			m.cursorRow++
		}
	case "left", "h":
		if m.cursorCol > 0 {
			m.cursorCol--
		}
	case "right", "l":
		if m.cursorCol < 8 {
			m.cursorCol++
		}
	case "1", "2", "3", "4", "5", "6", "7", "8", "9":
		num := int(key[0] - '0')
		if m.pencilMode {
			if err := m.game.TogglePencilMark(m.cursorRow, m.cursorCol, num); err != nil {
				m.message = err.Error()
			} else {
				m.message = ""
			}
		} else {
			if err := m.game.SetCell(m.cursorRow, m.cursorCol, num); err != nil {
				m.message = err.Error()
			} else {
				m.message = ""
				if m.game.Won {
					m.phase = phaseGameOver
					m.message = "Congratulations! Puzzle solved!"
				}
			}
		}
	case "0", "delete", "backspace":
		if err := m.game.ClearCell(m.cursorRow, m.cursorCol); err != nil {
			m.message = err.Error()
		} else {
			m.message = ""
		}
	case "p":
		m.pencilMode = !m.pencilMode
		if m.pencilMode {
			m.message = "Pencil mode ON"
		} else {
			m.message = "Pencil mode OFF"
		}
	case "z":
		val, err := m.game.Hint(m.cursorRow, m.cursorCol)
		if err != nil {
			m.message = err.Error()
		} else {
			_ = m.game.SetCell(m.cursorRow, m.cursorCol, val)
			m.message = fmt.Sprintf("Hint: %d", val)
			if m.game.Won {
				m.phase = phaseGameOver
				m.message = "Congratulations! Puzzle solved!"
			}
		}
	case "n":
		m.game = NewGame(m.game.Difficulty)
		m.cursorRow = 0
		m.cursorCol = 0
		m.pencilMode = false
		m.message = "New game started"
	case "q", "esc":
		m.done = true
	}
	return m, nil
}

func (m Model) updateGameOver(key string) (tea.Model, tea.Cmd) {
	switch key {
	case "n":
		if m.game != nil {
			m.game = NewGame(m.game.Difficulty)
		} else {
			m.game = NewGame(Medium)
		}
		m.phase = phasePlaying
		m.cursorRow = 0
		m.cursorCol = 0
		m.pencilMode = false
		m.message = "New game started"
	case "d":
		m.phase = phaseDifficulty
		m.message = ""
	case "q", "esc":
		m.done = true
	}
	return m, nil
}

// View renders the complete game screen.
func (m Model) View() string {
	var sections []string

	switch m.phase {
	case phaseDifficulty:
		sections = append(sections, titleStyle.Render("S U D O K U"))
		sections = append(sections, "")
		sections = append(sections, dimStyle.Render("Select Difficulty:"))
		sections = append(sections, "")
		sections = append(sections, numberStyle.Render("1")+" "+labelStyle.Render("Easy   (38 givens)"))
		sections = append(sections, numberStyle.Render("2")+" "+labelStyle.Render("Medium (30 givens)"))
		sections = append(sections, numberStyle.Render("3")+" "+labelStyle.Render("Hard   (24 givens)"))
		sections = append(sections, "")
		sections = append(sections, footerStyle.Render("1-3 Select | Q Quit"))

	case phasePlaying:
		diffLabel := ""
		if m.game != nil {
			diffLabel = m.game.Difficulty.String()
		}
		sections = append(sections, titleStyle.Render("Sudoku")+" "+dimStyle.Render("- "+diffLabel))
		sections = append(sections, "")
		sections = append(sections, m.renderGrid())
		sections = append(sections, "")
		pencilStatus := "OFF"
		if m.pencilMode {
			pencilStatus = "ON"
		}
		sections = append(sections, dimStyle.Render(fmt.Sprintf("Pencil Mode: %s  |  %d/81 filled", pencilStatus, m.game.FilledCount())))
		if m.message != "" {
			sections = append(sections, messageStyle.Render(m.message))
		}
		sections = append(sections, footerStyle.Render("Arrow/HJKL Move | 1-9 Place | 0 Clear | P Pencil | Z Hint | N New | Q Quit"))

	case phaseGameOver:
		sections = append(sections, titleStyle.Render("S U D O K U"))
		sections = append(sections, "")
		sections = append(sections, winStyle.Render("Congratulations! Puzzle Solved!"))
		sections = append(sections, "")
		if m.game != nil {
			sections = append(sections, m.renderGrid())
			sections = append(sections, "")
		}
		sections = append(sections, footerStyle.Render("N New Game | D Difficulty | Q Quit"))
	}

	content := lipgloss.JoinVertical(lipgloss.Center, sections...)
	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, content)
}

// renderGrid builds the Sudoku board with box-drawing characters.
func (m Model) renderGrid() string {
	var b strings.Builder

	// Top border
	b.WriteString(borderStyle.Render("┏━━━┯━━━┯━━━┳━━━┯━━━┯━━━┳━━━┯━━━┯━━━┓"))
	b.WriteByte('\n')

	for r := 0; r < 9; r++ {
		// Row of cells
		for c := 0; c < 9; c++ {
			// Left wall
			if c == 0 {
				b.WriteString(borderStyle.Render("┃"))
			} else if c%3 == 0 {
				b.WriteString(borderStyle.Render("┃"))
			} else {
				b.WriteString(borderStyle.Render("│"))
			}
			b.WriteString(m.renderCell(r, c))
		}
		// Right wall
		b.WriteString(borderStyle.Render("┃"))
		b.WriteByte('\n')

		// Horizontal separator
		if r < 8 {
			if (r+1)%3 == 0 {
				// Thick separator between 3x3 boxes
				b.WriteString(borderStyle.Render("┣━━━┿━━━┿━━━╋━━━┿━━━┿━━━╋━━━┿━━━┿━━━┫"))
			} else {
				// Thin separator within a 3x3 box
				b.WriteString(borderStyle.Render("┠───┼───┼───╂───┼───┼───╂───┼───┼───┨"))
			}
			b.WriteByte('\n')
		}
	}

	// Bottom border
	b.WriteString(borderStyle.Render("┗━━━┷━━━┷━━━┻━━━┷━━━┷━━━┻━━━┷━━━┷━━━┛"))

	return b.String()
}

// renderCell returns the 3-character content for a single cell.
func (m Model) renderCell(row, col int) string {
	cell := m.game.Board[row][col]
	isCursor := row == m.cursorRow && col == m.cursorCol
	isHighlight := row == m.cursorRow || col == m.cursorCol || sameBox(row, col, m.cursorRow, m.cursorCol)

	if cell.Value != 0 {
		ch := fmt.Sprintf(" %d ", cell.Value)
		switch {
		case isCursor:
			return cursorStyle.Render(ch)
		case m.game.HasConflict(row, col):
			return conflictStyle.Render(ch)
		case cell.Given:
			if isHighlight {
				return givenHighlightStyle.Render(ch)
			}
			return givenStyle.Render(ch)
		default:
			if isHighlight {
				return playerHighlightStyle.Render(ch)
			}
			return playerStyle.Render(ch)
		}
	}

	// Empty cell: show pencil marks or middle dot.
	marks := pencilString(cell.PencilMarks)
	if marks != "" {
		padded := padCenter(marks, 3)
		if isCursor {
			return cursorStyle.Render(padded)
		}
		if isHighlight {
			return pencilHighlightStyle.Render(padded)
		}
		return pencilStyle.Render(padded)
	}

	ch := " \u00b7 " // middle dot
	if isCursor {
		return cursorStyle.Render(ch)
	}
	if isHighlight {
		return emptyHighlightStyle.Render(ch)
	}
	return emptyStyle.Render(ch)
}

// sameBox returns true if (r1,c1) and (r2,c2) are in the same 3x3 box.
func sameBox(r1, c1, r2, c2 int) bool {
	return r1/3 == r2/3 && c1/3 == c2/3
}

// pencilString returns up to 3 characters representing active pencil marks.
func pencilString(marks [9]bool) string {
	var sb strings.Builder
	for i := 0; i < 9; i++ {
		if marks[i] {
			sb.WriteByte(byte('1' + i))
			if sb.Len() >= 3 {
				break
			}
		}
	}
	return sb.String()
}

// padCenter pads s to width, centering it.
func padCenter(s string, width int) string {
	gap := width - len(s)
	if gap <= 0 {
		return s[:width]
	}
	left := gap / 2
	right := gap - left
	return strings.Repeat(" ", left) + s + strings.Repeat(" ", right)
}

// --- Styles ---

var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("15"))

	dimStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("242"))

	labelStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("15"))

	numberStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FFD700"))

	footerStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("240"))

	messageStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#DCFFDC"))

	winStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#00E632"))

	borderStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("242"))

	givenStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("15"))

	givenHighlightStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("15")).
				Background(lipgloss.Color("236"))

	playerStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#00E632"))

	playerHighlightStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#00E632")).
				Background(lipgloss.Color("236"))

	conflictStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FF4444"))

	cursorStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("0")).
			Background(lipgloss.Color("#FFD700"))

	emptyStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("240"))

	emptyHighlightStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("240")).
				Background(lipgloss.Color("236"))

	pencilStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("244"))

	pencilHighlightStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("244")).
				Background(lipgloss.Color("236"))
)
