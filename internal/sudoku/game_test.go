package sudoku

import "testing"

// A well-known partial board for deterministic testing.
var testPartial = [9][9]int{
	{5, 3, 0, 0, 7, 0, 0, 0, 0},
	{6, 0, 0, 1, 9, 5, 0, 0, 0},
	{0, 9, 8, 0, 0, 0, 0, 6, 0},
	{8, 0, 0, 0, 6, 0, 0, 0, 3},
	{4, 0, 0, 8, 0, 3, 0, 0, 1},
	{7, 0, 0, 0, 2, 0, 0, 0, 6},
	{0, 6, 0, 0, 0, 0, 2, 8, 0},
	{0, 0, 0, 4, 1, 9, 0, 0, 5},
	{0, 0, 0, 0, 8, 0, 0, 7, 9},
}

var testSolution = [9][9]int{
	{5, 3, 4, 6, 7, 8, 9, 1, 2},
	{6, 7, 2, 1, 9, 5, 3, 4, 8},
	{1, 9, 8, 3, 4, 2, 5, 6, 7},
	{8, 5, 9, 7, 6, 1, 4, 2, 3},
	{4, 2, 6, 8, 5, 3, 7, 9, 1},
	{7, 1, 3, 9, 2, 4, 8, 5, 6},
	{9, 6, 1, 5, 3, 7, 2, 8, 4},
	{2, 8, 7, 4, 1, 9, 6, 3, 5},
	{3, 4, 5, 2, 8, 6, 1, 7, 9},
}

// testGivens marks which cells in testPartial are pre-filled.
func testGivens() [9][9]bool {
	var g [9][9]bool
	for r := 0; r < 9; r++ {
		for c := 0; c < 9; c++ {
			g[r][c] = testPartial[r][c] != 0
		}
	}
	return g
}

func TestSolverFindsCorrectSolution(t *testing.T) {
	board := testPartial
	count := solve(&board, 1)
	if count != 1 {
		t.Fatalf("expected 1 solution, solver returned %d", count)
	}
	if board != testSolution {
		t.Error("solver produced incorrect solution")
		for r := 0; r < 9; r++ {
			t.Logf("row %d: got %v want %v", r, board[r], testSolution[r])
		}
	}
}

func TestGeneratedPuzzleHasUniqueSolution(t *testing.T) {
	for _, diff := range []Difficulty{Easy, Medium, Hard} {
		t.Run(diff.String(), func(t *testing.T) {
			g := NewGame(diff)

			// Extract the puzzle as an int board (givens only).
			var board [9][9]int
			for r := 0; r < 9; r++ {
				for c := 0; c < 9; c++ {
					if g.Board[r][c].Given {
						board[r][c] = g.Board[r][c].Value
					}
				}
			}

			count := solve(&board, 2)
			if count != 1 {
				t.Errorf("expected unique solution, got %d solutions", count)
			}
		})
	}
}

func TestIsValidPlacement(t *testing.T) {
	board := testPartial

	tests := []struct {
		name string
		row  int
		col  int
		num  int
		want bool
	}{
		{"valid in empty cell", 0, 2, 4, true},
		{"row conflict", 0, 2, 5, false},
		{"col conflict", 0, 2, 8, false},
		{"box conflict", 0, 2, 6, false},
		{"another valid", 1, 1, 7, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isValidPlacement(&board, tt.row, tt.col, tt.num)
			if got != tt.want {
				t.Errorf("isValidPlacement(row=%d, col=%d, num=%d) = %v, want %v",
					tt.row, tt.col, tt.num, got, tt.want)
			}
		})
	}
}

