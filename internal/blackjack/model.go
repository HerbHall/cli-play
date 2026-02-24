package blackjack

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Styles for the blackjack UI.
var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FFD700"))

	cardWhiteStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("15"))

	cardRedStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FF4444"))

	winStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#00E632"))

	lossStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FF4444"))

	pushStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FFD700"))

	labelStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("242"))

	helpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("240"))

	highScoreStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FFD700"))

	holeCardText = "[??]"
)

// Model is the Bubbletea model for the blackjack game screen.
type Model struct {
	game      *Game
	width     int
	height    int
	done      bool
	HighScore int
}

// New creates a blackjack model with a starting balance of 1000.
func New() Model {
	return Model{game: NewGame(1000, nil)}
}

// Init returns nil; no initial command needed.
func (m Model) Init() tea.Cmd {
	return nil
}

// Update handles input and window size messages.
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

		switch m.game.Phase {
		case PhaseBetting:
			return m.updateBetting(key)
		case PhasePlayerTurn:
			return m.updatePlayerTurn(key)
		case PhaseResult:
			return m.updateResult(key)
		}
	}

	return m, nil
}

func (m Model) updateBetting(key string) (tea.Model, tea.Cmd) {
	switch key {
	case "1":
		m.game.PlaceBet(10) //nolint:errcheck // UI constrains valid input
	case "2":
		m.game.PlaceBet(25) //nolint:errcheck // UI constrains valid input
	case "3":
		m.game.PlaceBet(50) //nolint:errcheck // UI constrains valid input
	case "4":
		m.game.PlaceBet(100) //nolint:errcheck // UI constrains valid input
	case "q":
		m.done = true
	}
	return m, nil
}

func (m Model) updatePlayerTurn(key string) (tea.Model, tea.Cmd) {
	switch key {
	case "h":
		m.game.Hit()
	case "s":
		m.game.Stand()
	case "d":
		if m.game.CanDoubleDown() {
			m.game.DoubleDown() //nolint:errcheck // CanDoubleDown pre-validates
		}
	}
	return m, nil
}

// FinalScore returns the player's balance for score tracking.
func (m Model) FinalScore() int {
	if m.game == nil {
		return 0
	}
	return m.game.Balance
}

func (m Model) updateResult(key string) (tea.Model, tea.Cmd) {
	switch key {
	case "enter", "n":
		m.game.NewRound()
	case "q":
		m.done = true
	}
	return m, nil
}

// View renders the blackjack table.
func (m Model) View() string {
	var b strings.Builder

	b.WriteString(titleStyle.Render("--- BLACKJACK ---"))
	b.WriteString("\n\n")

	b.WriteString(m.renderDealerHand())
	b.WriteString("\n\n")
	b.WriteString(m.renderPlayerHand())
	b.WriteString("\n\n")
	b.WriteString(m.renderStats())
	b.WriteString("\n")

	if m.game.Phase == PhaseResult && m.game.Message != "" {
		b.WriteString("\n")
		b.WriteString(m.renderMessage())
		b.WriteString("\n")
	}

	b.WriteString("\n")
	b.WriteString(m.renderHelp())

	content := lipgloss.NewStyle().
		Align(lipgloss.Center).
		Render(b.String())

	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, content)
}

func (m Model) renderDealerHand() string {
	g := m.game
	var scoreStr string
	if g.HoleRevealed {
		scoreStr = fmt.Sprintf("(%d)", g.DealerHand.Score())
	} else if len(g.DealerHand.Cards) > 0 {
		scoreStr = "(?)"
	}

	label := labelStyle.Render("Dealer " + scoreStr)

	var cards string
	switch {
	case len(g.DealerHand.Cards) == 0:
		cards = ""
	case g.HoleRevealed:
		cards = renderCards(g.DealerHand.Cards)
	default:
		shown := styledCard(g.DealerHand.Cards[0])
		cards = shown + "  " + helpStyle.Render(holeCardText)
	}

	return label + "\n" + cards
}

func (m Model) renderPlayerHand() string {
	g := m.game
	var scoreStr string
	if len(g.PlayerHand.Cards) > 0 {
		scoreStr = fmt.Sprintf("(%d)", g.PlayerHand.Score())
	}

	label := labelStyle.Render("Player " + scoreStr)
	cards := renderCards(g.PlayerHand.Cards)
	return label + "\n" + cards
}

func renderCards(cards []Card) string {
	parts := make([]string, len(cards))
	for i, c := range cards {
		parts[i] = styledCard(c)
	}
	return strings.Join(parts, "  ")
}

func styledCard(c Card) string {
	rankStr := c.Rank.String()
	suitStr := c.Suit.Symbol()
	if c.Suit == Hearts || c.Suit == Diamonds {
		return cardWhiteStyle.Render(rankStr) + cardRedStyle.Render(suitStr)
	}
	return cardWhiteStyle.Render(rankStr + suitStr)
}

func (m Model) renderStats() string {
	g := m.game
	balance := labelStyle.Render(fmt.Sprintf("Balance: $%d", g.Balance))
	bet := ""
	if g.Bet > 0 {
		bet = labelStyle.Render(fmt.Sprintf("    Bet: $%d", g.Bet))
	}
	stats := labelStyle.Render(fmt.Sprintf("W: %d  L: %d  P: %d",
		g.Stats.Wins, g.Stats.Losses, g.Stats.Pushes))
	line := balance + bet + "\n" + stats
	if m.HighScore > 0 {
		if g.Balance > m.HighScore {
			line += "\n" + highScoreStyle.Render("NEW HIGH SCORE!")
		} else {
			line += "\n" + helpStyle.Render(fmt.Sprintf("Best: $%d", m.HighScore))
		}
	}
	return line
}

func (m Model) renderMessage() string {
	msg := m.game.Message
	switch m.game.Outcome {
	case OutcomePlayerWin, OutcomePlayerBlackjack:
		return winStyle.Render(msg)
	case OutcomeDealerWin:
		return lossStyle.Render(msg)
	case OutcomePush:
		return pushStyle.Render(msg)
	default:
		return msg
	}
}

func (m Model) renderHelp() string {
	switch m.game.Phase {
	case PhaseBetting:
		return helpStyle.Render("[1] $10  [2] $25  [3] $50  [4] $100  [Q] Quit")
	case PhasePlayerTurn:
		help := "[H] Hit  [S] Stand"
		if m.game.CanDoubleDown() {
			help += "  [D] Double Down"
		}
		return helpStyle.Render(help)
	case PhaseResult:
		return helpStyle.Render("[Enter/N] New Round  [Q] Quit")
	default:
		return ""
	}
}

// Done returns true when the player wants to return to the menu.
func (m Model) Done() bool {
	return m.done
}
