package yahtzee

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type phase int

const (
	phaseRolling  phase = iota // Rolling dice, toggling holds
	phaseScoring               // Selecting category to score
	phaseGameOver              // Final score displayed
)

// Model is the Bubbletea model for the Yahtzee game.
type Model struct {
	game      *Game
	phase     phase
	cursor    int
	width     int
	height    int
	done      bool
	message   string
	HighScore int
}

// New creates a fresh Yahtzee model.
func New() Model {
	return Model{
		game:  NewGame(),
		phase: phaseRolling,
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
		case phaseRolling:
			return m.updateRolling(key)
		case phaseScoring:
			return m.updateScoring(key)
		case phaseGameOver:
			return m.updateGameOver(key)
		}
	}

	return m, nil
}

func (m Model) updateRolling(key string) (tea.Model, tea.Cmd) {
	switch key {
	case "r", " ":
		if m.game.CanRoll() {
			m.game.Roll()
			m.message = fmt.Sprintf("Rolled! %d rolls left", m.game.RollsLeft)
		} else {
			m.message = "No rolls left — select a category"
		}
	case "1", "2", "3", "4", "5":
		idx := int(key[0]-'0') - 1
		if m.game.CanHold() {
			m.game.ToggleHold(idx)
			if m.game.Dice.Held[idx] {
				m.message = fmt.Sprintf("Holding die %d", idx+1)
			} else {
				m.message = fmt.Sprintf("Released die %d", idx+1)
			}
		} else {
			m.message = "Roll first before holding dice"
		}
	case "tab", "enter":
		if m.game.CanScore() {
			m.phase = phaseScoring
			m.cursor = m.firstUnusedCategory()
			m.message = "Select a category to score"
		} else {
			m.message = "Roll first"
		}
	case "q", "esc":
		m.done = true
	}
	return m, nil
}

func (m Model) updateScoring(key string) (tea.Model, tea.Cmd) {
	switch key {
	case "up", "k":
		m.cursor = m.nextUnusedCategory(-1)
	case "down", "j":
		m.cursor = m.nextUnusedCategory(1)
	case "enter":
		cat := Category(m.cursor)
		if err := m.game.Score(cat); err != nil {
			m.message = err.Error()
		} else if m.game.GameOver {
			m.phase = phaseGameOver
			total := m.game.GrandTotal()
			if m.HighScore > 0 && total > m.HighScore {
				m.message = fmt.Sprintf("Game over! Final score: %d — NEW HIGH SCORE!", total)
			} else if m.HighScore > 0 {
				m.message = fmt.Sprintf("Game over! Final score: %d (Best: %d)", total, m.HighScore)
			} else {
				m.message = fmt.Sprintf("Game over! Final score: %d", total)
			}
		} else {
			m.phase = phaseRolling
			m.message = fmt.Sprintf("Turn %d — press R to roll", m.game.Turn)
		}
	case "tab":
		if m.game.CanRoll() {
			m.phase = phaseRolling
			m.message = ""
		}
	case "q", "esc":
		m.done = true
	}
	return m, nil
}

func (m Model) updateGameOver(key string) (tea.Model, tea.Cmd) {
	switch key {
	case "enter", "r":
		m.game = NewGame()
		m.phase = phaseRolling
		m.message = "New game! Press R to roll"
		m.cursor = 0
	case "q", "esc":
		m.done = true
	}
	return m, nil
}

// Done returns true when the player wants to exit.
func (m Model) Done() bool {
	return m.done
}

// FinalScore returns the grand total for score tracking.
func (m Model) FinalScore() int {
	if m.game == nil {
		return 0
	}
	return m.game.GrandTotal()
}

// View renders the complete game screen.
func (m Model) View() string {
	var sections []string

	// Title
	title := titleStyle.Render("Y A H T Z E E")
	sections = append(sections, title)

	// Turn info
	turnInfo := turnStyle.Render(fmt.Sprintf("Turn %d/13  |  Rolls left: %d", m.game.Turn, m.game.RollsLeft))
	sections = append(sections, turnInfo, "")

	// Dice and scorecard side by side
	diceBlock := renderDice(m.game.Dice)
	scorecardBlock := m.renderScorecard()

	middle := lipgloss.JoinHorizontal(lipgloss.Top, diceBlock, "    ", scorecardBlock)
	sections = append(sections, middle, "")

	// Message
	if m.message != "" {
		sections = append(sections, messageStyle.Render(m.message))
	} else {
		sections = append(sections, "")
	}

	// Footer controls
	var footer string
	switch m.phase {
	case phaseRolling:
		footer = "R Roll  |  1-5 Hold/Release  |  Tab Score  |  Q Quit"
	case phaseScoring:
		footer = "↑↓ Select  |  Enter Confirm  |  Tab Back  |  Q Quit"
	case phaseGameOver:
		footer = "Enter New Game  |  Q Quit"
	}
	sections = append(sections, footerStyle.Render(footer))

	content := lipgloss.JoinVertical(lipgloss.Center, sections...)

	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, content)
}

// --- Dice rendering ---

func diceFace(value int) [5]string {
	blank := "         "
	left := "  *      "
	center := "    *    "
	right := "      *  "
	leftRight := "  *   *  "

	var rows [3]string

	switch value {
	case 1:
		rows = [3]string{blank, center, blank}
	case 2:
		rows = [3]string{right, blank, left}
	case 3:
		rows = [3]string{right, center, left}
	case 4:
		rows = [3]string{leftRight, blank, leftRight}
	case 5:
		rows = [3]string{leftRight, center, leftRight}
	case 6:
		rows = [3]string{leftRight, leftRight, leftRight}
	default:
		rows = [3]string{blank, blank, blank}
	}

	return [5]string{
		"┌─────────┐",
		"│" + rows[0] + "│",
		"│" + rows[1] + "│",
		"│" + rows[2] + "│",
		"└─────────┘",
	}
}

