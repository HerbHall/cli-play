package typingtest

import (
	"testing"
	"time"
)

func TestGenerateWords(t *testing.T) {
	words := generateWords(25)
	if len(words) != 25 {
		t.Errorf("generateWords(25) returned %d words, want 25", len(words))
	}
	for i, w := range words {
		if w == "" {
			t.Errorf("word[%d] is empty", i)
		}
	}
}

func TestGenerateWordsDifferentCounts(t *testing.T) {
	tests := []struct {
		name  string
		count int
	}{
		{"ten words", 10},
		{"one word", 1},
		{"fifty words", 50},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			words := generateWords(tt.count)
			if len(words) != tt.count {
				t.Errorf("generateWords(%d) returned %d words", tt.count, len(words))
			}
		})
	}
}

func TestNewGameInitialState(t *testing.T) {
	g := NewGameWithWords([]string{"hello", "world"})
	if g.State != Ready {
		t.Errorf("initial state = %d, want Ready(%d)", g.State, Ready)
	}
	if g.CurrentWord != 0 {
		t.Errorf("CurrentWord = %d, want 0", g.CurrentWord)
	}
	if g.CharPos != 0 {
		t.Errorf("CharPos = %d, want 0", g.CharPos)
	}
	if g.CorrectChars != 0 {
		t.Errorf("CorrectChars = %d, want 0", g.CorrectChars)
	}
	if g.TotalTyped != 0 {
		t.Errorf("TotalTyped = %d, want 0", g.TotalTyped)
	}
	if g.Errors != 0 {
		t.Errorf("Errors = %d, want 0", g.Errors)
	}
}

func TestTypeCharCorrect(t *testing.T) {
	g := NewGameWithWords([]string{"hello", "world"})
	g.TypeChar('h')
	if g.State != Playing {
		t.Errorf("state after first char = %d, want Playing(%d)", g.State, Playing)
	}
	if g.CorrectChars != 1 {
		t.Errorf("CorrectChars = %d, want 1", g.CorrectChars)
	}
	if g.TotalTyped != 1 {
		t.Errorf("TotalTyped = %d, want 1", g.TotalTyped)
	}
	if g.Errors != 0 {
		t.Errorf("Errors = %d, want 0", g.Errors)
	}
	if g.CharPos != 1 {
		t.Errorf("CharPos = %d, want 1", g.CharPos)
	}
}

func TestTypeCharIncorrect(t *testing.T) {
	g := NewGameWithWords([]string{"hello", "world"})
	g.TypeChar('x')
	if g.CorrectChars != 0 {
		t.Errorf("CorrectChars = %d, want 0", g.CorrectChars)
	}
	if g.TotalTyped != 1 {
		t.Errorf("TotalTyped = %d, want 1", g.TotalTyped)
	}
	if g.Errors != 1 {
		t.Errorf("Errors = %d, want 1", g.Errors)
	}
	if !g.CurrentErrors[0] {
		t.Error("CurrentErrors[0] should be true")
	}
}

func TestBackspace(t *testing.T) {
	g := NewGameWithWords([]string{"hello", "world"})
	g.TypeChar('h')
	g.TypeChar('x') // error
	if g.Errors != 1 {
		t.Fatalf("Errors before backspace = %d, want 1", g.Errors)
	}
	if g.CharPos != 2 {
		t.Fatalf("CharPos before backspace = %d, want 2", g.CharPos)
	}

	g.Backspace()
	if g.CharPos != 1 {
		t.Errorf("CharPos after backspace = %d, want 1", g.CharPos)
	}
	if g.Errors != 0 {
		t.Errorf("Errors after backspace = %d, want 0", g.Errors)
	}
	if g.TotalTyped != 1 {
		t.Errorf("TotalTyped after backspace = %d, want 1", g.TotalTyped)
	}
}

