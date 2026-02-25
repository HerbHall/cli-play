package solitaire

import "testing"

// noShuffle returns a ShuffleFunc that does not shuffle (identity).
func noShuffle(deck []Card) {}

func TestDeckCreation(t *testing.T) {
	deck := makeDeck()
	if len(deck) != 52 {
		t.Fatalf("deck length = %d, want 52", len(deck))
	}

	seen := make(map[string]bool)
	for _, c := range deck {
		key := c.Label()
		if seen[key] {
			t.Errorf("duplicate card: %s", key)
		}
		seen[key] = true
	}
	if len(seen) != 52 {
		t.Errorf("unique cards = %d, want 52", len(seen))
	}
}

func TestShuffleProducesDifferentOrder(t *testing.T) {
	deck1 := makeDeck()
	deck2 := makeDeck()

	// Use the default shuffle on deck2.
	g := NewGame(nil)
	_ = g // just testing deck creation works

	// Compare unshuffled decks are identical.
	same := true
	for i := range deck1 {
		if deck1[i].Rank != deck2[i].Rank || deck1[i].Suit != deck2[i].Suit {
			same = false
			break
		}
	}
	if !same {
		t.Error("two unshuffled decks should be identical")
	}

	// NewGame with nil shuffle should produce a different arrangement
	// than the canonical order. This test is probabilistic but the chance
	// of a shuffle producing the exact same order is astronomically low.
	g2 := NewGame(nil)
	allMatch := true
	canonical := makeDeck()
	pos := 0
	for col := 0; col < 7; col++ {
		for i := range g2.Tableau[col] {
			card := g2.Tableau[col][i]
			if card.Rank != canonical[pos].Rank || card.Suit != canonical[pos].Suit {
				allMatch = false
			}
			pos++
		}
	}
	for _, c := range g2.Stock {
		if c.Rank != canonical[pos].Rank || c.Suit != canonical[pos].Suit {
			allMatch = false
		}
		pos++
	}
	if allMatch {
		t.Error("shuffled game has same order as canonical deck (extremely unlikely)")
	}
}

func TestInitialDeal(t *testing.T) {
	g := NewGame(noShuffle)

	// Check tableau column sizes.
	for col := 0; col < 7; col++ {
		want := col + 1
		if len(g.Tableau[col]) != want {
			t.Errorf("Tableau[%d] length = %d, want %d", col, len(g.Tableau[col]), want)
		}
	}

	// Check face-up states: only top card of each column is face-up.
	for col := 0; col < 7; col++ {
		pile := g.Tableau[col]
		for i, card := range pile {
			if i == len(pile)-1 {
				if !card.FaceUp {
					t.Errorf("Tableau[%d] top card should be face-up", col)
				}
			} else {
				if card.FaceUp {
					t.Errorf("Tableau[%d][%d] should be face-down", col, i)
				}
			}
		}
	}

	// 28 cards dealt to tableau, 24 in stock.
	if len(g.Stock) != 24 {
		t.Errorf("Stock length = %d, want 24", len(g.Stock))
	}

	// Waste and foundations should be empty.
	if len(g.Waste) != 0 {
		t.Errorf("Waste should be empty, got %d", len(g.Waste))
	}
	for i := range g.Foundations {
		if len(g.Foundations[i]) != 0 {
			t.Errorf("Foundation[%d] should be empty, got %d", i, len(g.Foundations[i]))
		}
	}
}

func TestValidTableauMove(t *testing.T) {
	g := &Game{}

	// Set up: column 0 has a red 7 face-up, column 1 has a black 6 face-up.
	g.Tableau[0] = []Card{{Rank: Seven, Suit: Hearts, FaceUp: true}}
	g.Tableau[1] = []Card{{Rank: Six, Suit: Spades, FaceUp: true}}

	// Black 6 on red 7 -- valid (descending, alternating color).
	if !g.MoveTableauToTableau(1, 0, 0) {
		t.Error("expected valid move: black 6 on red 7")
	}
	if len(g.Tableau[0]) != 2 {
		t.Errorf("Tableau[0] length = %d, want 2", len(g.Tableau[0]))
	}
	if len(g.Tableau[1]) != 0 {
		t.Errorf("Tableau[1] length = %d, want 0", len(g.Tableau[1]))
	}
}

