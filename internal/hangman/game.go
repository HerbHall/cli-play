package hangman

import (
	"errors"
	"math/rand/v2"
	"sort"
	"strings"
	"unicode"
)

const maxWrongGuesses = 6

// wordList contains common 5-8 letter words for the game.
var wordList = []string{
	"apple", "beach", "brave", "bring", "candy",
	"chair", "chase", "clean", "cloud", "dance",
	"dream", "earth", "flame", "globe", "grape",
	"happy", "heart", "house", "judge", "knife",
	"laugh", "lemon", "light", "magic", "music",
	"night", "ocean", "paint", "plane", "plant",
	"queen", "quick", "river", "robot", "round",
	"seven", "shell", "shine", "snake", "space",
	"stone", "storm", "sugar", "table", "tiger",
	"train", "water", "whale", "world", "youth",
}

// Game holds the complete state of a Hangman game.
type Game struct {
	target  string
	guessed map[rune]bool
	wrong   int
}

// NewGame creates a new game with a randomly selected word.
func NewGame() *Game {
	word := wordList[rand.IntN(len(wordList))]
	return NewGameWithWord(word)
}

// NewGameWithWord creates a new game with a specific target word (for testing).
func NewGameWithWord(word string) *Game {
	return &Game{
		target:  strings.ToUpper(word),
		guessed: make(map[rune]bool),
	}
}

// Guess processes a letter guess. Returns an error if the letter was already
// guessed or if the input is not a letter.
func (g *Game) Guess(letter rune) error {
	letter = unicode.ToUpper(letter)

	if !unicode.IsLetter(letter) {
		return errors.New("input must be a letter")
	}
	if g.IsOver() {
		return errors.New("game is already over")
	}
	if g.guessed[letter] {
		return errors.New("letter already guessed")
	}

	g.guessed[letter] = true

	if !strings.ContainsRune(g.target, letter) {
		g.wrong++
	}

	return nil
}

// IsOver returns true if the game has ended (won or lost).
func (g *Game) IsOver() bool {
	return g.Won() || g.wrong >= maxWrongGuesses
}

// Won returns true if all letters in the target word have been guessed.
func (g *Game) Won() bool {
	for _, ch := range g.target {
		if !g.guessed[ch] {
			return false
		}
	}
	return true
}

// WrongGuesses returns the number of incorrect guesses made.
func (g *Game) WrongGuesses() int {
	return g.wrong
}

// MaxWrong returns the maximum number of wrong guesses allowed.
func (g *Game) MaxWrong() int {
	return maxWrongGuesses
}

// RevealedWord returns the word with guessed letters shown and unguessed
// letters replaced by underscores, separated by spaces.
func (g *Game) RevealedWord() string {
	var b strings.Builder
	for i, ch := range g.target {
		if i > 0 {
			b.WriteRune(' ')
		}
		if g.guessed[ch] {
			b.WriteRune(ch)
		} else {
			b.WriteRune('_')
		}
	}
	return b.String()
}

// UsedLetters returns a sorted slice of all guessed letters.
func (g *Game) UsedLetters() []rune {
	letters := make([]rune, 0, len(g.guessed))
	for ch := range g.guessed {
		letters = append(letters, ch)
	}
	sort.Slice(letters, func(i, j int) bool {
		return letters[i] < letters[j]
	})
	return letters
}

// Target returns the target word (for display after game ends).
func (g *Game) Target() string {
	return g.target
}
