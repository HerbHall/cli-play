package yahtzee

import (
	"errors"
	"math/rand/v2"
	"sort"
)

// Category represents a Yahtzee scoring category.
type Category int

const (
	Ones Category = iota
	Twos
	Threes
	Fours
	Fives
	Sixes
	ThreeOfAKind
	FourOfAKind
	FullHouse
	SmallStraight
	LargeStraight
	YahtzeeScore
	Chance
	NumCategories // sentinel = 13
)

var categoryNames = [NumCategories]string{
	"Ones", "Twos", "Threes", "Fours", "Fives", "Sixes",
	"Three of a Kind", "Four of a Kind", "Full House",
	"Small Straight", "Large Straight", "Yahtzee", "Chance",
}

// Name returns the display name of the category.
func (c Category) Name() string {
	if c >= 0 && c < NumCategories {
		return categoryNames[c]
	}
	return "Unknown"
}

// IsUpper returns true for the upper section (Ones through Sixes).
func (c Category) IsUpper() bool {
	return c >= Ones && c <= Sixes
}

// Dice holds the values and hold state of 5 dice.
type Dice struct {
	Values [5]int
	Held   [5]bool
}

// Scorecard tracks which categories have been scored and their values.
type Scorecard struct {
	Scores [NumCategories]int
	Used   [NumCategories]bool
}

// Game holds the complete state of a Yahtzee game.
type Game struct {
	Dice           Dice
	Scorecard      Scorecard
	RollsLeft      int
	Turn           int
	YahtzeeBonuses int
	GameOver       bool
}

// NewGame creates a fresh game ready to play.
func NewGame() *Game {
	return &Game{
		RollsLeft: 3,
		Turn:      1,
	}
}

// Roll rolls all non-held dice. On the first roll of a turn, held flags are cleared.
func (g *Game) Roll() {
	if !g.CanRoll() {
		return
	}
	if g.RollsLeft == 3 {
		g.Dice.Held = [5]bool{}
	}
	for i := 0; i < 5; i++ {
		if !g.Dice.Held[i] {
			g.Dice.Values[i] = rand.IntN(6) + 1
		}
	}
	g.RollsLeft--
}

// CanRoll returns true if the player has rolls remaining.
func (g *Game) CanRoll() bool {
	return g.RollsLeft > 0
}

// ToggleHold flips the held state of the die at the given index.
func (g *Game) ToggleHold(index int) {
	if index < 0 || index > 4 || !g.CanHold() {
		return
	}
	g.Dice.Held[index] = !g.Dice.Held[index]
}

// CanHold returns true if at least one roll has been made this turn.
func (g *Game) CanHold() bool {
	return g.RollsLeft < 3
}

// CanScore returns true if at least one roll has been made this turn.
func (g *Game) CanScore() bool {
	return g.RollsLeft < 3
}

// Score assigns the current dice to the given category.
func (g *Game) Score(cat Category) error {
	if cat < 0 || cat >= NumCategories {
		return errors.New("invalid category")
	}
	if g.Scorecard.Used[cat] {
		return errors.New("category already used")
	}
	if !g.CanScore() {
		return errors.New("must roll at least once before scoring")
	}

	// Check Yahtzee bonus: current dice are Yahtzee, YahtzeeScore already used with 50.
	if isYahtzee(g.Dice.Values) && g.Scorecard.Used[YahtzeeScore] && g.Scorecard.Scores[YahtzeeScore] == 50 {
		g.YahtzeeBonuses++
	}

	g.Scorecard.Scores[cat] = ScoreCategory(g.Dice.Values, cat)
	g.Scorecard.Used[cat] = true

	g.Turn++
	g.RollsLeft = 3
	g.Dice.Held = [5]bool{}
	g.Dice.Values = [5]int{}

	if g.Turn > 13 {
		g.GameOver = true
	}
	return nil
}

