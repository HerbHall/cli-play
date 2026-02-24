package blackjack

import (
	"errors"
	"fmt"
	"math/rand/v2"
)

// Suit represents a card suit.
type Suit int

const (
	Spades Suit = iota
	Hearts
	Diamonds
	Clubs
)

// Symbol returns the Unicode symbol for the suit.
func (s Suit) Symbol() string {
	switch s {
	case Spades:
		return "\u2660"
	case Hearts:
		return "\u2665"
	case Diamonds:
		return "\u2666"
	case Clubs:
		return "\u2663"
	}
	return "?"
}

// Rank represents a card rank from Ace to King.
type Rank int

const (
	Ace Rank = iota + 1
	Two
	Three
	Four
	Five
	Six
	Seven
	Eight
	Nine
	Ten
	Jack
	Queen
	King
)

// String returns the display string for a rank.
func (r Rank) String() string {
	switch r {
	case Ace:
		return "A"
	case Jack:
		return "J"
	case Queen:
		return "Q"
	case King:
		return "K"
	default:
		return fmt.Sprintf("%d", int(r))
	}
}

// Card is a playing card with a rank and suit.
type Card struct {
	Rank Rank
	Suit Suit
}

// Value returns the blackjack point value of the card.
// Ace returns 11 (caller must handle soft reduction).
// Face cards (J, Q, K) return 10.
func (c Card) Value() int {
	switch {
	case c.Rank == Ace:
		return 11
	case c.Rank >= Jack:
		return 10
	default:
		return int(c.Rank)
	}
}

// String returns a display string like "A♠" or "10♥".
func (c Card) String() string {
	return c.Rank.String() + c.Suit.Symbol()
}

// ShuffleFunc is a function that shuffles a slice of cards in place.
type ShuffleFunc func([]Card)

// Deck holds a shuffled collection of 52 cards.
type Deck struct {
	cards   []Card
	pos     int
	shuffle ShuffleFunc
}

// NewDeck creates a standard 52-card deck and shuffles it.
// If shuffle is nil, a default Fisher-Yates shuffle using math/rand/v2 is used.
func NewDeck(shuffle ShuffleFunc) *Deck {
	cards := make([]Card, 0, 52)
	for s := Spades; s <= Clubs; s++ {
		for r := Ace; r <= King; r++ {
			cards = append(cards, Card{Rank: r, Suit: s})
		}
	}

	d := &Deck{cards: cards, shuffle: shuffle}
	d.shuffleDeck()
	return d
}

func (d *Deck) shuffleDeck() {
	if d.shuffle != nil {
		d.shuffle(d.cards)
	} else {
		rand.Shuffle(len(d.cards), func(i, j int) {
			d.cards[i], d.cards[j] = d.cards[j], d.cards[i]
		})
	}
	d.pos = 0
}

// Draw returns the next card from the deck. If the deck is exhausted,
// it reshuffles automatically before drawing.
func (d *Deck) Draw() Card {
	if d.pos >= len(d.cards) {
		d.shuffleDeck()
	}
	c := d.cards[d.pos]
	d.pos++
	return c
}

// Remaining returns how many undrawn cards remain.
func (d *Deck) Remaining() int {
	return len(d.cards) - d.pos
}

// Hand represents a collection of cards held by a player or dealer.
type Hand struct {
	Cards []Card
}

// Add appends a card to the hand.
func (h *Hand) Add(c Card) {
	h.Cards = append(h.Cards, c)
}

// Score returns the best blackjack score for the hand (<=21 if possible).
// Aces are reduced from 11 to 1 as needed to avoid busting.
func (h *Hand) Score() int {
	total := 0
	aces := 0
	for _, c := range h.Cards {
		total += c.Value()
		if c.Rank == Ace {
			aces++
		}
	}
	for total > 21 && aces > 0 {
		total -= 10
		aces--
	}
	return total
}

// IsSoft returns true if the hand contains an ace currently counted as 11.
func (h *Hand) IsSoft() bool {
	total := 0
	aces := 0
	for _, c := range h.Cards {
		total += c.Value()
		if c.Rank == Ace {
			aces++
		}
	}
	reduced := 0
	for total > 21 && reduced < aces {
		total -= 10
		reduced++
	}
	return reduced < aces
}

// IsBusted returns true if the hand's score exceeds 21.
func (h *Hand) IsBusted() bool {
	return h.Score() > 21
}

// IsBlackjack returns true if the hand is a natural blackjack (exactly 2 cards totaling 21).
func (h *Hand) IsBlackjack() bool {
	return len(h.Cards) == 2 && h.Score() == 21
}

// Phase represents the current phase of a blackjack round.
type Phase int

const (
	PhaseBetting    Phase = iota
	PhasePlayerTurn Phase = iota
	PhaseDealerTurn Phase = iota
	PhaseResult     Phase = iota
)

// Outcome represents the result of a completed round.
type Outcome int

const (
	OutcomeNone            Outcome = iota
	OutcomePlayerWin       Outcome = iota
	OutcomeDealerWin       Outcome = iota
	OutcomePlayerBlackjack Outcome = iota
	OutcomePush            Outcome = iota
)

// Stats tracks cumulative win/loss/push counts.
type Stats struct {
	Wins   int
	Losses int
	Pushes int
}

// Game holds the complete state of a blackjack game across rounds.
type Game struct {
	Deck         *Deck
	PlayerHand   Hand
	DealerHand   Hand
	Phase        Phase
	Outcome      Outcome
	Stats        Stats
	Bet          int
	Balance      int
	HoleRevealed bool
	Message      string
}

