package menu

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/herbhall/cli-play/internal/scores"
)

// GameChoice represents a selectable game entry.
type GameChoice struct {
	Name        string
	Description string
}

// Games is the list of available games.
var Games = []GameChoice{
	{"Yahtzee", "Roll dice, fill your scorecard"},
	{"Blackjack", "Beat the dealer to 21"},
	{"Wordle", "Guess the 5-letter word"},
	{"Minesweeper", "Clear the field, dodge the mines"},
	{"Sudoku", "Fill the grid with logic"},
	{"2048", "Slide and merge to 2048"},
	{"Hangman", "Guess the word letter by letter"},
	{"Tic-Tac-Toe", "Beat the AI on a 3x3 grid"},
	{"Mastermind", "Break the secret color code"},
	{"Memory", "Find all matching card pairs"},
	{"Connect Four", "Drop discs, get four in a row"},
	{"Fifteen Puzzle", "Slide tiles into order"},
	{"Snake", "Eat food, grow, don't crash"},
	{"Tetris", "Stack and clear falling blocks"},
	{"Solitaire", "Classic Klondike card game"},
	{"Typing Test", "Test your typing speed"},
}

// SettingsIndex is the menu index for the Settings entry.
const SettingsIndex = 16

// allChoices returns Games plus the Settings entry.
var allChoices = append(Games, GameChoice{"Settings", "Preferences and configuration"})

// menuRow represents a single row in the rendered menu. It is either a
// category header (gameIndex == -1) or a selectable game/settings entry.
type menuRow struct {
	gameIndex    int // -1 for category header, 0-15 for games, SettingsIndex for settings
	displayIndex int // sequential position among selectable items (for shortcut keys)
}

// buildRows constructs the flat list of menu rows from categories + settings.
func buildRows() []menuRow {
	rows := make([]menuRow, 0, len(Games)+len(categories)+2) // games + headers + settings
	displayIdx := 0
	for catIdx := range categories {
		// Category header row.
		rows = append(rows, menuRow{gameIndex: -(catIdx + 1), displayIndex: -1})
		// Game rows in this category.
		for _, gi := range categories[catIdx].Indices {
			rows = append(rows, menuRow{gameIndex: gi, displayIndex: displayIdx})
			displayIdx++
		}
	}
	// Blank separator is handled in View, not as a row.
	// Settings row.
	rows = append(rows, menuRow{gameIndex: SettingsIndex, displayIndex: displayIdx})
	return rows
}

var menuRows = buildRows()

// isSelectable returns true if this row can be cursor-selected.
func (r menuRow) isSelectable() bool {
	return r.gameIndex >= 0
}

// Tick messages.
type (
	tipTickMsg   struct{}
	blinkTickMsg struct{}
	timerTickMsg struct{}
	animTickMsg  struct{}
)

const (
	tipInterval   = 4 * time.Second
	blinkInterval = 500 * time.Millisecond
	timerInterval = 1 * time.Minute
	animInterval  = 40 * time.Millisecond
)

func tipTick() tea.Cmd {
	return tea.Tick(tipInterval, func(time.Time) tea.Msg { return tipTickMsg{} })
}

func blinkTick() tea.Cmd {
	return tea.Tick(blinkInterval, func(time.Time) tea.Msg { return blinkTickMsg{} })
}

func timerTick() tea.Cmd {
	return tea.Tick(timerInterval, func(time.Time) tea.Msg { return timerTickMsg{} })
}

func animTick() tea.Cmd {
	return tea.Tick(animInterval, func(time.Time) tea.Msg { return animTickMsg{} })
}

// AnimCmd returns a tea.Cmd that starts the entrance animation tick.
// The app model calls this when returning from a game.
func AnimCmd() tea.Cmd {
	return animTick()
}

// Model is the game selection menu.
type Model struct {
	choices  []GameChoice
	cursor   int // index into menuRows (only lands on selectable rows)
	width    int
	height   int
	selected int
	quitting bool
	scores   *scores.Store

	// Tip ticker (#25).
	tipIndex int

	// Cursor blink (#28).
	blinkOn bool

	// Session stats (#26).
	gamesPlayed  int
	sessionStart time.Time
	sessionMins  int

	// Entrance animation (#27).
	animStep    int  // -1 = no animation, 0..N = rows revealed so far
	showWelcome bool // flash "Welcome back!" briefly
}

