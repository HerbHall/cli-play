package memory

import "testing"

// testBoard returns a deterministic board for testing.
//
//	A B C D
//	A B C D
//	E F G H
//	E F G H
func testBoard() [rows][cols]byte {
	return [rows][cols]byte{
		{'A', 'B', 'C', 'D'},
		{'A', 'B', 'C', 'D'},
		{'E', 'F', 'G', 'H'},
		{'E', 'F', 'G', 'H'},
	}
}

func TestInitialBoardSetup(t *testing.T) {
	g := NewGame()

	// Count symbol occurrences -- each should appear exactly twice.
	counts := make(map[byte]int)
	for r := 0; r < rows; r++ {
		for c := 0; c < cols; c++ {
			card := g.Board[r][c]
			counts[card.Symbol]++
			if card.State != FaceDown {
				t.Errorf("Board[%d][%d] should be FaceDown, got %d", r, c, card.State)
			}
		}
	}

	if len(counts) != numPairs {
		t.Errorf("expected %d unique symbols, got %d", numPairs, len(counts))
	}
	for sym, count := range counts {
		if count != 2 {
			t.Errorf("symbol %c appears %d times, want 2", sym, count)
		}
	}
}

func TestFlipCard(t *testing.T) {
	g := NewGameWithBoard(testBoard())

	// First flip
	ok := g.FlipCard(0, 0)
	if !ok {
		t.Fatal("FlipCard(0,0) returned false, want true")
	}
	if g.Board[0][0].State != FaceUp {
		t.Error("Card should be FaceUp after flip")
	}
	if !g.HasFirst {
		t.Error("HasFirst should be true after first flip")
	}
	if g.HasSecond {
		t.Error("HasSecond should be false after first flip")
	}
	if g.Moves != 0 {
		t.Errorf("Moves = %d after first flip, want 0", g.Moves)
	}

	// Can't flip the same card again
	ok = g.FlipCard(0, 0)
	if ok {
		t.Error("Should not be able to flip already face-up card")
	}
}

func TestMatchingPair(t *testing.T) {
	g := NewGameWithBoard(testBoard())

	// Flip A at (0,0) and A at (1,0) -- should match
	g.FlipCard(0, 0)
	g.FlipCard(1, 0)

	if !g.HasSecond {
		t.Fatal("HasSecond should be true after second flip")
	}
	if g.Moves != 1 {
		t.Errorf("Moves = %d, want 1", g.Moves)
	}
	if !g.CheckMatch() {
		t.Error("CheckMatch should return true for matching pair")
	}

	g.ResolveMatch()

	if g.Board[0][0].State != Matched {
		t.Error("First card should be Matched")
	}
	if g.Board[1][0].State != Matched {
		t.Error("Second card should be Matched")
	}
	if g.PairsFound != 1 {
		t.Errorf("PairsFound = %d, want 1", g.PairsFound)
	}
	if g.HasFirst || g.HasSecond {
		t.Error("Picks should be reset after match resolution")
	}
}

func TestNonMatchingPair(t *testing.T) {
	g := NewGameWithBoard(testBoard())

	// Flip A at (0,0) and B at (0,1) -- no match
	g.FlipCard(0, 0)
	g.FlipCard(0, 1)

	if !g.HasSecond {
		t.Fatal("HasSecond should be true")
	}
	if g.CheckMatch() {
		t.Error("CheckMatch should return false for non-matching pair")
	}

	g.ResolveNoMatch()

	if g.Board[0][0].State != FaceDown {
		t.Error("First card should be FaceDown after no match")
	}
	if g.Board[0][1].State != FaceDown {
		t.Error("Second card should be FaceDown after no match")
	}
	if g.HasFirst || g.HasSecond {
		t.Error("Picks should be reset after no-match resolution")
	}
	if g.PairsFound != 0 {
		t.Errorf("PairsFound = %d, want 0", g.PairsFound)
	}
}

