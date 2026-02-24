package yahtzee

import "testing"

// setDice sets dice values and ensures at least one roll has been made.
func setDice(g *Game, values [5]int) {
	g.Dice.Values = values
	if g.RollsLeft == 3 {
		g.RollsLeft = 2
	}
}

func TestScoreUpper(t *testing.T) {
	tests := []struct {
		name     string
		dice     [5]int
		category Category
		want     int
	}{
		{"ones in mixed", [5]int{1, 1, 2, 3, 4}, Ones, 2},
		{"twos in mixed", [5]int{1, 1, 2, 3, 4}, Twos, 2},
		{"threes in mixed", [5]int{1, 1, 2, 3, 4}, Threes, 3},
		{"fours in mixed", [5]int{1, 1, 2, 3, 4}, Fours, 4},
		{"fives in mixed", [5]int{1, 1, 2, 3, 4}, Fives, 0},
		{"sixes in mixed", [5]int{1, 1, 2, 3, 4}, Sixes, 0},
		{"all sixes", [5]int{6, 6, 6, 6, 6}, Sixes, 30},
		{"no match", [5]int{2, 3, 4, 5, 6}, Ones, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ScoreCategory(tt.dice, tt.category)
			if got != tt.want {
				t.Errorf("ScoreCategory(%v, %s) = %d, want %d", tt.dice, tt.category.Name(), got, tt.want)
			}
		})
	}
}

func TestScoreThreeOfAKind(t *testing.T) {
	tests := []struct {
		name string
		dice [5]int
		want int
	}{
		{"three threes", [5]int{3, 3, 3, 1, 2}, 12},
		{"no three of a kind", [5]int{3, 3, 2, 1, 4}, 0},
		{"four of a kind qualifies", [5]int{5, 5, 5, 5, 1}, 21},
		{"yahtzee qualifies", [5]int{2, 2, 2, 2, 2}, 10},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ScoreCategory(tt.dice, ThreeOfAKind)
			if got != tt.want {
				t.Errorf("ScoreCategory(%v, ThreeOfAKind) = %d, want %d", tt.dice, got, tt.want)
			}
		})
	}
}

func TestScoreFourOfAKind(t *testing.T) {
	tests := []struct {
		name string
		dice [5]int
		want int
	}{
		{"four fours", [5]int{4, 4, 4, 4, 2}, 18},
		{"only three", [5]int{4, 4, 4, 3, 2}, 0},
		{"yahtzee qualifies", [5]int{6, 6, 6, 6, 6}, 30},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ScoreCategory(tt.dice, FourOfAKind)
			if got != tt.want {
				t.Errorf("ScoreCategory(%v, FourOfAKind) = %d, want %d", tt.dice, got, tt.want)
			}
		})
	}
}

func TestScoreFullHouse(t *testing.T) {
	tests := []struct {
		name string
		dice [5]int
		want int
	}{
		{"three and two", [5]int{3, 3, 3, 2, 2}, 25},
		{"four of a kind not full house", [5]int{3, 3, 3, 3, 2}, 0},
		{"yahtzee not full house", [5]int{1, 1, 1, 1, 1}, 0},
		{"different full house", [5]int{5, 5, 6, 6, 6}, 25},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ScoreCategory(tt.dice, FullHouse)
			if got != tt.want {
				t.Errorf("ScoreCategory(%v, FullHouse) = %d, want %d", tt.dice, got, tt.want)
			}
		})
	}
}

func TestScoreSmallStraight(t *testing.T) {
	tests := []struct {
		name string
		dice [5]int
		want int
	}{
		{"1-2-3-4 with 6", [5]int{1, 2, 3, 4, 6}, 30},
		{"2-3-4-5 scrambled", [5]int{2, 3, 4, 5, 1}, 30},
		{"3-4-5-6 with dup", [5]int{3, 4, 5, 6, 3}, 30},
		{"gap in sequence", [5]int{1, 2, 3, 5, 6}, 0},
		{"all same", [5]int{4, 4, 4, 4, 4}, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ScoreCategory(tt.dice, SmallStraight)
			if got != tt.want {
				t.Errorf("ScoreCategory(%v, SmallStraight) = %d, want %d", tt.dice, got, tt.want)
			}
		})
	}
}

func TestScoreLargeStraight(t *testing.T) {
	tests := []struct {
		name string
		dice [5]int
		want int
	}{
		{"1-2-3-4-5", [5]int{1, 2, 3, 4, 5}, 40},
		{"2-3-4-5-6", [5]int{2, 3, 4, 5, 6}, 40},
		{"only small straight", [5]int{1, 2, 3, 4, 6}, 0},
		{"scrambled large", [5]int{5, 3, 1, 4, 2}, 40},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ScoreCategory(tt.dice, LargeStraight)
			if got != tt.want {
				t.Errorf("ScoreCategory(%v, LargeStraight) = %d, want %d", tt.dice, got, tt.want)
			}
		})
	}
}