// New creates a menu model with optional score display.
func New(s *scores.Store) Model {
	return Model{
		choices:      allChoices,
		cursor:       firstSelectableRow(),
		selected:     -1,
		scores:       s,
		blinkOn:      true,
		sessionStart: time.Now(),
		animStep:     -1,
	}
}

// firstSelectableRow returns the index of the first selectable row.
func firstSelectableRow() int {
	for i := range menuRows {
		if menuRows[i].isSelectable() {
			return i
		}
	}
	return 0
}

// Init starts background tickers.
func (m Model) Init() tea.Cmd {
	return tea.Batch(tipTick(), blinkTick(), timerTick())
}

// IncrementGamesPlayed bumps the session game counter. Called by the
// app model when returning from a game.
func (m *Model) IncrementGamesPlayed() {
	m.gamesPlayed++
}

// TriggerEntrance starts the entrance animation (items appear one by one).
func (m *Model) TriggerEntrance() {
	m.animStep = 0
	m.showWelcome = true
}

// Update handles key navigation, ticks, and quick-select.
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tipTickMsg:
		m.tipIndex = (m.tipIndex + 1) % len(tips)
		return m, tipTick()

	case blinkTickMsg:
		m.blinkOn = !m.blinkOn
		return m, blinkTick()

	case timerTickMsg:
		m.sessionMins = int(time.Since(m.sessionStart).Minutes())
		return m, timerTick()

	case animTickMsg:
		if m.animStep >= 0 {
			m.animStep++
			totalRows := len(menuRows)
			if m.animStep > totalRows+5 { // +5 for footer elements
				m.animStep = -1
				m.showWelcome = false
			}
			return m, animTick()
		}
		return m, nil

	case tea.KeyMsg:
		// During entrance animation, any key finishes it instantly.
		if m.animStep >= 0 {
			m.animStep = -1
			m.showWelcome = false
			return m, nil
		}

		cols := m.columnCount()
		switch msg.String() {
		case "up", "k":
			m.cursor = m.skipSelectableN(m.cursor, -cols)
		case "down", "j":
			m.cursor = m.skipSelectableN(m.cursor, cols)
		case "left", "h":
			m.cursor = m.prevSelectable(m.cursor)
		case "right", "l":
			m.cursor = m.nextSelectable(m.cursor)
		case "enter":
			m.selected = menuRows[m.cursor].gameIndex
		case "q", "esc":
			m.quitting = true
		default:
			// Number key quick-select (#23).
			if idx, ok := m.shortcutToGameIndex(msg.String()); ok {
				m.selected = idx
			}
		}
	}

	return m, nil
}

// nextSelectable finds the next selectable row after current, wrapping.
func (m Model) nextSelectable(current int) int {
	n := len(menuRows)
	for i := 1; i < n; i++ {
		idx := (current + i) % n
		if menuRows[idx].isSelectable() {
			return idx
		}
	}
	return current
}

// prevSelectable finds the previous selectable row before current, wrapping.
func (m Model) prevSelectable(current int) int {
	n := len(menuRows)
	for i := 1; i < n; i++ {
		idx := (current - i + n) % n
		if menuRows[idx].isSelectable() {
			return idx
		}
	}
	return current
}

// skipSelectableN jumps n selectable items forward (positive) or backward (negative).
func (m Model) skipSelectableN(current, n int) int {
	pos := current
	step := 1
	if n < 0 {
		step = -1
		n = -n
	}
	for i := 0; i < n; i++ {
		if step > 0 {
			pos = m.nextSelectable(pos)
		} else {
			pos = m.prevSelectable(pos)
		}
	}
	return pos
}