func TestInvalidTableauMove(t *testing.T) {
	tests := []struct {
		name   string
		target Card
		moving Card
		wantOK bool
	}{
		{
			"same color red on red",
			Card{Rank: Seven, Suit: Hearts, FaceUp: true},
			Card{Rank: Six, Suit: Diamonds, FaceUp: true},
			false,
		},
		{
			"wrong rank ascending",
			Card{Rank: Five, Suit: Spades, FaceUp: true},
			Card{Rank: Six, Suit: Hearts, FaceUp: true},
			false,
		},
		{
			"same rank",
			Card{Rank: Seven, Suit: Spades, FaceUp: true},
			Card{Rank: Seven, Suit: Hearts, FaceUp: true},
			false,
		},
		{
			"skip rank",
			Card{Rank: Nine, Suit: Spades, FaceUp: true},
			Card{Rank: Seven, Suit: Hearts, FaceUp: true},
			false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := &Game{}
			g.Tableau[0] = []Card{tt.target}
			g.Tableau[1] = []Card{tt.moving}

			if got := g.MoveTableauToTableau(1, 0, 0); got != tt.wantOK {
				t.Errorf("MoveTableauToTableau() = %v, want %v", got, tt.wantOK)
			}
		})
	}
}

func TestFoundationBuild(t *testing.T) {
	g := &Game{}

	// Ace goes to empty foundation.
	g.Tableau[0] = []Card{{Rank: Ace, Suit: Spades, FaceUp: true}}
	if !g.MoveTableauToFoundation(0) {
		t.Fatal("expected Ace to go to foundation")
	}
	if len(g.Foundations[0]) != 1 {
		t.Fatalf("Foundation[0] length = %d, want 1", len(g.Foundations[0]))
	}

	// Two of same suit goes on top of Ace.
	g.Tableau[0] = []Card{{Rank: Two, Suit: Spades, FaceUp: true}}
	if !g.MoveTableauToFoundation(0) {
		t.Fatal("expected 2 of Spades to go on Ace of Spades")
	}
	if len(g.Foundations[0]) != 2 {
		t.Fatalf("Foundation[0] length = %d, want 2", len(g.Foundations[0]))
	}

	// Wrong suit should fail.
	g.Tableau[0] = []Card{{Rank: Three, Suit: Hearts, FaceUp: true}}
	if g.MoveTableauToFoundation(0) {
		t.Error("should not place Hearts on Spades foundation")
	}

	// Wrong rank should fail.
	g.Tableau[0] = []Card{{Rank: Four, Suit: Spades, FaceUp: true}}
	if g.MoveTableauToFoundation(0) {
		t.Error("should not place 4 on foundation with 2")
	}
}

func TestStockDraw(t *testing.T) {
	g := NewGame(noShuffle)

	initialStock := len(g.Stock)
	g.DrawStock()

	if len(g.Stock) != initialStock-1 {
		t.Errorf("Stock length = %d, want %d", len(g.Stock), initialStock-1)
	}
	if len(g.Waste) != 1 {
		t.Errorf("Waste length = %d, want 1", len(g.Waste))
	}
	if !g.Waste[0].FaceUp {
		t.Error("drawn card should be face-up")
	}
	if g.Moves != 1 {
		t.Errorf("Moves = %d, want 1", g.Moves)
	}
}

func TestStockRecycling(t *testing.T) {
	g := NewGame(noShuffle)

	// Draw all stock cards.
	stockSize := len(g.Stock)
	for range stockSize {
		g.DrawStock()
	}

	if len(g.Stock) != 0 {
		t.Fatalf("Stock should be empty, got %d", len(g.Stock))
	}
	if len(g.Waste) != stockSize {
		t.Fatalf("Waste should have %d cards, got %d", stockSize, len(g.Waste))
	}

	// Drawing again should recycle.
	g.DrawStock()

	if len(g.Waste) != 0 {
		t.Errorf("Waste should be empty after recycle, got %d", len(g.Waste))
	}
	if len(g.Stock) != stockSize {
		t.Errorf("Stock should have %d cards after recycle, got %d", stockSize, len(g.Stock))
	}

	// All recycled cards should be face-down.
	for i, c := range g.Stock {
		if c.FaceUp {
			t.Errorf("Stock[%d] should be face-down after recycle", i)
		}
	}
}

