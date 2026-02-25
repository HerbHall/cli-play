package mastermind

import "testing"

func TestNewGame(t *testing.T) {
	g := NewGame()

	if g.GuessCount() != 0 {
		t.Errorf("new game should have 0 guesses, got %d", g.GuessCount())
	}
	if g.IsOver() {
		t.Error("new game should not be over")
	}
	if g.Won() {
		t.Error("new game should not be won")
	}

	secret := g.Secret()
	for i, v := range secret {
		if v < 1 || v > NumColors {
			t.Errorf("secret[%d] = %d, want 1-%d", i, v, NumColors)
		}
	}
}

func TestAllExact(t *testing.T) {
	secret := [CodeLength]int{1, 2, 3, 4}
	g := newGameWithSecret(secret)

	fb := g.Guess([CodeLength]int{1, 2, 3, 4})
	if fb.Exact != 4 || fb.Misplaced != 0 {
		t.Errorf("all exact: got (%d, %d), want (4, 0)", fb.Exact, fb.Misplaced)
	}
}

func TestAllMisplaced(t *testing.T) {
	secret := [CodeLength]int{1, 2, 3, 4}
	g := newGameWithSecret(secret)

	fb := g.Guess([CodeLength]int{4, 3, 2, 1})
	if fb.Exact != 0 || fb.Misplaced != 4 {
		t.Errorf("all misplaced: got (%d, %d), want (0, 4)", fb.Exact, fb.Misplaced)
	}
}

func TestMixed(t *testing.T) {
	tests := []struct {
		name      string
		secret    [CodeLength]int
		guess     [CodeLength]int
		wantExact int
		wantMisp  int
	}{
		{
			name:      "one exact one misplaced",
			secret:    [CodeLength]int{1, 2, 3, 4},
			guess:     [CodeLength]int{1, 3, 5, 6},
			wantExact: 1,
			wantMisp:  1,
		},
		{
			name:      "two exact one misplaced",
			secret:    [CodeLength]int{1, 2, 3, 4},
			guess:     [CodeLength]int{1, 2, 4, 6},
			wantExact: 2,
			wantMisp:  1,
		},
		{
			name:      "three exact zero misplaced",
			secret:    [CodeLength]int{1, 2, 3, 4},
			guess:     [CodeLength]int{1, 2, 3, 5},
			wantExact: 3,
			wantMisp:  0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := newGameWithSecret(tt.secret)
			fb := g.Guess(tt.guess)
			if fb.Exact != tt.wantExact || fb.Misplaced != tt.wantMisp {
				t.Errorf("got (%d, %d), want (%d, %d)",
					fb.Exact, fb.Misplaced, tt.wantExact, tt.wantMisp)
			}
		})
	}
}

func TestNoMatch(t *testing.T) {
	secret := [CodeLength]int{1, 2, 3, 4}
	g := newGameWithSecret(secret)

	fb := g.Guess([CodeLength]int{5, 5, 5, 5})
	if fb.Exact != 0 || fb.Misplaced != 0 {
		t.Errorf("no match: got (%d, %d), want (0, 0)", fb.Exact, fb.Misplaced)
	}
}

