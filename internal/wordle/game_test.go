package wordle

import "testing"

func TestGuessCorrect(t *testing.T) {
	g := NewGameWithWord("CRANE")
	result, err := g.Guess("CRANE")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	want := GuessResult{Correct, Correct, Correct, Correct, Correct}
	if result != want {
		t.Errorf("result = %v, want %v", result, want)
	}
	if !g.Won {
		t.Error("game should be won")
	}
	if !g.Over {
		t.Error("game should be over")
	}
}

func TestGuessAllAbsent(t *testing.T) {
	g := NewGameWithWord("CRANE")
	result, err := g.Guess("MOLDY")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	want := GuessResult{Absent, Absent, Absent, Absent, Absent}
	if result != want {
		t.Errorf("result = %v, want %v", result, want)
	}
}

func TestDuplicateLetterHandling(t *testing.T) {
	tests := []struct {
		name   string
		target string
		guess  string
		want   GuessResult
	}{
		{
			"ABLER/ALLAY: green L consumes target L, second L absent",
			"ABLER", "ALLAY",
			// A: position 0 exact match -> Correct
			// L: target L at position 2 already consumed by green match (guess[2]==target[2]) -> Absent
			// L: position 2 exact match -> Correct
			// A: no more A in target -> Absent
			// Y: not in target -> Absent
			GuessResult{Correct, Absent, Correct, Absent, Absent},
		},
		{
			"ABBEY/BABES: multiple shared letters",
			"ABBEY", "BABES",
			// B: target has B at 1,2. Guess B at 0 -> Present (matches target[1] or [2])
			// A: target has A at 0. Guess A at 1 -> Present
			// B: target has B remaining. Guess B at 2 -> Correct (position 2)
			// E: target has E at 3. Guess E at 3 -> Correct
			// S: not in target -> Absent
			GuessResult{Present, Present, Correct, Correct, Absent},
		},
		{
			"HELLO/LLAMA: double L in target, double L in guess",
			"HELLO", "LLAMA",
			// L: target has L at 2,3. Guess L at 0 -> Present (consumes target[2])
			// L: target has L at 3 remaining. Guess L at 1 -> Present (consumes target[3])
			// A: not in target -> Absent
			// M: not in target -> Absent
			// A: not in target -> Absent
			GuessResult{Present, Present, Absent, Absent, Absent},
		},
		{
			"CRANE/CREEP: correct C,R + present E, absent E,P",
			"CRANE", "CREEP",
			// C: Correct (pos 0)
			// R: Correct (pos 1)
			// E: target E at 4, guess E at 2 -> Present (consumes target[4])
			// E: no more E -> Absent
			// P: not in target -> Absent
			GuessResult{Correct, Correct, Present, Absent, Absent},
		},
		{
			"SKILL/LLAMA: one L correct, one absent",
			"SKILL", "LLAMA",
			// L: target has L at 3,4. Guess L at 0 -> Present (consumes target[3])
			// L: target has L at 4. Guess L at 1 -> Present (consumes target[4])
			// A: not in target -> Absent
			// M: not in target -> Absent
			// A: not in target -> Absent
			GuessResult{Present, Present, Absent, Absent, Absent},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewGameWithWord(tt.target)
			result, err := g.Guess(tt.guess)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if result != tt.want {
				t.Errorf("result = %v, want %v", result, tt.want)
			}
		})
	}
}

func TestKeyboardStateTracking(t *testing.T) {
	g := NewGameWithWord("CRANE")

	// First guess: CANDY -> C is Correct, A is Present, N is Present
	_, err := g.Guess("CANDY")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if g.KeyboardState['C'] != Correct {
		t.Errorf("C = %d, want Correct(%d)", g.KeyboardState['C'], Correct)
	}
	if g.KeyboardState['A'] != Present {
		t.Errorf("A = %d, want Present(%d)", g.KeyboardState['A'], Present)
	}
	if g.KeyboardState['D'] != Absent {
		t.Errorf("D = %d, want Absent(%d)", g.KeyboardState['D'], Absent)
	}

	// Second guess: CRANE -> all Correct. A should upgrade from Present to Correct.
	_, err = g.Guess("CRANE")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if g.KeyboardState['A'] != Correct {
		t.Errorf("A after second guess = %d, want Correct(%d)", g.KeyboardState['A'], Correct)
	}
	// D should remain Absent (never downgrade).
	if g.KeyboardState['D'] != Absent {
		t.Errorf("D should still be Absent(%d), got %d", Absent, g.KeyboardState['D'])
	}
}

func TestGameOverAfterSixGuesses(t *testing.T) {
	g := NewGameWithWord("CRANE")

	wrongGuesses := []string{"ABOUT", "BELOW", "DIRTY", "GHOST", "JUICE", "MONEY"}
	for _, guess := range wrongGuesses {
		_, err := g.Guess(guess)
		if err != nil {
			t.Fatalf("unexpected error on guess %q: %v", guess, err)
		}
	}

	if !g.Over {
		t.Error("game should be over after 6 guesses")
	}
	if g.Won {
		t.Error("game should not be won")
	}
	if g.RemainingGuesses() != 0 {
		t.Errorf("RemainingGuesses() = %d, want 0", g.RemainingGuesses())
	}
}

func TestInvalidGuessLength(t *testing.T) {
	tests := []struct {
		name  string
		guess string
	}{
		{"too short", "HI"},
		{"too long", "ABCDEF"},
		{"empty", ""},
		{"four letters", "WORD"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewGameWithWord("CRANE")
			_, err := g.Guess(tt.guess)
			if err == nil {
				t.Errorf("Guess(%q) should return error", tt.guess)
			}
		})
	}
}

func TestGuessAfterGameOver(t *testing.T) {
	g := NewGameWithWord("CRANE")
	_, err := g.Guess("CRANE")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !g.Over {
		t.Fatal("game should be over")
	}

	_, err = g.Guess("ABOUT")
	if err == nil {
		t.Error("Guess after game over should return error")
	}
}

func TestGuessNonAlpha(t *testing.T) {
	g := NewGameWithWord("CRANE")
	_, err := g.Guess("AB1CD")
	if err == nil {
		t.Error("Guess with digits should return error")
	}
}

func TestGuessIsCaseInsensitive(t *testing.T) {
	g := NewGameWithWord("CRANE")
	result, err := g.Guess("crane")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	want := GuessResult{Correct, Correct, Correct, Correct, Correct}
	if result != want {
		t.Errorf("lowercase guess result = %v, want %v", result, want)
	}
}

func TestRemainingGuesses(t *testing.T) {
	g := NewGameWithWord("CRANE")
	if g.RemainingGuesses() != 6 {
		t.Errorf("initial RemainingGuesses() = %d, want 6", g.RemainingGuesses())
	}

	g.Guess("ABOUT") //nolint:errcheck // error irrelevant; testing remaining count
	if g.RemainingGuesses() != 5 {
		t.Errorf("after 1 guess RemainingGuesses() = %d, want 5", g.RemainingGuesses())
	}
}

func TestNewGameSelectsFromWordList(t *testing.T) {
	g := NewGame()
	if len(g.Target) != 5 {
		t.Errorf("target length = %d, want 5", len(g.Target))
	}

	found := false
	for _, w := range solutionWords {
		if w == g.Target {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("target %q not in solutionWords", g.Target)
	}
}
