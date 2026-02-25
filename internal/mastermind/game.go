package mastermind

import "math/rand/v2"

const (
	CodeLength = 4
	MaxGuesses = 10
	NumColors  = 6
)

// Feedback holds the result of evaluating a guess against the secret.
type Feedback struct {
	Exact     int // correct color in correct position (black pegs)
	Misplaced int // correct color in wrong position (white pegs)
}

// GuessEntry records a submitted guess and its feedback.
type GuessEntry struct {
	Code     [CodeLength]int
	Feedback Feedback
}

// Game holds the complete state of a Mastermind game.
type Game struct {
	secret  [CodeLength]int
	guesses []GuessEntry
}

// NewGame creates a fresh game with a random secret code.
// Each peg is a value from 1 to NumColors (inclusive). Colors can repeat.
func NewGame() *Game {
	var secret [CodeLength]int
	for i := range secret {
		secret[i] = rand.IntN(NumColors) + 1
	}
	return &Game{secret: secret}
}

// newGameWithSecret creates a game with a specific secret (for testing).
func newGameWithSecret(secret [CodeLength]int) *Game {
	return &Game{secret: secret}
}

// Guess evaluates the given code against the secret and records the attempt.
// Returns the feedback (exact and misplaced counts).
// Returns (-1, -1) if the game is already over.
func (g *Game) Guess(code [CodeLength]int) Feedback {
	if g.IsOver() {
		return Feedback{-1, -1}
	}

	fb := evaluate(g.secret, code)
	g.guesses = append(g.guesses, GuessEntry{Code: code, Feedback: fb})
	return fb
}

// evaluate computes exact and misplaced counts between secret and guess.
// First pass: count exact matches. Second pass: count color matches among
// remaining (non-exact) positions.
func evaluate(secret, guess [CodeLength]int) Feedback {
	var exact, misplaced int
	var secretRemain, guessRemain [CodeLength]bool

	// First pass: exact matches
	for i := range CodeLength {
		if guess[i] == secret[i] {
			exact++
		} else {
			secretRemain[i] = true
			guessRemain[i] = true
		}
	}

	// Second pass: misplaced (right color, wrong position)
	for i := range CodeLength {
		if !guessRemain[i] {
			continue
		}
		for j := range CodeLength {
			if !secretRemain[j] {
				continue
			}
			if guess[i] == secret[j] {
				misplaced++
				secretRemain[j] = false
				break
			}
		}
	}

	return Feedback{Exact: exact, Misplaced: misplaced}
}

// IsOver returns true if the game has ended (won or out of guesses).
func (g *Game) IsOver() bool {
	return g.Won() || len(g.guesses) >= MaxGuesses
}

// Won returns true if the last guess was an exact match.
func (g *Game) Won() bool {
	if len(g.guesses) == 0 {
		return false
	}
	return g.guesses[len(g.guesses)-1].Feedback.Exact == CodeLength
}

// GuessCount returns the number of guesses made so far.
func (g *Game) GuessCount() int {
	return len(g.guesses)
}

// MaxGuessCount returns the maximum number of guesses allowed.
func (g *Game) MaxGuessCount() int {
	return MaxGuesses
}

// Secret reveals the secret code (for game-over display).
func (g *Game) Secret() [CodeLength]int {
	return g.secret
}

// Guesses returns all recorded guess entries.
func (g *Game) Guesses() []GuessEntry {
	return g.guesses
}