func TestBackspaceCorrectChar(t *testing.T) {
	g := NewGameWithWords([]string{"hello", "world"})
	g.TypeChar('h')
	if g.CorrectChars != 1 {
		t.Fatalf("CorrectChars = %d, want 1", g.CorrectChars)
	}

	g.Backspace()
	if g.CorrectChars != 0 {
		t.Errorf("CorrectChars after backspace = %d, want 0", g.CorrectChars)
	}
}

func TestBackspaceAtStart(t *testing.T) {
	g := NewGameWithWords([]string{"hello", "world"})
	g.TypeChar('h') // transition to Playing
	g.Backspace()   // back to pos 0
	g.Backspace()   // should be a no-op
	if g.CharPos != 0 {
		t.Errorf("CharPos = %d, want 0 (cannot go negative)", g.CharPos)
	}
}

func TestAdvanceWord(t *testing.T) {
	g := NewGameWithWords([]string{"hi", "world"})
	g.TypeChar('h')
	g.TypeChar('i')
	finished := g.AdvanceWord()
	if finished {
		t.Error("should not be finished after first word")
	}
	if g.CurrentWord != 1 {
		t.Errorf("CurrentWord = %d, want 1", g.CurrentWord)
	}
	if g.CharPos != 0 {
		t.Errorf("CharPos = %d, want 0", g.CharPos)
	}
	// Space counts as correct char.
	if g.CorrectChars != 3 {
		t.Errorf("CorrectChars = %d, want 3 (h, i, space)", g.CorrectChars)
	}
}

func TestGameCompletionAllWords(t *testing.T) {
	g := NewGameWithWords([]string{"ab", "cd"})

	// Type first word.
	g.TypeChar('a')
	g.TypeChar('b')
	g.AdvanceWord()

	// Type second word.
	g.TypeChar('c')
	g.TypeChar('d')
	finished := g.AdvanceWord()

	if !finished {
		t.Error("game should be finished after all words")
	}
	if g.State != Finished {
		t.Errorf("state = %d, want Finished(%d)", g.State, Finished)
	}
}

func TestTimerExpiry(t *testing.T) {
	g := NewGameWithWords([]string{"hello", "world", "testing"})
	g.TimeLimit = 1 // 1 second for testing
	g.TypeChar('h') // starts the game

	// Simulate time passing by setting start time in the past.
	g.StartTime = time.Now().Add(-2 * time.Second)

	expired := g.Tick()
	if !expired {
		t.Error("Tick should return true when time expired")
	}
	if g.State != Finished {
		t.Errorf("state = %d, want Finished(%d)", g.State, Finished)
	}
}

func TestTickNotExpired(t *testing.T) {
	g := NewGameWithWords([]string{"hello", "world"})
	g.TypeChar('h') // starts the game

	expired := g.Tick()
	if expired {
		t.Error("Tick should return false when time has not expired")
	}
	if g.State != Playing {
		t.Errorf("state = %d, want Playing(%d)", g.State, Playing)
	}
}

func TestWPMCalculation(t *testing.T) {
	g := NewGameWithWords([]string{"hello", "world"})
	g.State = Playing
	g.StartTime = time.Now().Add(-60 * time.Second)
	g.Elapsed = 60 * time.Second
	g.CorrectChars = 50 // 50 correct chars = 10 words at 5 chars/word

	wpm := g.WPM()
	if wpm != 10 {
		t.Errorf("WPM = %d, want 10", wpm)
	}
}

func TestWPMZeroElapsed(t *testing.T) {
	g := NewGameWithWords([]string{"hello"})
	g.CorrectChars = 25

	wpm := g.WPM()
	if wpm != 0 {
		t.Errorf("WPM with zero elapsed = %d, want 0", wpm)
	}
}

func TestAccuracyCalculation(t *testing.T) {
	g := NewGameWithWords([]string{"hello"})
	g.TotalTyped = 10
	g.CorrectChars = 8

	accuracy := g.Accuracy()
	want := 80.0
	if accuracy != want {
		t.Errorf("Accuracy = %.1f, want %.1f", accuracy, want)
	}
}