func TestDuplicateColors(t *testing.T) {
	tests := []struct {
		name      string
		secret    [CodeLength]int
		guess     [CodeLength]int
		wantExact int
		wantMisp  int
	}{
		{
			name:      "secret has dups guess has one match",
			secret:    [CodeLength]int{1, 1, 2, 3},
			guess:     [CodeLength]int{1, 4, 4, 4},
			wantExact: 1,
			wantMisp:  0,
		},
		{
			name:      "guess has dups secret has one",
			secret:    [CodeLength]int{1, 2, 3, 4},
			guess:     [CodeLength]int{1, 1, 1, 1},
			wantExact: 1,
			wantMisp:  0,
		},
		{
			name:      "both have dups partial match",
			secret:    [CodeLength]int{1, 1, 2, 2},
			guess:     [CodeLength]int{2, 2, 1, 1},
			wantExact: 0,
			wantMisp:  4,
		},
		{
			name:      "both have dups with exact",
			secret:    [CodeLength]int{1, 1, 2, 3},
			guess:     [CodeLength]int{1, 1, 3, 2},
			wantExact: 2,
			wantMisp:  2,
		},
		{
			name:      "guess duplicates exceed secret count",
			secret:    [CodeLength]int{1, 2, 3, 4},
			guess:     [CodeLength]int{2, 2, 2, 2},
			wantExact: 1,
			wantMisp:  0,
		},
		{
			name:      "secret all same guess all same different",
			secret:    [CodeLength]int{3, 3, 3, 3},
			guess:     [CodeLength]int{5, 5, 5, 5},
			wantExact: 0,
			wantMisp:  0,
		},
		{
			name:      "secret all same guess all same match",
			secret:    [CodeLength]int{3, 3, 3, 3},
			guess:     [CodeLength]int{3, 3, 3, 3},
			wantExact: 4,
			wantMisp:  0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := newGameWithSecret(tt.secret)
			fb := g.Guess(tt.guess)
			if fb.Exact != tt.wantExact || fb.Misplaced != tt.wantMisp {
				t.Errorf("got (%d, %d), want (%d, %d)",
					fb.Exact, fb.Misplaced, tt.wantExact, tt.wantMisp)
			}
		})
	}
}

func TestWinCondition(t *testing.T) {
	secret := [CodeLength]int{3, 5, 2, 6}
	g := newGameWithSecret(secret)

	// Wrong guess first
	g.Guess([CodeLength]int{1, 1, 1, 1})
	if g.Won() {
		t.Error("should not have won after wrong guess")
	}
	if g.IsOver() {
		t.Error("should not be over after 1 guess")
	}

	// Correct guess
	g.Guess(secret)
	if !g.Won() {
		t.Error("should have won after exact match")
	}
	if !g.IsOver() {
		t.Error("game should be over after winning")
	}
}

func TestLoseCondition(t *testing.T) {
	secret := [CodeLength]int{6, 6, 6, 6}
	g := newGameWithSecret(secret)

	wrong := [CodeLength]int{1, 2, 3, 4}
	for range MaxGuesses {
		g.Guess(wrong)
	}

	if !g.IsOver() {
		t.Error("game should be over after max guesses")
	}
	if g.Won() {
		t.Error("should not have won with all wrong guesses")
	}
}

func TestGuessCount(t *testing.T) {
	g := newGameWithSecret([CodeLength]int{1, 2, 3, 4})

	for i := 1; i <= 5; i++ {
		g.Guess([CodeLength]int{5, 5, 5, 5})
		if g.GuessCount() != i {
			t.Errorf("after %d guesses, GuessCount() = %d", i, g.GuessCount())
		}
	}
}

func TestGuessAfterGameOver(t *testing.T) {
	g := newGameWithSecret([CodeLength]int{1, 2, 3, 4})
	g.Guess([CodeLength]int{1, 2, 3, 4}) // win

	fb := g.Guess([CodeLength]int{5, 5, 5, 5})
	if fb.Exact != -1 || fb.Misplaced != -1 {
		t.Errorf("guess after game over: got (%d, %d), want (-1, -1)",
			fb.Exact, fb.Misplaced)
	}
	if g.GuessCount() != 1 {
		t.Errorf("guess count should not change after game over, got %d", g.GuessCount())
	}
}

func TestMaxGuessCount(t *testing.T) {
	g := NewGame()
	if g.MaxGuessCount() != MaxGuesses {
		t.Errorf("MaxGuessCount() = %d, want %d", g.MaxGuessCount(), MaxGuesses)
	}
}

func TestGuessesHistory(t *testing.T) {
	g := newGameWithSecret([CodeLength]int{1, 2, 3, 4})

	guess1 := [CodeLength]int{1, 1, 1, 1}
	guess2 := [CodeLength]int{2, 2, 2, 2}
	g.Guess(guess1)
	g.Guess(guess2)

	guesses := g.Guesses()
	if len(guesses) != 2 {
		t.Fatalf("expected 2 guesses, got %d", len(guesses))
	}
	if guesses[0].Code != guess1 {
		t.Errorf("guesses[0].Code = %v, want %v", guesses[0].Code, guess1)
	}
	if guesses[1].Code != guess2 {
		t.Errorf("guesses[1].Code = %v, want %v", guesses[1].Code, guess2)
	}
}