// NewGame creates a new blackjack game with the given starting balance.
func NewGame(balance int, shuffle ShuffleFunc) *Game {
	return &Game{
		Deck:    NewDeck(shuffle),
		Balance: balance,
		Phase:   PhaseBetting,
	}
}

// NewRound resets the hands and phase for a new round.
func (g *Game) NewRound() {
	g.PlayerHand = Hand{}
	g.DealerHand = Hand{}
	g.Phase = PhaseBetting
	g.Outcome = OutcomeNone
	g.Bet = 0
	g.HoleRevealed = false
	g.Message = ""
}

// PlaceBet validates the bet amount, deducts it from balance, deals cards,
// and checks for immediate blackjacks.
func (g *Game) PlaceBet(amount int) error {
	if amount <= 0 {
		return errors.New("bet must be positive")
	}
	if amount > g.Balance {
		return errors.New("insufficient balance")
	}

	g.Bet = amount
	g.Balance -= amount

	// Deal: player, dealer, player, dealer.
	g.PlayerHand.Add(g.Deck.Draw())
	g.DealerHand.Add(g.Deck.Draw())
	g.PlayerHand.Add(g.Deck.Draw())
	g.DealerHand.Add(g.Deck.Draw())

	playerBJ := g.PlayerHand.IsBlackjack()
	dealerBJ := g.DealerHand.IsBlackjack()

	switch {
	case playerBJ && dealerBJ:
		g.HoleRevealed = true
		g.Phase = PhaseResult
		g.Outcome = OutcomePush
		g.Balance += g.Bet
		g.Stats.Pushes++
		g.Message = "Both have Blackjack! Push."
	case playerBJ:
		g.HoleRevealed = true
		g.Phase = PhaseResult
		g.Outcome = OutcomePlayerBlackjack
		payout := g.Bet + g.Bet*3/2
		g.Balance += payout
		g.Stats.Wins++
		g.Message = fmt.Sprintf("BLACKJACK! +$%d", g.Bet*3/2)
	case dealerBJ:
		g.HoleRevealed = true
		g.Phase = PhaseResult
		g.Outcome = OutcomeDealerWin
		g.Stats.Losses++
		g.Message = "Dealer has Blackjack!"
	default:
		g.Phase = PhasePlayerTurn
	}

	return nil
}

// Hit draws a card for the player. If the player busts, the round ends.
func (g *Game) Hit() {
	if g.Phase != PhasePlayerTurn {
		return
	}

	g.PlayerHand.Add(g.Deck.Draw())

	if g.PlayerHand.IsBusted() {
		g.Phase = PhaseResult
		g.Outcome = OutcomeDealerWin
		g.HoleRevealed = true
		g.Stats.Losses++
		g.Message = fmt.Sprintf("BUST! -%d", g.Bet)
	}
}

// Stand ends the player's turn and triggers the dealer's play and resolution.
func (g *Game) Stand() {
	if g.Phase != PhasePlayerTurn {
		return
	}
	g.Phase = PhaseDealerTurn
	g.HoleRevealed = true
	g.PlayDealer()
	g.Resolve()
}

// CanDoubleDown returns true if double down is allowed:
// exactly 2 cards and sufficient balance to double the bet.
func (g *Game) CanDoubleDown() bool {
	return g.Phase == PhasePlayerTurn &&
		len(g.PlayerHand.Cards) == 2 &&
		g.Balance >= g.Bet
}

// DoubleDown doubles the bet, draws exactly one card, and resolves.
func (g *Game) DoubleDown() error {
	if g.Phase != PhasePlayerTurn {
		return errors.New("not player's turn")
	}
	if len(g.PlayerHand.Cards) != 2 {
		return errors.New("can only double down on initial two cards")
	}
	if g.Balance < g.Bet {
		return errors.New("insufficient balance to double down")
	}

	g.Balance -= g.Bet
	g.Bet *= 2
	g.PlayerHand.Add(g.Deck.Draw())

	if g.PlayerHand.IsBusted() {
		g.Phase = PhaseResult
		g.Outcome = OutcomeDealerWin
		g.HoleRevealed = true
		g.Stats.Losses++
		g.Message = fmt.Sprintf("BUST! -%d", g.Bet)
		return nil
	}

	g.Phase = PhaseDealerTurn
	g.HoleRevealed = true
	g.PlayDealer()
	g.Resolve()
	return nil
}

// PlayDealer draws cards for the dealer until the score is 17 or higher.
func (g *Game) PlayDealer() {
	for g.DealerHand.Score() < 17 {
		g.DealerHand.Add(g.Deck.Draw())
	}
}

// Resolve compares hands, determines the outcome, updates stats and balance,
// and sets the result message.
func (g *Game) Resolve() {
	g.Phase = PhaseResult
	g.HoleRevealed = true

	playerScore := g.PlayerHand.Score()
	dealerScore := g.DealerHand.Score()

	switch {
	case g.DealerHand.IsBusted():
		g.Outcome = OutcomePlayerWin
		g.Balance += g.Bet * 2
		g.Stats.Wins++
		g.Message = fmt.Sprintf("Dealer busts! +$%d", g.Bet)
	case playerScore > dealerScore:
		g.Outcome = OutcomePlayerWin
		g.Balance += g.Bet * 2
		g.Stats.Wins++
		g.Message = fmt.Sprintf("You win! +$%d", g.Bet)
	case playerScore < dealerScore:
		g.Outcome = OutcomeDealerWin
		g.Stats.Losses++
		g.Message = fmt.Sprintf("Dealer wins. -$%d", g.Bet)
	default:
		g.Outcome = OutcomePush
		g.Balance += g.Bet
		g.Stats.Pushes++
		g.Message = "Push!"
	}
}