func renderDice(d Dice) string {
	if d.Values == [5]int{} {
		return dimStyle.Render("  Press R to roll the dice")
	}

	// Build each die face and label
	faces := [5][5]string{}
	for i := 0; i < 5; i++ {
		faces[i] = diceFace(d.Values[i])
	}

	// Render each row across all dice
	var lines []string
	for row := 0; row < 5; row++ {
		var parts []string
		for i := 0; i < 5; i++ {
			var style lipgloss.Style
			if d.Held[i] {
				style = heldDieStyle
			} else {
				style = freeDieStyle
			}
			parts = append(parts, style.Render(faces[i][row]))
		}
		lines = append(lines, strings.Join(parts, " "))
	}

	// Labels below dice
	var labels []string
	for i := 0; i < 5; i++ {
		label := fmt.Sprintf("    [%d]    ", i+1)
		if d.Held[i] {
			labels = append(labels, heldLabelStyle.Render(label))
		} else {
			labels = append(labels, dimStyle.Render(label))
		}
	}
	lines = append(lines, strings.Join(labels, " "))

	return strings.Join(lines, "\n")
}

// --- Scorecard rendering ---

func (m Model) renderScorecard() string {
	var left, right []string

	left = append(left, headerStyle.Render("UPPER SECTION"))
	right = append(right, headerStyle.Render("LOWER SECTION"))

	// Upper section
	for cat := Ones; cat <= Sixes; cat++ {
		left = append(left, m.renderCategoryLine(cat))
	}

	// Upper totals
	upperTotal := m.game.Scorecard.UpperTotal()
	left = append(left, "", dimStyle.Render(fmt.Sprintf("Upper Total ... %d/63", upperTotal)))
	bonus := m.game.Scorecard.UpperBonus()
	if bonus > 0 {
		left = append(left, potentialStyle.Render(fmt.Sprintf("Bonus ......... %d", bonus)))
	} else {
		left = append(left, dimStyle.Render("Bonus ......... -"))
	}

	// Lower section
	for cat := ThreeOfAKind; cat <= Chance; cat++ {
		right = append(right, m.renderCategoryLine(cat))
	}

	// Yahtzee bonus and total
	right = append(right, "")
	if m.game.YahtzeeBonuses > 0 {
		right = append(right, potentialStyle.Render(fmt.Sprintf("Yahtzee Bonus . %d", m.game.YahtzeeBonuses*100)))
	} else {
		right = append(right, dimStyle.Render("Yahtzee Bonus . -"))
	}
	right = append(right, categoryNameStyle.Render(fmt.Sprintf("TOTAL ......... %d", m.game.GrandTotal())))

	// Pad shorter column
	for len(left) < len(right) {
		left = append(left, "")
	}
	for len(right) < len(left) {
		right = append(right, "")
	}

	leftCol := strings.Join(left, "\n")
	rightCol := strings.Join(right, "\n")

	return lipgloss.JoinHorizontal(lipgloss.Top, leftCol, "   ", rightCol)
}

func (m Model) renderCategoryLine(cat Category) string {
	name := cat.Name()
	// Pad name to consistent width
	padded := name + strings.Repeat(".", 20-len(name))

	if m.game.Scorecard.Used[cat] {
		val := m.game.Scorecard.Scores[cat]
		if val == 0 {
			return zeroScoreStyle.Render(padded+" ") + zeroScoreStyle.Render("0")
		}
		return categoryNameStyle.Render(padded+" ") + scoredStyle.Render(fmt.Sprintf("%d", val))
	}

	// Show potential score
	potential := ""
	if m.game.CanScore() {
		p := ScoreCategory(m.game.Dice.Values, cat)
		potential = fmt.Sprintf("[%d]", p)
	} else {
		potential = " -"
	}

	isCursor := m.phase == phaseScoring && int(cat) == m.cursor

	if isCursor {
		return cursorCatStyle.Render(padded+" ") + cursorCatStyle.Render(potential) + cursorCatStyle.Render(" <")
	}
	return categoryNameStyle.Render(padded+" ") + potentialStyle.Render(potential)
}

// --- Cursor navigation ---

func (m Model) firstUnusedCategory() int {
	for i := 0; i < int(NumCategories); i++ {
		if !m.game.Scorecard.Used[i] {
			return i
		}
	}
	return 0
}

func (m Model) nextUnusedCategory(direction int) int {
	start := m.cursor
	pos := start
	for range int(NumCategories) {
		pos += direction
		if pos < 0 {
			pos = int(NumCategories) - 1
		}
		if pos >= int(NumCategories) {
			pos = 0
		}
		if !m.game.Scorecard.Used[pos] {
			return pos
		}
	}
	return start
}

// --- Styles ---

var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("15"))

	turnStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("242"))

	headerStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("15")).
			Underline(true)

	categoryNameStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("15"))

	scoredStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("15"))

	zeroScoreStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("240"))

	potentialStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#00E632"))

	cursorCatStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFD700")).
			Bold(true)

	heldDieStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFD700"))

	freeDieStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("242"))

	heldLabelStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFD700")).
			Bold(true)

	dimStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("240"))

	messageStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#DCFFDC"))

	footerStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("240"))
)