// renderEntry renders a single game/settings entry. When compact is true
// (multi-column mode), descriptions and high scores are omitted.
func (m Model) renderEntry(row menuRow, rowIdx int, compact bool) string {
	var e strings.Builder

	// Cursor indicator.
	indicator := "   "
	ns := nameStyle
	if rowIdx == m.cursor {
		if m.blinkOn {
			indicator = " \u25b6 " // filled triangle
		} else {
			indicator = " \u25b7 " // outline triangle
		}
		ns = nameSelectedStyle
	}
	e.WriteString(cursorStyle.Render(indicator))

	// Shortcut key.
	if row.gameIndex < SettingsIndex && row.displayIndex >= 0 {
		label := shortcutLabel(row.displayIndex)
		e.WriteString(shortcutStyle.Render(fmt.Sprintf("[%s] ", label)))
	} else {
		e.WriteString("    ")
	}

	// Icon.
	if icon, ok := gameIcon[row.gameIndex]; ok {
		e.WriteString(iconStyle.Render(fmt.Sprintf("%-3s", icon)))
	} else if row.gameIndex == SettingsIndex {
		e.WriteString(iconStyle.Render("\u2699  "))
	}

	// Name.
	name := ""
	if row.gameIndex == SettingsIndex {
		name = "Settings"
	} else if row.gameIndex >= 0 && row.gameIndex < len(Games) {
		name = Games[row.gameIndex].Name
	}
	e.WriteString(ns.Render(fmt.Sprintf("%-16s", name)))

	// Description and high score only in single-column mode.
	if !compact {
		desc := ""
		if row.gameIndex == SettingsIndex {
			desc = "Preferences and configuration"
		} else if row.gameIndex >= 0 && row.gameIndex < len(Games) {
			desc = Games[row.gameIndex].Description
		}
		e.WriteString(descStyle.Render(desc))

		if hs := m.highScoreLabel(row.gameIndex); hs != "" {
			e.WriteString("  ")
			e.WriteString(highScoreStyle.Render(hs))
		}
	}

	return e.String()
}

// shortcutToGameIndex maps a key press to a game index.
// 1-9 -> display positions 0-8, 0 -> position 9, a-f -> positions 10-15.
// Returns the original Games index and true if valid.
func (m Model) shortcutToGameIndex(key string) (int, bool) {
	displayIdx := -1
	if len(key) == 1 {
		ch := key[0]
		switch {
		case ch >= '1' && ch <= '9':
			displayIdx = int(ch - '1')
		case ch == '0':
			displayIdx = 9
		case ch >= 'a' && ch <= 'f':
			displayIdx = int(ch-'a') + 10
		}
	}
	if displayIdx < 0 {
		return -1, false
	}
	// Find the row with this display index.
	for i := range menuRows {
		if menuRows[i].displayIndex == displayIdx && menuRows[i].gameIndex >= 0 && menuRows[i].gameIndex < SettingsIndex {
			return menuRows[i].gameIndex, true
		}
	}
	return -1, false
}

// Styles.
var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#00FF87"))

	categoryStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("243")).
			Italic(true)

	cursorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFD700"))

	nameStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FFFFFF"))

	nameSelectedStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("#FFD700"))

	descStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("243"))

	footerStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("240"))

	tipStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("240")).
			Italic(true)

	highScoreStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFD700"))

	statsStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("243"))

	previewBorder = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("240")).
			Foreground(lipgloss.Color("250")).
			Padding(0, 1).
			Width(42)

	previewTitleStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("#FFFFFF"))

	previewDimStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("243"))

	panelBorder = lipgloss.RoundedBorder()

	shortcutStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("243"))

	iconStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("250"))

	welcomeStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#00FF87"))
)

// compactTitle is a single-line styled title for the menu header.
const compactTitle = "CLI PLAY"

// colWidth is the fixed visual width of a single game entry in multi-column mode.
// cursor(3) + shortcut(4) + icon(3) + name(16) + gap(2) = 28.
const colWidth = 28

// columnCount returns how many game columns fit in the available width.
func (m Model) columnCount() int {
	innerW := m.width - 6 // border(2) + padding(4)
	cols := innerW / colWidth
	if cols < 1 {
		cols = 1
	}
	if cols > 3 {
		cols = 3
	}
	return cols
}

// contentHeight estimates the number of inner lines used by the menu,
// excluding the border/padding chrome (4 lines: top border+pad, bottom pad+border).
func (m Model) contentHeight(showTitle, showStats, showPreview, showTip bool) int {
	cols := m.columnCount()
	lines := 0
	if showTitle {
		lines++ // compact single-line title
	}
	if showStats {
		lines++ // stats line
	}
	// Game list: category headers + game rows in grid (no inter-category blanks).
	for _, cat := range categories {
		lines++                                       // header
		lines += (len(cat.Indices) + cols - 1) / cols // game rows
	}
	lines++ // settings row
	if showPreview {
		lines += 8 // border+title+rules+controls ~ 7-8 lines
	}
	if showTip {
		lines++
	}
	lines++ // footer
	return lines
}

