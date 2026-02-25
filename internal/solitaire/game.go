package solitaire

import "math/rand/v2"

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

// IsRed returns true for Hearts and Diamonds.
func (s Suit) IsRed() bool {
	return s == Hearts || s == Diamonds
}

// Rank represents a card rank from Ace (1) to King (13).
type Rank int

const (
	Ace   Rank = 1
	Two   Rank = 2
	Three Rank = 3
	Four  Rank = 4
	Five  Rank = 5
	Six   Rank = 6
	Seven Rank = 7
	Eight Rank = 8
	Nine  Rank = 9
	Ten   Rank = 10
	Jack  Rank = 11
	Queen Rank = 12
	King  Rank = 13
)

// String returns the display string for a rank.
func (r Rank) String() string {
	switch r {
	case Ace:
		return "A"
	case Ten:
		return "10"
	case Jack:
		return "J"
	case Queen:
		return "Q"
	case King:
		return "K"
	default:
		return string(rune('0' + int(r)))
	}
}

// Card is a playing card with a rank, suit, and visibility state.
type Card struct {
	Rank   Rank
	Suit   Suit
	FaceUp bool
}

// Label returns a display string like "A\u2660" or "10\u2665".
func (c Card) Label() string {
	return c.Rank.String() + c.Suit.Symbol()
}

// ShuffleFunc is a function that shuffles a slice of cards in place.
type ShuffleFunc func([]Card)

// Game holds the complete state of a Klondike Solitaire game.
type Game struct {
	Tableau     [7][]Card
	Stock       []Card
	Waste       []Card
	Foundations [4][]Card
	Score       int
	Moves       int
	Won         bool
	shuffle     ShuffleFunc
}

// NewGame creates a new solitaire game, shuffles the deck, and deals
// the initial tableau. If shuffle is nil, a default Fisher-Yates shuffle
// is used.
func NewGame(shuffle ShuffleFunc) *Game {
	g := &Game{shuffle: shuffle}
	g.deal()
	return g
}

// deal creates a 52-card deck, shuffles it, and deals the tableau.
func (g *Game) deal() {
	deck := makeDeck()
	if g.shuffle != nil {
		g.shuffle(deck)
	} else {
		rand.Shuffle(len(deck), func(i, j int) {
			deck[i], deck[j] = deck[j], deck[i]
		})
	}

	pos := 0
	for col := 0; col < 7; col++ {
		g.Tableau[col] = make([]Card, col+1)
		copy(g.Tableau[col], deck[pos:pos+col+1])
		pos += col + 1
		// Only the top card is face-up.
		for i := range g.Tableau[col] {
			g.Tableau[col][i].FaceUp = i == col
		}
	}

	g.Stock = make([]Card, 52-pos)
	copy(g.Stock, deck[pos:])
}

// makeDeck creates a standard 52-card deck (all face-down).
func makeDeck() []Card {
	deck := make([]Card, 0, 52)
	for s := Spades; s <= Clubs; s++ {
		for r := Ace; r <= King; r++ {
			deck = append(deck, Card{Rank: r, Suit: s})
		}
	}
	return deck
}

// DrawStock flips the top card from stock to waste. If stock is empty,
// recycles waste back to stock.
func (g *Game) DrawStock() {
	if len(g.Stock) == 0 {
		if len(g.Waste) == 0 {
			return
		}
		g.recycleWaste()
		return
	}

	card := g.Stock[len(g.Stock)-1]
	g.Stock = g.Stock[:len(g.Stock)-1]
	card.FaceUp = true
	g.Waste = append(g.Waste, card)
	g.Moves++
}

// recycleWaste moves all waste cards back to stock (reversed, face-down).
func (g *Game) recycleWaste() {
	g.Stock = make([]Card, len(g.Waste))
	for i, c := range g.Waste {
		c.FaceUp = false
		g.Stock[len(g.Waste)-1-i] = c
	}
	g.Waste = nil
	g.Moves++
}

// MoveWasteToTableau moves the top waste card to the given tableau column.
// Returns false if the move is invalid.
func (g *Game) MoveWasteToTableau(col int) bool {
	if col < 0 || col > 6 || len(g.Waste) == 0 {
		return false
	}
	card := g.Waste[len(g.Waste)-1]
	if !g.canPlaceOnTableau(card, col) {
		return false
	}
	g.Waste = g.Waste[:len(g.Waste)-1]
	g.Tableau[col] = append(g.Tableau[col], card)
	g.Score += 5
	g.Moves++
	g.autoFlip(col)
	return true
}