func TestHasConflict(t *testing.T) {
	givens := testGivens()
	g := NewGameWithBoard(testPartial, testSolution, givens)

	// No conflicts in the initial partial board.
	for r := 0; r < 9; r++ {
		for c := 0; c < 9; c++ {
			if g.Board[r][c].Value != 0 && g.HasConflict(r, c) {
				t.Errorf("unexpected conflict at (%d, %d) value=%d", r, c, g.Board[r][c].Value)
			}
		}
	}

	// Place a conflicting value: row 0 already has 5 at col 0.
	_ = g.SetCell(0, 2, 5)
	if !g.HasConflict(0, 2) {
		t.Error("expected conflict at (0,2) after placing duplicate 5 in row")
	}
}

func TestSetCellRejectsGiven(t *testing.T) {
	givens := testGivens()
	g := NewGameWithBoard(testPartial, testSolution, givens)

	err := g.SetCell(0, 0, 1) // cell (0,0) is given=5
	if err == nil {
		t.Error("expected error when setting a given cell")
	}
}

func TestClearCellRejectsGiven(t *testing.T) {
	givens := testGivens()
	g := NewGameWithBoard(testPartial, testSolution, givens)

	err := g.ClearCell(0, 0)
	if err == nil {
		t.Error("expected error when clearing a given cell")
	}
}

func TestPencilMarkToggle(t *testing.T) {
	givens := testGivens()
	g := NewGameWithBoard(testPartial, testSolution, givens)

	// Cell (0,2) is empty in the test board.
	err := g.TogglePencilMark(0, 2, 4)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !g.Board[0][2].PencilMarks[3] {
		t.Error("pencil mark 4 should be set")
	}

	// Toggle it off.
	err = g.TogglePencilMark(0, 2, 4)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if g.Board[0][2].PencilMarks[3] {
		t.Error("pencil mark 4 should be cleared after second toggle")
	}

	// Pencil marking a given cell should fail.
	err = g.TogglePencilMark(0, 0, 1)
	if err == nil {
		t.Error("expected error when pencil marking a given cell")
	}
}

func TestIsSolved(t *testing.T) {
	tests := []struct {
		name string
		fn   func() *Game
		want bool
	}{
		{
			name: "complete valid board",
			fn: func() *Game {
				var givens [9][9]bool
				for r := 0; r < 9; r++ {
					for c := 0; c < 9; c++ {
						givens[r][c] = true
					}
				}
				return NewGameWithBoard(testSolution, testSolution, givens)
			},
			want: true,
		},
		{
			name: "board with empty cell",
			fn: func() *Game {
				return NewGameWithBoard(testPartial, testSolution, testGivens())
			},
			want: false,
		},
		{
			name: "board with conflict",
			fn: func() *Game {
				bad := testSolution
				bad[0][0] = 3 // duplicates the 3 already at row 0, col 1
				var givens [9][9]bool
				for r := 0; r < 9; r++ {
					for c := 0; c < 9; c++ {
						givens[r][c] = true
					}
				}
				return NewGameWithBoard(bad, testSolution, givens)
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := tt.fn()
			got := g.IsSolved()
			if got != tt.want {
				t.Errorf("IsSolved() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestHint(t *testing.T) {
	givens := testGivens()
	g := NewGameWithBoard(testPartial, testSolution, givens)

	// Cell (0,2) is empty; solution is 4.
	val, err := g.Hint(0, 2)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if val != 4 {
		t.Errorf("Hint(0,2) = %d, want 4", val)
	}

	// Hint on a given cell should fail.
	_, err = g.Hint(0, 0)
	if err == nil {
		t.Error("expected error when hinting a given cell")
	}
}

func TestDifficultyGivenCounts(t *testing.T) {
	tests := []struct {
		diff    Difficulty
		wantMin int
		wantMax int
	}{
		{Easy, 34, 42},
		{Medium, 26, 34},
		{Hard, 20, 28},
	}

	for _, tt := range tests {
		t.Run(tt.diff.String(), func(t *testing.T) {
			g := NewGame(tt.diff)
			count := g.GivenCount()
			if count < tt.wantMin || count > tt.wantMax {
				t.Errorf("%s given count = %d, want [%d, %d]",
					tt.diff.String(), count, tt.wantMin, tt.wantMax)
			}
		})
	}
}