// View renders the menu with categories, icons, preview, stats, and tips.
func (m Model) View() string {
	var b strings.Builder

	// Available inner height: total minus border (2).
	innerH := m.height - 2

	// Progressively hide elements to fit: preview first, then title, then stats, then tips.
	showTitle := true
	showStats := true
	showPreview := true
	showTip := true

	if m.contentHeight(showTitle, showStats, showPreview, showTip) > innerH {
		showPreview = false
	}
	if m.contentHeight(showTitle, showStats, showPreview, showTip) > innerH {
		showTitle = false
	}
	if m.contentHeight(showTitle, showStats, showPreview, showTip) > innerH {
		showStats = false
	}
	if m.contentHeight(showTitle, showStats, showPreview, showTip) > innerH {
		showTip = false
	}

	// Compact title on a single line (#22).
	if showTitle {
		b.WriteString(titleStyle.Render(compactTitle))
		b.WriteString("\n")
	}

	// Session stats bar (#26).
	if showStats {
		elapsed := m.sessionMins
		if elapsed == 0 && time.Since(m.sessionStart) >= 30*time.Second {
			elapsed = 1
		}
		statsLine := fmt.Sprintf("Games played: %d | Session: %dm", m.gamesPlayed, elapsed)
		b.WriteString(statsStyle.Render(statsLine))
		b.WriteString("\n")
	}

	// Welcome back flash (#27).
	if m.showWelcome {
		b.WriteString(welcomeStyle.Render("  Welcome back!"))
		b.WriteString("\n")
	}

	// Game list with categories (#30), icons (#24), shortcuts (#23).
	cols := m.columnCount()
	compact := cols > 1
	visualRow := 0
	gameRows := make([]int, 0, 8) // reused per category

	for catI := range categories {
		if m.animStep >= 0 && visualRow > m.animStep {
			break
		}

		// Category header (no inter-category blank lines to save vertical space).
		cat := categories[catI]
		b.WriteString(categoryStyle.Render(fmt.Sprintf("  %s %s", cat.Icon, cat.Name)))
		b.WriteString("\n")
		visualRow++

		// Collect menuRow indices for this category (in Indices order).
		gameRows = gameRows[:0]
		for _, gi := range cat.Indices {
			for ri, row := range menuRows {
				if row.gameIndex == gi {
					gameRows = append(gameRows, ri)
					break
				}
			}
		}

		// Render games in a grid.
		for i := 0; i < len(gameRows); i += cols {
			if m.animStep >= 0 && visualRow > m.animStep {
				break
			}
			for j := 0; j < cols && i+j < len(gameRows); j++ {
				ri := gameRows[i+j]
				b.WriteString(m.renderEntry(menuRows[ri], ri, compact))
				if compact && j < cols-1 && i+j+1 < len(gameRows) {
					b.WriteString("  ")
				}
			}
			b.WriteString("\n")
			visualRow++
		}
	}

	// Settings entry.
	for ri, row := range menuRows {
		if row.gameIndex == SettingsIndex {
			b.WriteString(m.renderEntry(row, ri, compact))
			b.WriteString("\n")
			break
		}
	}

	// Preview panel (#29).
	if showPreview {
		preview := m.renderPreview()
		if preview != "" {
			b.WriteString("\n")
			b.WriteString(preview)
		}
	}

	b.WriteString("\n")

	// Tip ticker (#25).
	if showTip && m.tipIndex < len(tips) {
		b.WriteString(tipStyle.Render(tips[m.tipIndex]))
		b.WriteString("\n")
	}

	// Footer.
	b.WriteString(footerStyle.Render("  \u2190\u2191\u2193\u2192 Navigate | Enter Select | 1-9/0/a-f Quick Select | Q Quit"))

	// Wrap in bordered panel (#21).
	panel := lipgloss.NewStyle().
		Border(panelBorder).
		BorderForeground(lipgloss.Color("240")).
		Padding(0, 2).
		Render(b.String())

	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, panel)
}