func TestScoreYahtzee(t *testing.T) {
	tests := []struct {
		name string
		dice [5]int
		want int
	}{
		{"all fives", [5]int{5, 5, 5, 5, 5}, 50},
		{"not yahtzee", [5]int{5, 5, 5, 5, 4}, 0},
		{"all ones", [5]int{1, 1, 1, 1, 1}, 50},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ScoreCategory(tt.dice, YahtzeeScore)
			if got != tt.want {
				t.Errorf("ScoreCategory(%v, Yahtzee) = %d, want %d", tt.dice, got, tt.want)
			}
		})
	}
}

func TestScoreChance(t *testing.T) {
	tests := []struct {
		name string
		dice [5]int
		want int
	}{
		{"sequential", [5]int{1, 2, 3, 4, 5}, 15},
		{"all sixes", [5]int{6, 6, 6, 6, 6}, 30},
		{"mixed", [5]int{1, 1, 1, 1, 1}, 5},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ScoreCategory(tt.dice, Chance)
			if got != tt.want {
				t.Errorf("ScoreCategory(%v, Chance) = %d, want %d", tt.dice, got, tt.want)
			}
		})
	}
}

func TestUpperBonus(t *testing.T) {
	tests := []struct {
		name  string
		total int
		want  int
	}{
		{"exactly 63", 63, 35},
		{"above 63", 80, 35},
		{"below 63", 62, 0},
		{"zero", 0, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sc := &Scorecard{}
			// Distribute the total across upper categories to reach desired amount.
			// Use Sixes to hold the bulk.
			if tt.total > 0 {
				sc.Used[Sixes] = true
				sc.Scores[Sixes] = tt.total
			}
			got := sc.UpperBonus()
			if got != tt.want {
				t.Errorf("UpperBonus() with total %d = %d, want %d", tt.total, got, tt.want)
			}
		})
	}
}

func TestYahtzeeBonus(t *testing.T) {
	t.Run("bonus after scoring yahtzee 50", func(t *testing.T) {
		g := NewGame()
		// Score a Yahtzee first
		setDice(g, [5]int{5, 5, 5, 5, 5})
		if err := g.Score(YahtzeeScore); err != nil {
			t.Fatalf("scoring yahtzee: %v", err)
		}
		if g.Scorecard.Scores[YahtzeeScore] != 50 {
			t.Fatalf("expected yahtzee score 50, got %d", g.Scorecard.Scores[YahtzeeScore])
		}

		// Roll another Yahtzee and score in chance
		setDice(g, [5]int{3, 3, 3, 3, 3})
		if err := g.Score(Chance); err != nil {
			t.Fatalf("scoring chance: %v", err)
		}
		if g.YahtzeeBonuses != 1 {
			t.Errorf("expected 1 yahtzee bonus, got %d", g.YahtzeeBonuses)
		}
	})

	t.Run("no bonus when yahtzee scored as zero", func(t *testing.T) {
		g := NewGame()
		// Score a zero in Yahtzee (not a yahtzee roll)
		setDice(g, [5]int{1, 2, 3, 4, 5})
		if err := g.Score(YahtzeeScore); err != nil {
			t.Fatalf("scoring yahtzee as zero: %v", err)
		}
		if g.Scorecard.Scores[YahtzeeScore] != 0 {
			t.Fatalf("expected yahtzee score 0, got %d", g.Scorecard.Scores[YahtzeeScore])
		}

		// Roll a real Yahtzee — no bonus because original was 0
		setDice(g, [5]int{4, 4, 4, 4, 4})
		if err := g.Score(Chance); err != nil {
			t.Fatalf("scoring chance: %v", err)
		}
		if g.YahtzeeBonuses != 0 {
			t.Errorf("expected 0 yahtzee bonuses, got %d", g.YahtzeeBonuses)
		}
	})
}

