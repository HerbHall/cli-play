package solitaire

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type phase int

const (
	phasePlaying  phase = iota
	phaseGameOver       // win or quit with score
)

// selection tracks what the player has picked up.
type selection struct {
	source  string // "waste", "tableau"
	col     int    // tableau column (0-6)
	cardIdx int    // index within tableau column
	active  bool
}

// Model is the Bubbletea model for the Solitaire game.
type Model struct {
	game      *Game
	phase     phase
	cursor    string // "stock", "waste", "foundation", "tableau"
	tabCol    int    // current tableau column (0-6)
	tabRow    int    // current row within tableau column
	sel       selection
	width     int
	height    int
	done      bool
	message   string
	HighScore int
}

// New creates a fresh Solitaire model.
func New() Model {
	return Model{
		game:   NewGame(nil),
		phase:  phasePlaying,
		cursor: "tableau",
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
		case phaseGameOver:
			return m.updateGameOver(key)
		}
	}

	return m, nil
}

func (m Model) updatePlaying(key string) (tea.Model, tea.Cmd) {
	switch key {
	case "d":
		m.game.DrawStock()
		m.clearSelection()
		m.message = ""
	case "1", "2", "3", "4", "5", "6", "7":
		col := int(key[0]-'0') - 1
		m.cursor = "tableau"
		m.tabCol = col
		m.tabRow = m.defaultTabRow(col)
		m.message = ""
	case "s":
		m.cursor = "stock"
		m.message = ""
	case "w":
		m.cursor = "waste"
		m.message = ""
	case "left":
		m.moveLeft()
	case "right":
		m.moveRight()
	case "up":
		m.moveUp()
	case "down":
		m.moveDown()
	case "f":
		m.tryMoveToFoundation()
	case "enter", " ":
		m.handleSelect()
	case "tab":
		m.cycleArea()
	case "n":
		m.game = NewGame(nil)
		m.phase = phasePlaying
		m.clearSelection()
		m.message = "New game!"
	case "q", "esc":
		m.done = true
	}

	if m.game.Won && m.phase == phasePlaying {
		m.phase = phaseGameOver
		if m.HighScore > 0 && m.game.Score > m.HighScore {
			m.message = fmt.Sprintf("YOU WIN! Score: %d -- NEW HIGH SCORE!", m.game.Score)
		} else {
			m.message = fmt.Sprintf("YOU WIN! Score: %d in %d moves!", m.game.Score, m.game.Moves)
		}
	}

	return m, nil
}

func (m Model) updateGameOver(key string) (tea.Model, tea.Cmd) {
	switch key {
	case "n", "enter":
		m.game = NewGame(nil)
		m.phase = phasePlaying
		m.clearSelection()
		m.cursor = "tableau"
		m.tabCol = 0
		m.tabRow = 0
		m.message = "New game!"
	case "q", "esc":
		m.done = true
	}
	return m, nil
}

// Done returns true when the player wants to exit.
func (m Model) Done() bool {
	return m.done
}

// FinalScore returns the score for score tracking.
func (m Model) FinalScore() int {
	if m.game == nil {
		return 0
	}
	return m.game.Score
}

// --- Navigation helpers ---

func (m *Model) clearSelection() {
	m.sel = selection{}
}

func (m *Model) defaultTabRow(col int) int {
	pile := m.game.Tableau[col]
	if len(pile) == 0 {
		return 0
	}
	return len(pile) - 1
}

func (m *Model) moveLeft() {
	switch m.cursor {
	case "tableau":
		if m.tabCol > 0 {
			m.tabCol--
			m.tabRow = m.defaultTabRow(m.tabCol)
		}
	case "waste":
		m.cursor = "stock"
	case "foundation":
		m.cursor = "waste"
	}
}

func (m *Model) moveRight() {
	switch m.cursor {
	case "tableau":
		if m.tabCol < 6 {
			m.tabCol++
			m.tabRow = m.defaultTabRow(m.tabCol)
		}
	case "stock":
		m.cursor = "waste"
	case "waste":
		m.cursor = "foundation"
	}
}

func (m *Model) moveUp() {
	if m.cursor == "tableau" {
		fui := m.game.FaceUpIndex(m.tabCol)
		if fui >= 0 && m.tabRow > fui {
			m.tabRow--
		} else {
			m.cursor = "stock"
		}
	}
}

