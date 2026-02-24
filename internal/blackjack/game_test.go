package blackjack

import "testing"

// fixedDeck returns a ShuffleFunc that places the given cards at the front
// of the deck, leaving remaining positions untouched. This provides
// deterministic draws for testing.
func fixedDeck(cards ...Card) ShuffleFunc {
	return func(deck []Card) {
		copy(deck, cards)
	}
}

// card is a shorthand constructor for tests.
func card(r Rank, s Suit) Card {
	return Card{Rank: r, Suit: s}
}

func TestHandScore(t *testing.T) {
	tests := []struct {
		name  string
		cards []Card
		want  int
	}{
		{"5+7=12", []Card{card(Five, Spades), card(Seven, Hearts)}, 12},
		{"K+Q=20", []Card{card(King, Spades), card(Queen, Hearts)}, 20},
		{"A+9=20", []Card{card(Ace, Spades), card(Nine, Hearts)}, 20},
		{"A+9+5=15", []Card{card(Ace, Spades), card(Nine, Hearts), card(Five, Diamonds)}, 15},
		{"A+A=12", []Card{card(Ace, Spades), card(Ace, Hearts)}, 12},
		{"A+A+9=21", []Card{card(Ace, Spades), card(Ace, Hearts), card(Nine, Diamonds)}, 21},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := Hand{Cards: tt.cards}
			if got := h.Score(); got != tt.want {
				t.Errorf("Score() = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestBlackjack(t *testing.T) {
	tests := []struct {
		name  string
		cards []Card
		want  bool
	}{
		{"A+K is blackjack", []Card{card(Ace, Spades), card(King, Hearts)}, true},
		{"7+7+7 is not blackjack", []Card{card(Seven, Spades), card(Seven, Hearts), card(Seven, Diamonds)}, false},
		{"10+Q is not blackjack", []Card{card(Ten, Spades), card(Queen, Hearts)}, false},
		{"A+5 is not blackjack", []Card{card(Ace, Spades), card(Five, Hearts)}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := Hand{Cards: tt.cards}
			if got := h.IsBlackjack(); got != tt.want {
				t.Errorf("IsBlackjack() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBust(t *testing.T) {
	tests := []struct {
		name  string
		cards []Card
		want  bool
	}{
		{"K+Q+5 busts", []Card{card(King, Spades), card(Queen, Hearts), card(Five, Diamonds)}, true},
		{"A+K+Q=21 no bust", []Card{card(Ace, Spades), card(King, Hearts), card(Queen, Diamonds)}, false},
		{"9+9+4 busts", []Card{card(Nine, Spades), card(Nine, Hearts), card(Four, Diamonds)}, true},
		{"10+10 no bust", []Card{card(Ten, Spades), card(Ten, Hearts)}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := Hand{Cards: tt.cards}
			if got := h.IsBusted(); got != tt.want {
				t.Errorf("IsBusted() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSoftHand(t *testing.T) {
	tests := []struct {
		name  string
		cards []Card
		want  bool
	}{
		{"A+6 is soft", []Card{card(Ace, Spades), card(Six, Hearts)}, true},
		{"A+6+K is hard", []Card{card(Ace, Spades), card(Six, Hearts), card(King, Diamonds)}, false},
		{"K+Q is hard", []Card{card(King, Spades), card(Queen, Hearts)}, false},
		{"A+A is soft", []Card{card(Ace, Spades), card(Ace, Hearts)}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := Hand{Cards: tt.cards}
			if got := h.IsSoft(); got != tt.want {
				t.Errorf("IsSoft() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDealerPlay(t *testing.T) {
	tests := []struct {
		name     string
		cards    []Card
		wantMin  int
		wantBust bool
	}{
		{
			"hits on 16 stands on 17+",
			// Dealer starts with 6+10=16, draws 2 -> 18.
			[]Card{card(Six, Spades), card(Ten, Hearts), card(Two, Diamonds)},
			17,
			false,
		},
		{
			"stands on 17",
			[]Card{card(Seven, Spades), card(Ten, Hearts)},
			17,
			false,
		},
		{
			"dealer busts",
			// 6+10=16, draws 10 -> 26.
			[]Card{card(Six, Spades), card(Ten, Hearts), card(Ten, Diamonds)},
			0,
			true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := &Game{
				Deck:       NewDeck(fixedDeck(tt.cards...)),
				DealerHand: Hand{},
			}
			// Deal from the fixed deck.
			for _, c := range tt.cards[:2] {
				g.DealerHand.Add(c)
			}
			// Remaining cards in deck start at position 0; advance past dealt cards.
			g.Deck.pos = 2

			g.PlayDealer()

			score := g.DealerHand.Score()
			if tt.wantBust {
				if !g.DealerHand.IsBusted() {
					t.Errorf("expected dealer to bust, got score %d", score)
				}
			} else {
				if score < tt.wantMin {
					t.Errorf("dealer score %d < minimum %d", score, tt.wantMin)
				}
				if g.DealerHand.IsBusted() {
					t.Errorf("dealer unexpectedly busted with score %d", score)
				}
			}
		})
	}
}

func TestOutcomes(t *testing.T) {
	tests := []struct {
		name string
		// Fixed deck order: player1, dealer1, player2, dealer2, then dealer draw cards.
		deckCards []Card
		action    string // "stand" or "hit"
		want      Outcome
	}{
		{
			"player 20 vs dealer 18",
			// P: 10+10=20, D: 8+10=18.
			[]Card{card(Ten, Spades), card(Eight, Hearts), card(Ten, Diamonds), card(Ten, Clubs)},
			"stand",
			OutcomePlayerWin,
		},
		{
			"player bust",
			// P: 10+6=16, D: 10+7=17. Player hits 10 -> bust.
			[]Card{card(Ten, Spades), card(Ten, Hearts), card(Six, Diamonds), card(Seven, Clubs), card(Ten, Clubs)},
			"hit",
			OutcomeDealerWin,
		},
		{
			"player blackjack vs dealer 21",
			// P: A+K=BJ, D: 7+4=11, then dealer draws...
			// Dealer will draw: 10 -> 21.
			[]Card{card(Ace, Spades), card(Seven, Hearts), card(King, Diamonds), card(Four, Clubs), card(Ten, Clubs)},
			"stand", // won't be reached; BJ resolves immediately
			OutcomePlayerBlackjack,
		},
		{
			"push at 20",
			// P: 10+10=20, D: 10+10=20.
			[]Card{card(Ten, Spades), card(Ten, Hearts), card(Ten, Diamonds), card(Ten, Clubs)},
			"stand",
			OutcomePush,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewGame(1000, fixedDeck(tt.deckCards...))
			err := g.PlaceBet(50)
			if err != nil {
				t.Fatalf("PlaceBet: %v", err)
			}

			if g.Phase == PhaseResult {
				// Blackjack was resolved immediately.
				if g.Outcome != tt.want {
					t.Errorf("Outcome = %d, want %d", g.Outcome, tt.want)
				}
				return
			}

			switch tt.action {
			case "hit":
				g.Hit()
			case "stand":
				g.Stand()
			}

			if g.Outcome != tt.want {
				t.Errorf("Outcome = %d, want %d", g.Outcome, tt.want)
			}
		})
	}
}

func TestPayout(t *testing.T) {
	tests := []struct {
		name        string
		deckCards   []Card
		bet         int
		wantBalance int
	}{
		{
			"win 1:1",
			// P: 10+10=20, D: 8+9=17.
			[]Card{card(Ten, Spades), card(Eight, Hearts), card(Ten, Diamonds), card(Nine, Clubs)},
			50,
			1050,
		},
		{
			"blackjack 3:2",
			// P: A+K=BJ, D: 7+4=11.
			[]Card{card(Ace, Spades), card(Seven, Hearts), card(King, Diamonds), card(Four, Clubs)},
			100,
			1150, // 1000 - 100 + 100 + 150
		},
		{
			"push returns bet",
			// P: 10+10=20, D: 10+10=20.
			[]Card{card(Ten, Spades), card(Ten, Hearts), card(Ten, Diamonds), card(Ten, Clubs)},
			50,
			1000,
		},
		{
			"loss deducts bet",
			// P: 10+6=16, D: 10+8=18. Player stands, loses.
			[]Card{card(Ten, Spades), card(Ten, Hearts), card(Six, Diamonds), card(Eight, Clubs)},
			50,
			950,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewGame(1000, fixedDeck(tt.deckCards...))
			err := g.PlaceBet(tt.bet)
			if err != nil {
				t.Fatalf("PlaceBet: %v", err)
			}

			// If not already resolved (blackjack), stand.
			if g.Phase == PhasePlayerTurn {
				g.Stand()
			}

			if g.Balance != tt.wantBalance {
				t.Errorf("Balance = %d, want %d", g.Balance, tt.wantBalance)
			}
		})
	}
}

func TestDoubleDown(t *testing.T) {
	tests := []struct {
		name    string
		setup   func() *Game
		wantErr bool
		desc    string
	}{
		{
			"allowed with 2 cards and sufficient balance",
			func() *Game {
				// P: 5+6=11, D: 7+8=15. Double draw: 10 -> 21.
				g := NewGame(1000, fixedDeck(
					card(Five, Spades), card(Seven, Hearts),
					card(Six, Diamonds), card(Eight, Clubs),
					card(Ten, Spades), // double down draw
					card(Ten, Hearts), // dealer draw
				))
				g.PlaceBet(50) //nolint:errcheck
				return g
			},
			false,
			"should succeed",
		},
		{
			"not allowed with 3 cards",
			func() *Game {
				// P: 5+3=8, D: 7+8=15. Hit 2 -> 10. Then try double.
				g := NewGame(1000, fixedDeck(
					card(Five, Spades), card(Seven, Hearts),
					card(Three, Diamonds), card(Eight, Clubs),
					card(Two, Spades), // hit draw
				))
				g.PlaceBet(50) //nolint:errcheck
				g.Hit()
				return g
			},
			true,
			"should fail after hit",
		},
		{
			"not allowed with insufficient balance",
			func() *Game {
				// Balance 80, bet 50 -> remaining 30, can't double to 100.
				g := NewGame(80, fixedDeck(
					card(Five, Spades), card(Seven, Hearts),
					card(Six, Diamonds), card(Eight, Clubs),
				))
				g.PlaceBet(50) //nolint:errcheck
				return g
			},
			true,
			"should fail with low balance",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := tt.setup()
			err := g.DoubleDown()
			if (err != nil) != tt.wantErr {
				t.Errorf("DoubleDown() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestCardString(t *testing.T) {
	tests := []struct {
		card Card
		want string
	}{
		{card(Ace, Spades), "A\u2660"},
		{card(Ten, Hearts), "10\u2665"},
		{card(King, Diamonds), "K\u2666"},
		{card(Two, Clubs), "2\u2663"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			if got := tt.card.String(); got != tt.want {
				t.Errorf("String() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestDeckRemainingAndReshuffle(t *testing.T) {
	d := NewDeck(nil)
	if d.Remaining() != 52 {
		t.Errorf("new deck Remaining() = %d, want 52", d.Remaining())
	}

	for range 52 {
		d.Draw()
	}

	if d.Remaining() != 0 {
		t.Errorf("after 52 draws Remaining() = %d, want 0", d.Remaining())
	}

	// Drawing from exhausted deck should auto-reshuffle.
	c := d.Draw()
	if c.Rank < Ace || c.Rank > King {
		t.Errorf("unexpected card after reshuffle: %v", c)
	}
	if d.Remaining() != 51 {
		t.Errorf("after reshuffle draw Remaining() = %d, want 51", d.Remaining())
	}
}

func TestPlaceBetValidation(t *testing.T) {
	g := NewGame(100, nil)

	if err := g.PlaceBet(0); err == nil {
		t.Error("PlaceBet(0) should fail")
	}

	if err := g.PlaceBet(-10); err == nil {
		t.Error("PlaceBet(-10) should fail")
	}

	if err := g.PlaceBet(200); err == nil {
		t.Error("PlaceBet(200) should fail with balance 100")
	}
}