// renderPreview returns the preview panel for the currently highlighted game.
func (m Model) renderPreview() string {
	gameIdx := -1
	if m.cursor >= 0 && m.cursor < len(menuRows) {
		row := menuRows[m.cursor]
		if row.gameIndex >= 0 && row.gameIndex < len(Games) {
			gameIdx = row.gameIndex
		}
	}
	if gameIdx < 0 {
		return ""
	}

	p, ok := previews[gameIdx]
	if !ok {
		return ""
	}

	var pb strings.Builder
	pb.WriteString(previewTitleStyle.Render(Games[gameIdx].Name))
	pb.WriteString("\n")
	pb.WriteString(previewDimStyle.Render(p.Rules))
	pb.WriteString("\n\n")
	pb.WriteString(previewDimStyle.Render("Controls: " + p.Controls))

	return previewBorder.Render(pb.String())
}


// TransitionText returns the menu layout as plain text (no ANSI, no border)
// for the transition reveal animation. It mirrors the View() content using
// the same responsive logic (column count, show/hide flags) so the revealed
// text matches the real menu layout. The caller (transition) centers it.
func (m Model) TransitionText() string {
	var b strings.Builder

	innerH := m.height - 2
	showTitle := true
	showStats := true
	showPreview := true
	showTip := true

	if m.contentHeight(showTitle, showStats, showPreview, showTip) > innerH {
		showPreview = false
	}
	if m.contentHeight(showTitle, showStats, showPreview, showTip) > innerH {
		showTitle = false
	}
	if m.contentHeight(showTitle, showStats, showPreview, showTip) > innerH {
		showStats = false
	}
	if m.contentHeight(showTitle, showStats, showPreview, showTip) > innerH {
		showTip = false
	}

	if showTitle {
		b.WriteString(compactTitle)
		b.WriteString("\n")
	}
	if showStats {
		b.WriteString(fmt.Sprintf("Games played: %d | Session: 0m", m.gamesPlayed))
		b.WriteString("\n")
	}

	cols := m.columnCount()
	compact := cols > 1

	for catI := range categories {
		cat := categories[catI]
		b.WriteString(fmt.Sprintf("  %s\n", cat.Name))

		// Collect game rows for this category.
		var catGames []int
		for _, gi := range cat.Indices {
			for ri, row := range menuRows {
				if row.gameIndex == gi {
					catGames = append(catGames, ri)
					break
				}
			}
		}

		// Render in grid.
		for i := 0; i < len(catGames); i += cols {
			for j := 0; j < cols && i+j < len(catGames); j++ {
				ri := catGames[i+j]
				row := menuRows[ri]
				b.WriteString(plainEntry(row, ri == m.cursor, compact))
				if compact && j < cols-1 && i+j+1 < len(catGames) {
					b.WriteString("  ")
				}
			}
			b.WriteString("\n")
		}
	}

	// Settings.
	for ri, row := range menuRows {
		if row.gameIndex == SettingsIndex {
			b.WriteString(plainEntry(row, ri == m.cursor, compact))
			b.WriteString("\n")
			break
		}
	}

	b.WriteString("\n")
	if showTip && m.tipIndex < len(tips) {
		b.WriteString(tips[m.tipIndex])
		b.WriteString("\n")
	}
	b.WriteString("  \u2190\u2191\u2193\u2192 Navigate | Enter Select | 1-9/0/a-f Quick Select | Q Quit")

	return b.String()
}

// plainEntry renders a single menu entry as plain text (no ANSI).
func plainEntry(row menuRow, selected, compact bool) string {
	var e strings.Builder

	// Cursor.
	if selected {
		e.WriteString(" \u25b6 ")
	} else {
		e.WriteString("   ")
	}

	// Shortcut.
	if row.gameIndex < SettingsIndex && row.displayIndex >= 0 {
		e.WriteString(fmt.Sprintf("[%s] ", shortcutLabel(row.displayIndex)))
	} else {
		e.WriteString("    ")
	}

	// Icon.
	if icon, ok := gameIcon[row.gameIndex]; ok {
		e.WriteString(fmt.Sprintf("%-3s", icon))
	} else if row.gameIndex == SettingsIndex {
		e.WriteString("\u2699  ")
	}

	// Name.
	name := ""
	if row.gameIndex == SettingsIndex {
		name = "Settings"
	} else if row.gameIndex >= 0 && row.gameIndex < len(Games) {
		name = Games[row.gameIndex].Name
	}
	e.WriteString(fmt.Sprintf("%-16s", name))

	// Description only in single-column.
	if !compact {
		if row.gameIndex == SettingsIndex {
			e.WriteString("Preferences and configuration")
		} else if row.gameIndex >= 0 && row.gameIndex < len(Games) {
			e.WriteString(Games[row.gameIndex].Description)
		}
	}

	return e.String()
}