func (m *Model) moveDown() {
	switch m.cursor {
	case "stock", "waste", "foundation":
		m.cursor = "tableau"
		m.tabRow = m.defaultTabRow(m.tabCol)
	case "tableau":
		pile := m.game.Tableau[m.tabCol]
		if m.tabRow < len(pile)-1 {
			m.tabRow++
		}
	}
}

func (m *Model) cycleArea() {
	switch m.cursor {
	case "stock":
		m.cursor = "waste"
	case "waste":
		m.cursor = "foundation"
	case "foundation":
		m.cursor = "tableau"
		m.tabRow = m.defaultTabRow(m.tabCol)
	case "tableau":
		m.cursor = "stock"
	}
}

func (m *Model) handleSelect() {
	switch m.cursor {
	case "stock":
		m.game.DrawStock()
		m.clearSelection()
		m.message = ""
	case "waste":
		if m.sel.active && m.sel.source == "waste" {
			m.clearSelection()
			return
		}
		if _, ok := m.game.WasteTop(); ok {
			m.sel = selection{source: "waste", active: true}
			m.message = "Card selected. Press 1-7 or F to place."
		}
	case "tableau":
		if m.sel.active {
			m.placeSelection()
			return
		}
		pile := m.game.Tableau[m.tabCol]
		if len(pile) == 0 {
			return
		}
		if m.tabRow >= 0 && m.tabRow < len(pile) && pile[m.tabRow].FaceUp {
			m.sel = selection{
				source:  "tableau",
				col:     m.tabCol,
				cardIdx: m.tabRow,
				active:  true,
			}
			m.message = "Stack selected. Press 1-7 or F to place."
		}
	case "foundation":
		// Foundation cards are not movable in this implementation.
		m.message = "Cannot pick up from foundation."
	}
}

func (m *Model) placeSelection() {
	switch m.sel.source {
	case "waste":
		if m.cursor == "tableau" {
			if m.game.MoveWasteToTableau(m.tabCol) {
				m.message = ""
			} else {
				m.message = "Invalid move."
			}
		}
	case "tableau":
		if m.cursor == "tableau" {
			if m.game.MoveTableauToTableau(m.sel.col, m.sel.cardIdx, m.tabCol) {
				m.message = ""
			} else {
				m.message = "Invalid move."
			}
		}
	}
	m.clearSelection()
	m.tabRow = m.defaultTabRow(m.tabCol)
}

func (m *Model) tryMoveToFoundation() {
	switch m.cursor {
	case "waste":
		if m.game.MoveWasteToFoundation() {
			m.message = ""
		} else {
			m.message = "Cannot move to foundation."
		}
	case "tableau":
		if m.game.MoveTableauToFoundation(m.tabCol) {
			m.message = ""
		} else {
			m.message = "Cannot move to foundation."
		}
	default:
		m.message = "Select waste or tableau first."
	}
	m.clearSelection()
}

// --- View rendering ---

// View renders the complete game screen.
func (m Model) View() string {
	// Score and moves
	info := scoreStyle.Render(fmt.Sprintf("Score: %d  |  Moves: %d", m.game.Score, m.game.Moves))
	if m.HighScore > 0 {
		if m.game.Score > m.HighScore {
			info += "  " + winStyle.Render("NEW HIGH SCORE!")
		} else {
			info += "  " + footerStyle.Render(fmt.Sprintf("Best: %d", m.HighScore))
		}
	}

	// Message line
	msg := ""
	if m.message != "" {
		msg = messageStyle.Render(m.message)
	}

	// Footer
	var footer string
	switch m.phase {
	case phasePlaying:
		footer = "D Draw | 1-7 Column | Enter Select/Place | F Foundation | N New | Q Quit"
	case phaseGameOver:
		footer = "N New Game | Q Quit"
	}

	sections := []string{
		titleStyle.Render("S O L I T A I R E"),
		info,
		"",
		m.renderTopRow(),
		"",
		m.renderTableau(),
		"",
		msg,
		footerStyle.Render(footer),
	}

	content := lipgloss.JoinVertical(lipgloss.Left, sections...)

	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, content)
}

