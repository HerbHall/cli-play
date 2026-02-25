package hangman

import "testing"

func TestNewGame(t *testing.T) {
	g := NewGame()

	if g.target == "" {
		t.Error("expected target word to be set")
	}
	if len(g.target) < 5 || len(g.target) > 8 {
		t.Errorf("expected target length 5-8, got %d (%s)", len(g.target), g.target)
	}
	if len(g.guessed) != 0 {
		t.Errorf("expected no guesses, got %d", len(g.guessed))
	}
	if g.WrongGuesses() != 0 {
		t.Errorf("expected 0 wrong guesses, got %d", g.WrongGuesses())
	}
	if g.IsOver() {
		t.Error("game should not be over at start")
	}
}

func TestCorrectGuess(t *testing.T) {
	g := NewGameWithWord("apple")

	err := g.Guess('a')
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if g.WrongGuesses() != 0 {
		t.Errorf("correct guess should not increment wrong count, got %d", g.WrongGuesses())
	}

	revealed := g.RevealedWord()
	if revealed != "A _ _ _ _" {
		t.Errorf("expected 'A _ _ _ _', got %q", revealed)
	}
}

func TestWrongGuess(t *testing.T) {
	g := NewGameWithWord("apple")

	err := g.Guess('z')
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if g.WrongGuesses() != 1 {
		t.Errorf("expected 1 wrong guess, got %d", g.WrongGuesses())
	}
}

func TestDuplicateGuess(t *testing.T) {
	g := NewGameWithWord("apple")

	if err := g.Guess('a'); err != nil {
		t.Fatalf("first guess error: %v", err)
	}

	err := g.Guess('a')
	if err == nil {
		t.Error("expected error for duplicate guess")
	}
}

func TestWinCondition(t *testing.T) {
	g := NewGameWithWord("cat")

	for _, ch := range "cat" {
		if err := g.Guess(ch); err != nil {
			t.Fatalf("guessing '%c': %v", ch, err)
		}
	}

	if !g.Won() {
		t.Error("expected game to be won")
	}
	if !g.IsOver() {
		t.Error("expected game to be over")
	}
	if g.WrongGuesses() != 0 {
		t.Errorf("expected 0 wrong guesses, got %d", g.WrongGuesses())
	}
}

func TestLoseCondition(t *testing.T) {
	g := NewGameWithWord("apple")

	wrongLetters := []rune{'z', 'x', 'w', 'v', 'u', 't'}
	for _, ch := range wrongLetters {
		if err := g.Guess(ch); err != nil {
			t.Fatalf("guessing '%c': %v", ch, err)
		}
	}

	if g.WrongGuesses() != 6 {
		t.Errorf("expected 6 wrong guesses, got %d", g.WrongGuesses())
	}
	if !g.IsOver() {
		t.Error("expected game to be over")
	}
	if g.Won() {
		t.Error("expected game to not be won")
	}
}

func TestRevealedWord(t *testing.T) {
	tests := []struct {
		name    string
		word    string
		guesses []rune
		want    string
	}{
		{"no guesses", "hello", nil, "_ _ _ _ _"},
		{"one correct", "hello", []rune{'h'}, "H _ _ _ _"},
		{"repeated letter", "hello", []rune{'l'}, "_ _ L L _"},
		{"all guessed", "hello", []rune{'h', 'e', 'l', 'o'}, "H E L L O"},
		{"mixed correct and wrong", "hello", []rune{'h', 'z', 'o'}, "H _ _ _ O"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewGameWithWord(tt.word)
			for _, ch := range tt.guesses {
				if err := g.Guess(ch); err != nil {
					t.Fatalf("guessing '%c': %v", ch, err)
				}
			}
			got := g.RevealedWord()
			if got != tt.want {
				t.Errorf("RevealedWord() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestUsedLetters(t *testing.T) {
	g := NewGameWithWord("apple")

	if err := g.Guess('c'); err != nil {
		t.Fatal(err)
	}
	if err := g.Guess('a'); err != nil {
		t.Fatal(err)
	}
	if err := g.Guess('b'); err != nil {
		t.Fatal(err)
	}

	used := g.UsedLetters()
	if len(used) != 3 {
		t.Fatalf("expected 3 used letters, got %d", len(used))
	}
	expected := []rune{'A', 'B', 'C'}
	for i, want := range expected {
		if used[i] != want {
			t.Errorf("UsedLetters()[%d] = '%c', want '%c'", i, used[i], want)
		}
	}
}

func TestGuessAfterGameOver(t *testing.T) {
	g := NewGameWithWord("hi")

	// Win the game
	for _, ch := range "hi" {
		if err := g.Guess(ch); err != nil {
			t.Fatalf("guessing '%c': %v", ch, err)
		}
	}

	err := g.Guess('z')
	if err == nil {
		t.Error("expected error when guessing after game over")
	}
}

func TestMaxWrong(t *testing.T) {
	g := NewGame()
	if g.MaxWrong() != 6 {
		t.Errorf("MaxWrong() = %d, want 6", g.MaxWrong())
	}
}

func TestCaseInsensitive(t *testing.T) {
	g := NewGameWithWord("apple")

	// Guess uppercase
	if err := g.Guess('A'); err != nil {
		t.Fatalf("uppercase guess error: %v", err)
	}

	revealed := g.RevealedWord()
	if revealed != "A _ _ _ _" {
		t.Errorf("expected 'A _ _ _ _', got %q", revealed)
	}

	// Duplicate via lowercase should fail
	err := g.Guess('a')
	if err == nil {
		t.Error("expected error for case-insensitive duplicate")
	}
}