func TestTurnFlow(t *testing.T) {
	g := NewGame()

	// Initial state
	if g.RollsLeft != 3 {
		t.Fatalf("expected 3 rolls, got %d", g.RollsLeft)
	}
	if g.Turn != 1 {
		t.Fatalf("expected turn 1, got %d", g.Turn)
	}

	// Cannot hold before rolling
	if g.CanHold() {
		t.Error("should not be able to hold before rolling")
	}

	// Roll decrements
	g.Roll()
	if g.RollsLeft != 2 {
		t.Errorf("expected 2 rolls left, got %d", g.RollsLeft)
	}

	// Can hold after rolling
	if !g.CanHold() {
		t.Error("should be able to hold after rolling")
	}

	// Toggle hold
	g.ToggleHold(0)
	if !g.Dice.Held[0] {
		t.Error("die 0 should be held")
	}
	g.ToggleHold(0)
	if g.Dice.Held[0] {
		t.Error("die 0 should be released")
	}

	// Roll again
	g.Roll()
	if g.RollsLeft != 1 {
		t.Errorf("expected 1 roll left, got %d", g.RollsLeft)
	}

	// Last roll
	g.Roll()
	if g.RollsLeft != 0 {
		t.Errorf("expected 0 rolls left, got %d", g.RollsLeft)
	}

	// Cannot roll anymore
	if g.CanRoll() {
		t.Error("should not be able to roll with 0 rolls left")
	}

	// Score advances turn
	if err := g.Score(Ones); err != nil {
		t.Fatalf("scoring: %v", err)
	}
	if g.Turn != 2 {
		t.Errorf("expected turn 2, got %d", g.Turn)
	}
	if g.RollsLeft != 3 {
		t.Errorf("expected 3 rolls for new turn, got %d", g.RollsLeft)
	}
}

func TestGameOver(t *testing.T) {
	g := NewGame()
	for cat := Category(0); cat < NumCategories; cat++ {
		setDice(g, [5]int{1, 2, 3, 4, 5})
		if err := g.Score(cat); err != nil {
			t.Fatalf("scoring category %s: %v", cat.Name(), err)
		}
	}
	if !g.GameOver {
		t.Error("game should be over after 13 categories")
	}
	if g.Turn != 14 {
		t.Errorf("expected turn 14, got %d", g.Turn)
	}
}

func TestGrandTotal(t *testing.T) {
	g := NewGame()

	// Manually set upper section to exactly 63 (3*1 + 3*2 + 3*3 + 3*4 + 3*5 + 3*6 = 63)
	g.Scorecard.Used[Ones] = true
	g.Scorecard.Scores[Ones] = 3
	g.Scorecard.Used[Twos] = true
	g.Scorecard.Scores[Twos] = 6
	g.Scorecard.Used[Threes] = true
	g.Scorecard.Scores[Threes] = 9
	g.Scorecard.Used[Fours] = true
	g.Scorecard.Scores[Fours] = 12
	g.Scorecard.Used[Fives] = true
	g.Scorecard.Scores[Fives] = 15
	g.Scorecard.Used[Sixes] = true
	g.Scorecard.Scores[Sixes] = 18

	// Lower section
	g.Scorecard.Used[FullHouse] = true
	g.Scorecard.Scores[FullHouse] = 25
	g.Scorecard.Used[YahtzeeScore] = true
	g.Scorecard.Scores[YahtzeeScore] = 50

	// Add a Yahtzee bonus
	g.YahtzeeBonuses = 1

	// Upper: 63, Bonus: 35, Lower: 25+50=75, YahtzeeBonus: 100
	expected := 63 + 35 + 75 + 100
	got := g.GrandTotal()
	if got != expected {
		t.Errorf("GrandTotal() = %d, want %d", got, expected)
	}
}

func TestScoreErrors(t *testing.T) {
	t.Run("cannot score before rolling", func(t *testing.T) {
		g := NewGame()
		err := g.Score(Ones)
		if err == nil {
			t.Error("expected error when scoring before rolling")
		}
	})

	t.Run("cannot score same category twice", func(t *testing.T) {
		g := NewGame()
		setDice(g, [5]int{1, 1, 1, 1, 1})
		if err := g.Score(Ones); err != nil {
			t.Fatalf("first score: %v", err)
		}
		setDice(g, [5]int{1, 1, 1, 1, 1})
		err := g.Score(Ones)
		if err == nil {
			t.Error("expected error when scoring same category twice")
		}
	})
}

func TestCategoryName(t *testing.T) {
	tests := []struct {
		cat  Category
		want string
	}{
		{Ones, "Ones"},
		{YahtzeeScore, "Yahtzee"},
		{Chance, "Chance"},
		{FullHouse, "Full House"},
	}
	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			got := tt.cat.Name()
			if got != tt.want {
				t.Errorf("Category(%d).Name() = %q, want %q", tt.cat, got, tt.want)
			}
		})
	}
}

func TestIsUpper(t *testing.T) {
	for cat := Ones; cat <= Sixes; cat++ {
		if !cat.IsUpper() {
			t.Errorf("%s should be upper", cat.Name())
		}
	}
	for cat := ThreeOfAKind; cat <= Chance; cat++ {
		if cat.IsUpper() {
			t.Errorf("%s should not be upper", cat.Name())
		}
	}
}