// ScoreCategory calculates the score for a set of dice in a given category.
func ScoreCategory(dice [5]int, cat Category) int {
	switch cat {
	case Ones:
		return scoreUpper(dice, 1)
	case Twos:
		return scoreUpper(dice, 2)
	case Threes:
		return scoreUpper(dice, 3)
	case Fours:
		return scoreUpper(dice, 4)
	case Fives:
		return scoreUpper(dice, 5)
	case Sixes:
		return scoreUpper(dice, 6)
	case ThreeOfAKind:
		return scoreNOfAKind(dice, 3)
	case FourOfAKind:
		return scoreNOfAKind(dice, 4)
	case FullHouse:
		return scoreFullHouse(dice)
	case SmallStraight:
		return scoreSmallStraight(dice)
	case LargeStraight:
		return scoreLargeStraight(dice)
	case YahtzeeScore:
		return scoreYahtzee(dice)
	case Chance:
		return scoreChance(dice)
	}
	return 0
}

func scoreUpper(dice [5]int, target int) int {
	total := 0
	for _, v := range dice {
		if v == target {
			total += v
		}
	}
	return total
}

func scoreNOfAKind(dice [5]int, n int) int {
	c := counts(dice)
	for v := 1; v <= 6; v++ {
		if c[v] >= n {
			return scoreChance(dice)
		}
	}
	return 0
}

func scoreFullHouse(dice [5]int) int {
	c := counts(dice)
	hasTwo, hasThree := false, false
	for v := 1; v <= 6; v++ {
		if c[v] == 2 {
			hasTwo = true
		}
		if c[v] == 3 {
			hasThree = true
		}
	}
	if hasTwo && hasThree {
		return 25
	}
	return 0
}

func scoreSmallStraight(dice [5]int) int {
	has := [7]bool{}
	for _, v := range dice {
		has[v] = true
	}
	for start := 1; start <= 3; start++ {
		if has[start] && has[start+1] && has[start+2] && has[start+3] {
			return 30
		}
	}
	return 0
}

func scoreLargeStraight(dice [5]int) int {
	s := sortedDice(dice)
	if (s == [5]int{1, 2, 3, 4, 5}) || (s == [5]int{2, 3, 4, 5, 6}) {
		return 40
	}
	return 0
}

func scoreYahtzee(dice [5]int) int {
	if isYahtzee(dice) {
		return 50
	}
	return 0
}

func scoreChance(dice [5]int) int {
	total := 0
	for _, v := range dice {
		total += v
	}
	return total
}

// UpperTotal returns the sum of scored upper section categories.
func (s *Scorecard) UpperTotal() int {
	total := 0
	for cat := Ones; cat <= Sixes; cat++ {
		if s.Used[cat] {
			total += s.Scores[cat]
		}
	}
	return total
}

// UpperBonus returns 35 if the upper total is 63 or more, else 0.
func (s *Scorecard) UpperBonus() int {
	if s.UpperTotal() >= 63 {
		return 35
	}
	return 0
}

// LowerTotal returns the sum of scored lower section categories.
func (s *Scorecard) LowerTotal() int {
	total := 0
	for cat := ThreeOfAKind; cat <= Chance; cat++ {
		if s.Used[cat] {
			total += s.Scores[cat]
		}
	}
	return total
}

// GrandTotal returns the complete game score.
func (g *Game) GrandTotal() int {
	return g.Scorecard.UpperTotal() + g.Scorecard.UpperBonus() +
		g.Scorecard.LowerTotal() + g.YahtzeeBonuses*100
}

func counts(dice [5]int) [7]int {
	var c [7]int
	for _, v := range dice {
		if v >= 1 && v <= 6 {
			c[v]++
		}
	}
	return c
}

func sortedDice(dice [5]int) [5]int {
	s := dice
	sort.Ints(s[:])
	return s
}

func isYahtzee(dice [5]int) bool {
	for i := 1; i < 5; i++ {
		if dice[i] != dice[0] {
			return false
		}
	}
	return dice[0] >= 1
}