func (m Model) renderTopRow() string {
	// Stock
	var stockStr string
	if len(m.game.Stock) > 0 {
		stockStr = faceDownStyle.Render("[##]")
	} else {
		stockStr = emptyStyle.Render("[  ]")
	}
	if m.cursor == "stock" {
		stockStr = selectedStyle.Render("[##]")
		if len(m.game.Stock) == 0 {
			stockStr = selectedStyle.Render("[  ]")
		}
	}

	// Waste
	var wasteStr string
	if card, ok := m.game.WasteTop(); ok {
		style := m.cardStyle(card)
		if m.cursor == "waste" || (m.sel.active && m.sel.source == "waste") {
			style = selectedStyle
		}
		wasteStr = style.Render(m.cardText(card))
	} else {
		if m.cursor == "waste" {
			wasteStr = selectedStyle.Render("[  ]")
		} else {
			wasteStr = emptyStyle.Render("[  ]")
		}
	}

	// Gap
	gap := "    "

	// Foundations
	fStrs := make([]string, 4)
	for i := range fStrs {
		if card, ok := m.game.FoundationTop(i); ok {
			fLen := len(m.game.Foundations[i])
			style := m.cardStyle(card)
			if fLen == 13 {
				style = foundationCompleteStyle
			}
			if m.cursor == "foundation" {
				style = selectedStyle
			}
			fStrs[i] = style.Render(m.cardText(card))
		} else {
			if m.cursor == "foundation" {
				fStrs[i] = selectedStyle.Render("[  ]")
			} else {
				fStrs[i] = emptyStyle.Render("[  ]")
			}
		}
	}

	return stockStr + " " + wasteStr + gap +
		fStrs[0] + " " + fStrs[1] + " " + fStrs[2] + " " + fStrs[3]
}

func (m Model) renderTableau() string {
	// Find max column height for alignment.
	maxLen := 0
	for col := 0; col < 7; col++ {
		if len(m.game.Tableau[col]) > maxLen {
			maxLen = len(m.game.Tableau[col])
		}
	}
	if maxLen == 0 {
		maxLen = 1
	}

	var rows []string
	for row := 0; row < maxLen; row++ {
		var cols []string
		for col := 0; col < 7; col++ {
			pile := m.game.Tableau[col]
			if row >= len(pile) {
				if row == 0 {
					// Empty column marker.
					if m.cursor == "tableau" && m.tabCol == col {
						cols = append(cols, selectedStyle.Render("[  ]"))
					} else {
						cols = append(cols, emptyStyle.Render("[  ]"))
					}
				} else {
					cols = append(cols, "    ")
				}
				continue
			}

			card := pile[row]
			isSelected := m.sel.active && m.sel.source == "tableau" &&
				m.sel.col == col && row >= m.sel.cardIdx
			isCursor := m.cursor == "tableau" && m.tabCol == col && m.tabRow == row

			switch {
			case !card.FaceUp:
				cols = append(cols, faceDownStyle.Render("[##]"))
			case isSelected:
				cols = append(cols, selectedStyle.Render(m.cardText(card)))
			case isCursor:
				cols = append(cols, selectedStyle.Render(m.cardText(card)))
			default:
				cols = append(cols, m.cardStyle(card).Render(m.cardText(card)))
			}
		}
		rows = append(rows, strings.Join(cols, " "))
	}

	return strings.Join(rows, "\n")
}

func (m Model) cardText(c Card) string {
	// Format: [A♠] or [10♥]. The rank is 1 char (A,2-9,J,Q,K) or 2 (10).
	// The suit symbol is 1 display character. We use the same bracket format
	// for all cards; 10 is naturally one char wider.
	return "[" + c.Label() + "]"
}

func (m Model) cardStyle(c Card) lipgloss.Style {
	if c.Suit.IsRed() {
		return redCardStyle
	}
	return blackCardStyle
}

// --- Styles ---

var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#DCFFDC"))

	scoreStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#DCFFDC"))

	redCardStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF0000"))

	blackCardStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("15"))

	faceDownStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("242"))

	emptyStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("242"))

	selectedStyle = lipgloss.NewStyle().
			Background(lipgloss.Color("#333333")).
			Foreground(lipgloss.Color("15"))

	foundationCompleteStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#00E632"))

	winStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#00E632"))

	messageStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#DCFFDC"))

	footerStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("240"))
)