func TestAutoFlipFaceDown(t *testing.T) {
	g := &Game{}

	// Column with face-down card under a face-up card.
	g.Tableau[0] = []Card{
		{Rank: King, Suit: Spades, FaceUp: false},
		{Rank: Queen, Suit: Hearts, FaceUp: true},
	}
	g.Tableau[1] = []Card{} // empty column for King

	// Move Queen away -- King should auto-flip.
	// King can't go to empty col from tableau move; move Queen to another col.
	g.Tableau[2] = []Card{{Rank: King, Suit: Clubs, FaceUp: true}}

	if !g.MoveTableauToTableau(0, 1, 2) {
		t.Fatal("expected valid move: Queen of Hearts on King of Clubs")
	}

	if !g.Tableau[0][0].FaceUp {
		t.Error("face-down card should auto-flip after top card removed")
	}
}

func TestWinDetection(t *testing.T) {
	g := &Game{}

	// Fill all four foundations with 13 cards each.
	suits := [4]Suit{Spades, Hearts, Diamonds, Clubs}
	for i, s := range suits {
		g.Foundations[i] = make([]Card, 0, 13)
		for r := Ace; r <= King; r++ {
			g.Foundations[i] = append(g.Foundations[i], Card{Rank: r, Suit: s, FaceUp: true})
		}
	}

	g.checkWin()
	if !g.Won {
		t.Error("expected Win to be true with all foundations complete")
	}
}

func TestWinNotDetectedIncomplete(t *testing.T) {
	g := &Game{}

	// Only 3 complete foundations.
	suits := [3]Suit{Spades, Hearts, Diamonds}
	for i, s := range suits {
		g.Foundations[i] = make([]Card, 0, 13)
		for r := Ace; r <= King; r++ {
			g.Foundations[i] = append(g.Foundations[i], Card{Rank: r, Suit: s, FaceUp: true})
		}
	}
	// Fourth foundation has only 12 cards.
	g.Foundations[3] = make([]Card, 0, 12)
	for r := Ace; r <= Queen; r++ {
		g.Foundations[3] = append(g.Foundations[3], Card{Rank: r, Suit: Clubs, FaceUp: true})
	}

	g.checkWin()
	if g.Won {
		t.Error("should not win with incomplete foundation")
	}
}

func TestMoveCount(t *testing.T) {
	g := NewGame(noShuffle)

	g.DrawStock() // move 1
	g.DrawStock() // move 2

	if g.Moves != 2 {
		t.Errorf("Moves = %d, want 2", g.Moves)
	}
}

func TestKingOnEmptyTableau(t *testing.T) {
	g := &Game{}

	// Empty column 0, King in column 1.
	g.Tableau[0] = []Card{}
	g.Tableau[1] = []Card{{Rank: King, Suit: Hearts, FaceUp: true}}

	if !g.MoveTableauToTableau(1, 0, 0) {
		t.Error("King should be placeable on empty tableau column")
	}

	// Non-King should not go on empty column.
	g.Tableau[0] = []Card{}
	g.Tableau[1] = []Card{{Rank: Queen, Suit: Spades, FaceUp: true}}

	if g.MoveTableauToTableau(1, 0, 0) {
		t.Error("non-King should not be placeable on empty tableau column")
	}
}

func TestStackMove(t *testing.T) {
	g := &Game{}

	// Column 0: King of Clubs (black), Queen of Hearts (red) -- valid stack.
	g.Tableau[0] = []Card{
		{Rank: King, Suit: Clubs, FaceUp: true},
		{Rank: Queen, Suit: Hearts, FaceUp: true},
	}
	// Column 1: empty.
	g.Tableau[1] = []Card{}

	// Move the King+Queen stack to empty column 1.
	if !g.MoveTableauToTableau(0, 0, 1) {
		t.Error("expected valid stack move of King+Queen to empty column")
	}
	if len(g.Tableau[1]) != 2 {
		t.Errorf("Tableau[1] length = %d, want 2", len(g.Tableau[1]))
	}
	if len(g.Tableau[0]) != 0 {
		t.Errorf("Tableau[0] length = %d, want 0", len(g.Tableau[0]))
	}
}