func TestGameCompletion(t *testing.T) {
	g := NewGameWithBoard(testBoard())

	// Match all pairs in order: A, B, C, D, E, F, G, H
	pairs := [][2][2]int{
		{{0, 0}, {1, 0}}, // A
		{{0, 1}, {1, 1}}, // B
		{{0, 2}, {1, 2}}, // C
		{{0, 3}, {1, 3}}, // D
		{{2, 0}, {3, 0}}, // E
		{{2, 1}, {3, 1}}, // F
		{{2, 2}, {3, 2}}, // G
		{{2, 3}, {3, 3}}, // H
	}

	for _, pair := range pairs {
		g.FlipCard(pair[0][0], pair[0][1])
		g.FlipCard(pair[1][0], pair[1][1])
		if !g.CheckMatch() {
			t.Fatalf("Expected match for %c at (%d,%d) and (%d,%d)",
				g.Board[pair[0][0]][pair[0][1]].Symbol,
				pair[0][0], pair[0][1], pair[1][0], pair[1][1])
		}
		g.ResolveMatch()
	}

	if !g.GameOver {
		t.Error("GameOver should be true after all pairs found")
	}
	if g.PairsFound != numPairs {
		t.Errorf("PairsFound = %d, want %d", g.PairsFound, numPairs)
	}
	if g.Moves != 8 {
		t.Errorf("Moves = %d, want 8 (perfect game)", g.Moves)
	}
}

func TestMoveCount(t *testing.T) {
	g := NewGameWithBoard(testBoard())

	// First attempt: A and B -- no match (move 1)
	g.FlipCard(0, 0) // A
	g.FlipCard(0, 1) // B
	g.ResolveNoMatch()

	// Second attempt: A and A -- match (move 2)
	g.FlipCard(0, 0) // A
	g.FlipCard(1, 0) // A
	g.ResolveMatch()

	if g.Moves != 2 {
		t.Errorf("Moves = %d, want 2", g.Moves)
	}
}

func TestCannotFlipMatchedCard(t *testing.T) {
	g := NewGameWithBoard(testBoard())

	// Match A pair
	g.FlipCard(0, 0)
	g.FlipCard(1, 0)
	g.ResolveMatch()

	// Try to flip a matched card
	ok := g.FlipCard(0, 0)
	if ok {
		t.Error("Should not be able to flip a matched card")
	}
}

func TestCannotFlipDuringSecondReveal(t *testing.T) {
	g := NewGameWithBoard(testBoard())

	// Flip two non-matching cards
	g.FlipCard(0, 0) // A
	g.FlipCard(0, 1) // B

	// HasSecond is true, should block further flips
	ok := g.FlipCard(0, 2)
	if ok {
		t.Error("Should not be able to flip a third card while two are revealed")
	}
}

func TestCannotFlipAfterGameOver(t *testing.T) {
	g := NewGameWithBoard(testBoard())

	// Complete the game
	pairs := [][2][2]int{
		{{0, 0}, {1, 0}}, {{0, 1}, {1, 1}},
		{{0, 2}, {1, 2}}, {{0, 3}, {1, 3}},
		{{2, 0}, {3, 0}}, {{2, 1}, {3, 1}},
		{{2, 2}, {3, 2}}, {{2, 3}, {3, 3}},
	}
	for _, pair := range pairs {
		g.FlipCard(pair[0][0], pair[0][1])
		g.FlipCard(pair[1][0], pair[1][1])
		g.ResolveMatch()
	}

	ok := g.FlipCard(0, 0)
	if ok {
		t.Error("Should not be able to flip after game over")
	}
}

func TestBoundsCheck(t *testing.T) {
	g := NewGameWithBoard(testBoard())

	tests := []struct {
		name string
		row  int
		col  int
	}{
		{"negative row", -1, 0},
		{"negative col", 0, -1},
		{"row too large", rows, 0},
		{"col too large", 0, cols},
		{"both negative", -1, -1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if g.FlipCard(tt.row, tt.col) {
				t.Errorf("FlipCard(%d,%d) should return false for out-of-bounds",
					tt.row, tt.col)
			}
		})
	}
}

func TestNewGameShuffles(t *testing.T) {
	// Create multiple games and verify they're not all identical.
	// With 16! / (2^8) arrangements, identical boards are astronomically unlikely.
	boards := make([][rows][cols]byte, 10)
	for i := range boards {
		g := NewGame()
		for r := 0; r < rows; r++ {
			for c := 0; c < cols; c++ {
				boards[i][r][c] = g.Board[r][c].Symbol
			}
		}
	}

	allSame := true
	for i := 1; i < len(boards); i++ {
		if boards[i] != boards[0] {
			allSame = false
			break
		}
	}
	if allSame {
		t.Error("10 consecutive games produced identical boards -- shuffle may be broken")
	}
}