func TestAccuracyPerfect(t *testing.T) {
	g := NewGameWithWords([]string{"hello"})
	g.TotalTyped = 5
	g.CorrectChars = 5

	accuracy := g.Accuracy()
	if accuracy != 100.0 {
		t.Errorf("Accuracy = %.1f, want 100.0", accuracy)
	}
}

func TestAccuracyZeroTyped(t *testing.T) {
	g := NewGameWithWords([]string{"hello"})

	accuracy := g.Accuracy()
	if accuracy != 0 {
		t.Errorf("Accuracy with zero typed = %.1f, want 0", accuracy)
	}
}

func TestTimeRemaining(t *testing.T) {
	g := NewGameWithWords([]string{"hello"})
	g.TimeLimit = 60
	g.Elapsed = 25 * time.Second

	remaining := g.TimeRemaining()
	if remaining != 35 {
		t.Errorf("TimeRemaining = %d, want 35", remaining)
	}
}

func TestTimeRemainingNegative(t *testing.T) {
	g := NewGameWithWords([]string{"hello"})
	g.TimeLimit = 60
	g.Elapsed = 65 * time.Second

	remaining := g.TimeRemaining()
	if remaining != 0 {
		t.Errorf("TimeRemaining = %d, want 0 (should not go negative)", remaining)
	}
}

func TestTypeCharAfterFinished(t *testing.T) {
	g := NewGameWithWords([]string{"ab"})
	g.TypeChar('a')
	g.TypeChar('b')
	g.AdvanceWord() // finishes the game

	before := g.TotalTyped
	g.TypeChar('x') // should be ignored
	if g.TotalTyped != before {
		t.Error("TypeChar should be ignored when game is Finished")
	}
}

func TestBackspaceBeforePlaying(t *testing.T) {
	g := NewGameWithWords([]string{"hello"})
	g.Backspace() // should be a no-op (state is Ready)
	if g.CharPos != 0 {
		t.Errorf("CharPos = %d, want 0", g.CharPos)
	}
}

func TestAdvanceWordBeforePlaying(t *testing.T) {
	g := NewGameWithWords([]string{"hello", "world"})
	finished := g.AdvanceWord() // should be a no-op (state is Ready)
	if finished {
		t.Error("AdvanceWord should not finish when state is Ready")
	}
	if g.CurrentWord != 0 {
		t.Errorf("CurrentWord = %d, want 0", g.CurrentWord)
	}
}

func TestWordErrorTracking(t *testing.T) {
	g := NewGameWithWords([]string{"hi", "go"})
	g.TypeChar('h')
	g.TypeChar('x') // error at position 1
	g.AdvanceWord()

	if !g.WordErrors[0][1] {
		t.Error("WordErrors[0][1] should be true (error at second char)")
	}
	if g.WordErrors[0][0] {
		t.Error("WordErrors[0][0] should be false (correct first char)")
	}
	if !g.WordCompleted[0] {
		t.Error("WordCompleted[0] should be true")
	}
}

func TestCannotTypeMoreThanWordLength(t *testing.T) {
	g := NewGameWithWords([]string{"hi", "go"})
	g.TypeChar('h')
	g.TypeChar('i')
	g.TypeChar('x') // past the end of "hi", should be ignored
	if g.CharPos != 2 {
		t.Errorf("CharPos = %d, want 2 (should not exceed word length)", g.CharPos)
	}
	if g.TotalTyped != 2 {
		t.Errorf("TotalTyped = %d, want 2", g.TotalTyped)
	}
}

func TestWordListNotEmpty(t *testing.T) {
	if len(wordList) == 0 {
		t.Fatal("wordList should not be empty")
	}
	for i, w := range wordList {
		if w == "" {
			t.Errorf("wordList[%d] is empty", i)
		}
	}
}