// Selected returns the index of the selected game, or -1 if none.
func (m Model) Selected() int {
	return m.selected
}

// ResetSelection clears the selected state so the menu can be reused
// after returning from a game.
func (m *Model) ResetSelection() {
	m.selected = -1
}

// highScoreLabel returns a formatted high score string for the given game index.
func (m Model) highScoreLabel(index int) string {
	if m.scores == nil {
		return ""
	}
	switch index {
	case 0: // Yahtzee
		if e := m.scores.Get("yahtzee"); e != nil {
			return fmt.Sprintf("[Best: %d]", e.Value)
		}
	case 1: // Blackjack
		if e := m.scores.Get("blackjack"); e != nil {
			return fmt.Sprintf("[Best: $%d]", e.Value)
		}
	case 2: // Wordle
		if e := m.scores.Get("wordle"); e != nil {
			return fmt.Sprintf("[Best: %d/6]", e.Value)
		}
	case 3: // Minesweeper
		return m.bestTimeLabel("minesweeper")
	case 4: // Sudoku
		return m.bestTimeLabel("sudoku")
	case 5: // 2048
		if e := m.scores.Get("2048"); e != nil {
			return fmt.Sprintf("[Best: %d]", e.Value)
		}
	case 6: // Hangman
		if e := m.scores.Get("hangman"); e != nil {
			return fmt.Sprintf("[Best: %d wrong]", e.Value)
		}
	case 7: // Tic-Tac-Toe
		if e := m.scores.Get("tictactoe"); e != nil {
			return fmt.Sprintf("[Wins: %d]", e.Value)
		}
	case 8: // Mastermind
		if e := m.scores.Get("mastermind"); e != nil {
			return fmt.Sprintf("[Best: %d guesses]", e.Value)
		}
	case 9: // Memory
		if e := m.scores.Get("memory"); e != nil {
			return fmt.Sprintf("[Best: %d moves]", e.Value)
		}
	case 10: // Connect Four
		if e := m.scores.Get("connectfour"); e != nil {
			return fmt.Sprintf("[Wins: %d]", e.Value)
		}
	case 11: // Fifteen Puzzle
		if e := m.scores.Get("fifteenpuzzle"); e != nil {
			return fmt.Sprintf("[Best: %d moves]", e.Value)
		}
	case 12: // Snake
		if e := m.scores.Get("snake"); e != nil {
			return fmt.Sprintf("[Best: %d]", e.Value)
		}
	case 13: // Tetris
		if e := m.scores.Get("tetris"); e != nil {
			return fmt.Sprintf("[Best: %d]", e.Value)
		}
	case 14: // Solitaire
		if e := m.scores.Get("solitaire"); e != nil {
			return fmt.Sprintf("[Best: %d]", e.Value)
		}
	case 15: // Typing Test
		if e := m.scores.Get("typingtest"); e != nil {
			return fmt.Sprintf("[Best: %d WPM]", e.Value)
		}
	}
	return ""
}

// bestTimeLabel returns the best time across all difficulties for a timed game.
func (m Model) bestTimeLabel(game string) string {
	var best *scores.Entry
	for _, diff := range []string{"beginner", "intermediate", "expert", "easy", "medium", "hard"} {
		e := m.scores.GetDifficulty(game, diff)
		if e != nil && (best == nil || e.Value < best.Value) {
			best = e
		}
	}
	if best == nil {
		return ""
	}
	mins := best.Value / 60
	secs := best.Value % 60
	return fmt.Sprintf("[Best: %d:%02d]", mins, secs)
}

// Quitting returns true if the user pressed quit.
func (m Model) Quitting() bool {
	return m.quitting
}
