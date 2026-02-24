package wordle

import (
	"errors"
	"math/rand/v2"
	"strings"
	"unicode"
)

// LetterResult indicates how a guessed letter relates to the target word.
type LetterResult int

const (
	Untested LetterResult = iota
	Absent
	Present
	Correct
)

// GuessResult holds the result for each letter in a 5-letter guess.
type GuessResult [5]LetterResult

// Game holds the complete state of a Wordle game.
type Game struct {
	Target        string
	Guesses       []string
	Results       []GuessResult
	MaxGuesses    int
	Won           bool
	Over          bool
	KeyboardState map[rune]LetterResult
}

// NewGame creates a new game with a randomly selected target word.
func NewGame() *Game {
	word := solutionWords[rand.IntN(len(solutionWords))]
	return NewGameWithWord(word)
}

// NewGameWithWord creates a new game with a specific target word (for testing).
func NewGameWithWord(word string) *Game {
	return &Game{
		Target:        strings.ToUpper(word),
		MaxGuesses:    6,
		KeyboardState: make(map[rune]LetterResult),
	}
}

// Guess validates and scores a guess against the target word.
// Returns the result for each letter position and updates game state.
func (g *Game) Guess(word string) (GuessResult, error) {
	if g.Over {
		return GuessResult{}, errors.New("game is already over")
	}

	word = strings.ToUpper(word)

	if len(word) != 5 {
		return GuessResult{}, errors.New("guess must be exactly 5 letters")
	}
	for _, ch := range word {
		if !unicode.IsLetter(ch) {
			return GuessResult{}, errors.New("guess must contain only letters")
		}
	}

	result := g.evaluate(word)

	g.Guesses = append(g.Guesses, word)
	g.Results = append(g.Results, result)
	g.updateKeyboard(word, result)

	if word == g.Target {
		g.Won = true
		g.Over = true
	} else if len(g.Guesses) >= g.MaxGuesses {
		g.Over = true
	}

	return result, nil
}

// evaluate scores a guess using the two-pass algorithm:
// Pass 1 (green): exact position matches consume target letters.
// Pass 2 (yellow): remaining guess letters match unconsumed target letters.
func (g *Game) evaluate(guess string) GuessResult {
	var result GuessResult
	target := []rune(g.Target)
	guessRunes := []rune(guess)

	consumed := [5]bool{}

	// Pass 1: mark exact matches (Correct/green).
	for i := 0; i < 5; i++ {
		if guessRunes[i] == target[i] {
			result[i] = Correct
			consumed[i] = true
		}
	}

	// Pass 2: mark present (yellow) or absent (gray).
	for i := 0; i < 5; i++ {
		if result[i] == Correct {
			continue
		}

		found := false
		for j := 0; j < 5; j++ {
			if !consumed[j] && guessRunes[i] == target[j] {
				result[i] = Present
				consumed[j] = true
				found = true
				break
			}
		}
		if !found {
			result[i] = Absent
		}
	}

	return result
}

// updateKeyboard updates the on-screen keyboard state.
// A letter's state only upgrades: Correct > Present > Absent.
func (g *Game) updateKeyboard(guess string, result GuessResult) {
	for i, ch := range guess {
		current, exists := g.KeyboardState[ch]
		if !exists || result[i] > current {
			g.KeyboardState[ch] = result[i]
		}
	}
}

// IsOver returns true if the game has ended (won or out of guesses).
func (g *Game) IsOver() bool {
	return g.Over
}

// RemainingGuesses returns how many guesses the player has left.
func (g *Game) RemainingGuesses() int {
	return g.MaxGuesses - len(g.Guesses)
}