func TestWasteToFoundation(t *testing.T) {
	g := &Game{}
	g.Waste = []Card{{Rank: Ace, Suit: Hearts, FaceUp: true}}

	if !g.MoveWasteToFoundation() {
		t.Error("Ace from waste should go to foundation")
	}
	if len(g.Foundations[0]) != 1 {
		t.Errorf("Foundation length = %d, want 1", len(g.Foundations[0]))
	}
	if g.Score != 10 {
		t.Errorf("Score = %d, want 10", g.Score)
	}
}

func TestWasteToTableau(t *testing.T) {
	g := &Game{}
	g.Waste = []Card{{Rank: Six, Suit: Spades, FaceUp: true}}
	g.Tableau[0] = []Card{{Rank: Seven, Suit: Hearts, FaceUp: true}}

	if !g.MoveWasteToTableau(0) {
		t.Error("black 6 from waste should go on red 7 in tableau")
	}
	if len(g.Tableau[0]) != 2 {
		t.Errorf("Tableau[0] length = %d, want 2", len(g.Tableau[0]))
	}
	if g.Score != 5 {
		t.Errorf("Score = %d, want 5", g.Score)
	}
}

func TestCardLabel(t *testing.T) {
	tests := []struct {
		card Card
		want string
	}{
		{Card{Rank: Ace, Suit: Spades}, "A\u2660"},
		{Card{Rank: Ten, Suit: Hearts}, "10\u2665"},
		{Card{Rank: King, Suit: Diamonds}, "K\u2666"},
		{Card{Rank: Two, Suit: Clubs}, "2\u2663"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			if got := tt.card.Label(); got != tt.want {
				t.Errorf("Label() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestSuitIsRed(t *testing.T) {
	tests := []struct {
		suit Suit
		want bool
	}{
		{Spades, false},
		{Hearts, true},
		{Diamonds, true},
		{Clubs, false},
	}

	for _, tt := range tests {
		if got := tt.suit.IsRed(); got != tt.want {
			t.Errorf("Suit(%d).IsRed() = %v, want %v", tt.suit, got, tt.want)
		}
	}
}

func TestFaceUpIndex(t *testing.T) {
	g := &Game{}
	g.Tableau[0] = []Card{
		{Rank: Ace, Suit: Spades, FaceUp: false},
		{Rank: Two, Suit: Spades, FaceUp: false},
		{Rank: Three, Suit: Spades, FaceUp: true},
	}

	if got := g.FaceUpIndex(0); got != 2 {
		t.Errorf("FaceUpIndex(0) = %d, want 2", got)
	}

	g.Tableau[1] = []Card{}
	if got := g.FaceUpIndex(1); got != -1 {
		t.Errorf("FaceUpIndex(1) = %d, want -1", got)
	}
}

func TestDrawFromEmptyStockAndWaste(t *testing.T) {
	g := &Game{}
	g.Stock = nil
	g.Waste = nil

	// Should not panic or change state.
	g.DrawStock()
	if g.Moves != 0 {
		t.Errorf("Moves = %d, want 0 (no-op draw)", g.Moves)
	}
}

func TestCannotMoveFaceDownCard(t *testing.T) {
	g := &Game{}
	g.Tableau[0] = []Card{{Rank: King, Suit: Spades, FaceUp: false}}
	g.Tableau[1] = []Card{}

	if g.MoveTableauToTableau(0, 0, 1) {
		t.Error("should not move face-down card")
	}
}

func TestCannotMoveToFoundationFaceDown(t *testing.T) {
	g := &Game{}
	g.Tableau[0] = []Card{{Rank: Ace, Suit: Spades, FaceUp: false}}

	if g.MoveTableauToFoundation(0) {
		t.Error("should not move face-down card to foundation")
	}
}
