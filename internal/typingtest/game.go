package typingtest

import (
	"math/rand/v2"
	"time"
)

// State represents the current phase of the typing test.
type State int

const (
	Ready    State = iota // Waiting for the first keypress
	Playing               // Timer running, accepting input
	Finished              // All words typed or time expired
)

const (
	timeLimit  = 60 // seconds
	targetSize = 25 // words per test
)

// Game holds the complete state of a typing test.
type Game struct {
	Words          []string
	State          State
	CurrentWord    int
	CharPos        int
	CorrectChars   int
	TotalTyped     int
	Errors         int
	StartTime      time.Time
	Elapsed        time.Duration
	TimeLimit      int
	WordErrors     [][]bool // per-character error tracking for each word
	WordCompleted  []bool   // true once a word is submitted via space
	CurrentErrors  []bool   // error state for characters in the current word
}

// NewGame creates a new typing test with random words.
func NewGame() *Game {
	return NewGameWithWords(generateWords(targetSize))
}

// NewGameWithWords creates a game with specific words (for testing).
func NewGameWithWords(words []string) *Game {
	g := &Game{
		Words:         words,
		State:         Ready,
		TimeLimit:     timeLimit,
		WordErrors:    make([][]bool, len(words)),
		WordCompleted: make([]bool, len(words)),
	}
	for i, w := range words {
		g.WordErrors[i] = make([]bool, len(w))
	}
	g.CurrentErrors = make([]bool, len(words[0]))
	return g
}

// TypeChar processes a typed character at the current position.
func (g *Game) TypeChar(ch rune) {
	if g.State == Finished {
		return
	}
	if g.State == Ready {
		g.State = Playing
		g.StartTime = time.Now()
	}
	if g.CurrentWord >= len(g.Words) {
		return
	}

	word := g.Words[g.CurrentWord]
	if g.CharPos >= len(word) {
		return
	}

	g.TotalTyped++
	if byte(ch) == word[g.CharPos] {
		g.CorrectChars++
		g.CurrentErrors[g.CharPos] = false
	} else {
		g.Errors++
		g.CurrentErrors[g.CharPos] = true
	}
	g.CharPos++
}

// Backspace removes the last typed character within the current word.
func (g *Game) Backspace() {
	if g.State != Playing || g.CharPos == 0 {
		return
	}

	g.CharPos--
	// Undo the previous character's stats.
	if g.CurrentErrors[g.CharPos] {
		g.Errors--
	} else {
		g.CorrectChars--
	}
	g.TotalTyped--
	g.CurrentErrors[g.CharPos] = false
}

// AdvanceWord moves to the next word when space is pressed.
// Returns true if the game is now finished (all words typed).
func (g *Game) AdvanceWord() bool {
	if g.State != Playing {
		return false
	}
	if g.CurrentWord >= len(g.Words) {
		return false
	}

	// Copy error state for the completed word.
	copy(g.WordErrors[g.CurrentWord], g.CurrentErrors)
	g.WordCompleted[g.CurrentWord] = true

	// Count space as a correct character (word separator).
	g.TotalTyped++
	g.CorrectChars++

	g.CurrentWord++
	g.CharPos = 0

	if g.CurrentWord >= len(g.Words) {
		g.finish()
		return true
	}

	g.CurrentErrors = make([]bool, len(g.Words[g.CurrentWord]))
	return false
}

// Tick updates the elapsed timer. Returns true if time expired.
func (g *Game) Tick() bool {
	if g.State != Playing {
		return false
	}
	g.Elapsed = time.Since(g.StartTime)
	if int(g.Elapsed.Seconds()) >= g.TimeLimit {
		g.finish()
		return true
	}
	return false
}

// WPM calculates words-per-minute using the standard 5-chars-per-word formula.
func (g *Game) WPM() int {
	elapsed := g.Elapsed.Minutes()
	if elapsed <= 0 {
		return 0
	}
	return int(float64(g.CorrectChars) / 5.0 / elapsed)
}

// Accuracy returns the typing accuracy as a percentage.
func (g *Game) Accuracy() float64 {
	if g.TotalTyped == 0 {
		return 0
	}
	return float64(g.CorrectChars) / float64(g.TotalTyped) * 100
}

// TimeRemaining returns the number of seconds left.
func (g *Game) TimeRemaining() int {
	elapsed := int(g.Elapsed.Seconds())
	remaining := g.TimeLimit - elapsed
	if remaining < 0 {
		return 0
	}
	return remaining
}

func (g *Game) finish() {
	g.State = Finished
	g.Elapsed = time.Since(g.StartTime)
}

// generateWords picks n random words from the word list.
func generateWords(n int) []string {
	result := make([]string, n)
	for i := range result {
		result[i] = wordList[rand.IntN(len(wordList))]
	}
	return result
}

// wordList contains common English words for typing practice.
var wordList = []string{
	"the", "be", "to", "of", "and", "a", "in", "that", "have", "it",
	"for", "not", "on", "with", "he", "as", "you", "do", "at", "this",
	"but", "his", "by", "from", "they", "we", "say", "her", "she", "or",
	"an", "will", "my", "one", "all", "would", "there", "their", "what",
	"so", "up", "out", "if", "about", "who", "get", "which", "go", "me",
	"when", "make", "can", "like", "time", "no", "just", "him", "know",
	"take", "people", "into", "year", "your", "good", "some", "could",
	"them", "see", "other", "than", "then", "now", "look", "only", "come",
	"its", "over", "think", "also", "back", "after", "use", "two", "how",
	"our", "work", "first", "well", "way", "even", "new", "want", "because",
	"any", "these", "give", "day", "most", "find", "here", "thing", "many",
	"right", "still", "place", "every", "where", "much", "should", "long",
	"great", "hand", "high", "small", "large", "next", "early", "young",
	"keep", "last", "never", "start", "city", "run", "while", "turn",
	"help", "home", "side", "been", "off", "play", "move", "live", "point",
	"read", "group", "began", "few", "near", "own", "left", "might", "head",
}