// MoveWasteToFoundation moves the top waste card to the appropriate foundation.
// Returns false if the move is invalid.
func (g *Game) MoveWasteToFoundation() bool {
	if len(g.Waste) == 0 {
		return false
	}
	card := g.Waste[len(g.Waste)-1]
	fi := g.findFoundation(card)
	if fi < 0 {
		return false
	}
	g.Waste = g.Waste[:len(g.Waste)-1]
	g.Foundations[fi] = append(g.Foundations[fi], card)
	g.Score += 10
	g.Moves++
	g.checkWin()
	return true
}

// MoveTableauToFoundation moves the top card of a tableau column to its
// foundation pile. Returns false if the move is invalid.
func (g *Game) MoveTableauToFoundation(col int) bool {
	if col < 0 || col > 6 || len(g.Tableau[col]) == 0 {
		return false
	}
	card := g.Tableau[col][len(g.Tableau[col])-1]
	if !card.FaceUp {
		return false
	}
	fi := g.findFoundation(card)
	if fi < 0 {
		return false
	}
	g.Tableau[col] = g.Tableau[col][:len(g.Tableau[col])-1]
	g.Foundations[fi] = append(g.Foundations[fi], card)
	g.Score += 10
	g.Moves++
	g.autoFlip(col)
	g.checkWin()
	return true
}

// MoveTableauToTableau moves a stack of face-up cards from one tableau
// column to another. cardIndex is the index of the first card in the stack
// to move. Returns false if the move is invalid.
func (g *Game) MoveTableauToTableau(fromCol, cardIndex, toCol int) bool {
	if fromCol < 0 || fromCol > 6 || toCol < 0 || toCol > 6 || fromCol == toCol {
		return false
	}
	pile := g.Tableau[fromCol]
	if cardIndex < 0 || cardIndex >= len(pile) {
		return false
	}
	if !pile[cardIndex].FaceUp {
		return false
	}

	movingCard := pile[cardIndex]
	if !g.canPlaceOnTableau(movingCard, toCol) {
		return false
	}

	// Move the stack.
	moving := make([]Card, len(pile)-cardIndex)
	copy(moving, pile[cardIndex:])
	g.Tableau[fromCol] = pile[:cardIndex]
	g.Tableau[toCol] = append(g.Tableau[toCol], moving...)
	g.Moves++
	g.autoFlip(fromCol)
	return true
}

// canPlaceOnTableau checks if a card can be placed on top of a tableau column.
// Empty columns accept only Kings. Otherwise: descending rank, alternating color.
func (g *Game) canPlaceOnTableau(card Card, col int) bool {
	pile := g.Tableau[col]
	if len(pile) == 0 {
		return card.Rank == King
	}
	top := pile[len(pile)-1]
	return top.FaceUp &&
		top.Rank == card.Rank+1 &&
		top.Suit.IsRed() != card.Suit.IsRed()
}

// findFoundation returns the foundation index where the card can be placed,
// or -1 if no valid foundation exists.
func (g *Game) findFoundation(card Card) int {
	for i := range g.Foundations {
		pile := g.Foundations[i]
		if len(pile) == 0 {
			if card.Rank == Ace {
				return i
			}
			continue
		}
		top := pile[len(pile)-1]
		if top.Suit == card.Suit && top.Rank == card.Rank-1 {
			return i
		}
	}
	return -1
}

// autoFlip flips the top card of a tableau column face-up if it is face-down.
func (g *Game) autoFlip(col int) {
	pile := g.Tableau[col]
	if len(pile) > 0 && !pile[len(pile)-1].FaceUp {
		g.Tableau[col][len(pile)-1].FaceUp = true
	}
}

// checkWin sets Won to true if all four foundations have 13 cards.
func (g *Game) checkWin() {
	for i := range g.Foundations {
		if len(g.Foundations[i]) != 13 {
			return
		}
	}
	g.Won = true
}

// FaceUpIndex returns the index of the first face-up card in a tableau column,
// or -1 if no face-up cards exist.
func (g *Game) FaceUpIndex(col int) int {
	for i := range g.Tableau[col] {
		if g.Tableau[col][i].FaceUp {
			return i
		}
	}
	return -1
}

// WasteTop returns the top waste card and true, or a zero Card and false
// if the waste is empty.
func (g *Game) WasteTop() (Card, bool) {
	if len(g.Waste) == 0 {
		return Card{}, false
	}
	return g.Waste[len(g.Waste)-1], true
}

// FoundationTop returns the top card and true for the given foundation pile,
// or a zero Card and false if empty.
func (g *Game) FoundationTop(index int) (Card, bool) {
	if index < 0 || index > 3 || len(g.Foundations[index]) == 0 {
		return Card{}, false
	}
	pile := g.Foundations[index]
	return pile[len(pile)-1], true
}
