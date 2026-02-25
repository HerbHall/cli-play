package fifteenpuzzle

import "testing"

func solvedBoard() Board {
	return Board{
		{1, 2, 3, 4},
		{5, 6, 7, 8},
		{9, 10, 11, 12},
		{13, 14, 15, 0},
	}
}

func almostSolvedBoard() Board {
	// One move away: empty at (3,2), tile 15 at (3,3)
	return Board{
		{1, 2, 3, 4},
		{5, 6, 7, 8},
		{9, 10, 11, 12},
		{13, 14, 0, 15},
	}
}

func TestNewGameIsSolvable(t *testing.T) {
	for range 100 {
		g := NewGame()
		if !IsSolvable(g.Board) {
			t.Fatal("NewGame() produced an unsolvable board")
		}
	}
}

func TestNewGameIsNotSolved(t *testing.T) {
	for range 100 {
		g := NewGame()
		if g.Won {
			t.Fatal("NewGame() should not start in a won state")
		}
	}
}

func TestIsSolvable(t *testing.T) {
	tests := []struct {
		name     string
		board    Board
		solvable bool
	}{
		{
			name:     "solved position is solvable",
			board:    solvedBoard(),
			solvable: true,
		},
		{
			name:     "almost solved is solvable",
			board:    almostSolvedBoard(),
			solvable: true,
		},
		{
			name: "classic unsolvable (14-15 swap)",
			board: Board{
				{1, 2, 3, 4},
				{5, 6, 7, 8},
				{9, 10, 11, 12},
				{13, 15, 14, 0},
			},
			solvable: false,
		},
		{
			name: "unsolvable (1-2 swap)",
			board: Board{
				{2, 1, 3, 4},
				{5, 6, 7, 8},
				{9, 10, 11, 12},
				{13, 14, 15, 0},
			},
			solvable: false,
		},
		{
			name: "solvable scrambled",
			board: Board{
				{5, 1, 3, 4},
				{2, 6, 7, 8},
				{9, 10, 11, 12},
				{13, 14, 15, 0},
			},
			solvable: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsSolvable(tt.board)
			if got != tt.solvable {
				t.Errorf("IsSolvable() = %v, want %v", got, tt.solvable)
			}
		})
	}
}

func TestMoveUp(t *testing.T) {
	// Empty at bottom-right (3,3). MoveUp slides tile from (3+1) which is out of bounds -- invalid.
	// Use a board where empty is not on the bottom row.
	b := Board{
		{1, 2, 3, 4},
		{5, 6, 0, 8},
		{9, 10, 7, 12},
		{13, 14, 15, 11},
	}
	g := NewGameFromBoard(b)

	if !g.MoveUp() {
		t.Fatal("MoveUp should succeed when tile exists below empty")
	}
	if g.Board[1][2] != 7 {
		t.Errorf("expected tile 7 to slide up, got %d", g.Board[1][2])
	}
	if g.Board[2][2] != 0 {
		t.Error("expected empty space to move down")
	}
	if g.Moves != 1 {
		t.Errorf("expected 1 move, got %d", g.Moves)
	}
}

func TestMoveDown(t *testing.T) {
	b := Board{
		{1, 2, 3, 4},
		{5, 6, 0, 8},
		{9, 10, 7, 12},
		{13, 14, 15, 11},
	}
	g := NewGameFromBoard(b)

	if !g.MoveDown() {
		t.Fatal("MoveDown should succeed when tile exists above empty")
	}
	if g.Board[1][2] != 3 {
		t.Errorf("expected tile 3 to slide down, got %d", g.Board[1][2])
	}
	if g.Board[0][2] != 0 {
		t.Error("expected empty space to move up")
	}
}

func TestMoveLeft(t *testing.T) {
	b := Board{
		{1, 2, 3, 4},
		{5, 6, 0, 8},
		{9, 10, 7, 12},
		{13, 14, 15, 11},
	}
	g := NewGameFromBoard(b)

	if !g.MoveLeft() {
		t.Fatal("MoveLeft should succeed when tile exists to the right of empty")
	}
	if g.Board[1][2] != 8 {
		t.Errorf("expected tile 8 to slide left, got %d", g.Board[1][2])
	}
	if g.Board[1][3] != 0 {
		t.Error("expected empty space to move right")
	}
}

func TestMoveRight(t *testing.T) {
	b := Board{
		{1, 2, 3, 4},
		{5, 6, 0, 8},
		{9, 10, 7, 12},
		{13, 14, 15, 11},
	}
	g := NewGameFromBoard(b)

	if !g.MoveRight() {
		t.Fatal("MoveRight should succeed when tile exists to the left of empty")
	}
	if g.Board[1][2] != 6 {
		t.Errorf("expected tile 6 to slide right, got %d", g.Board[1][2])
	}
	if g.Board[1][1] != 0 {
		t.Error("expected empty space to move left")
	}
}

func TestMoveInvalidEdge(t *testing.T) {
	// Empty at top-left corner (0,0)
	b := Board{
		{0, 1, 2, 3},
		{5, 6, 7, 4},
		{9, 10, 11, 8},
		{13, 14, 15, 12},
	}
	g := NewGameFromBoard(b)

	if g.MoveDown() {
		t.Error("MoveDown should fail when empty is on top row (no tile above)")
	}
	if g.MoveRight() {
		t.Error("MoveRight should fail when empty is on left edge (no tile to the left)")
	}
	if g.Moves != 0 {
		t.Errorf("expected 0 moves after invalid attempts, got %d", g.Moves)
	}
}

func TestMoveInvalidBottomRight(t *testing.T) {
	g := NewGameFromBoard(solvedBoard())
	g.Won = false // Override so we can test movement

	if g.MoveUp() {
		t.Error("MoveUp should fail when empty is on bottom row (no tile below)")
	}
	if g.MoveLeft() {
		t.Error("MoveLeft should fail when empty is on right edge (no tile to the right)")
	}
}

func TestWinDetection(t *testing.T) {
	g := NewGameFromBoard(almostSolvedBoard())

	if g.Won {
		t.Fatal("game should not be won before the final move")
	}

	// MoveLeft slides tile 15 from (3,3) to (3,2), empty moves to (3,3)
	if !g.MoveLeft() {
		t.Fatal("MoveLeft should succeed")
	}

	if !g.Won {
		t.Error("game should be won after completing the puzzle")
	}
}

func TestMoveCountTracking(t *testing.T) {
	b := Board{
		{1, 2, 3, 4},
		{5, 6, 0, 8},
		{9, 10, 7, 12},
		{13, 14, 15, 11},
	}
	g := NewGameFromBoard(b)

	g.MoveUp()
	g.MoveLeft()
	g.MoveDown()

	if g.Moves != 3 {
		t.Errorf("expected 3 moves, got %d", g.Moves)
	}
}

func TestNoMoveAfterWin(t *testing.T) {
	g := NewGameFromBoard(almostSolvedBoard())
	g.MoveLeft() // Win

	if !g.Won {
		t.Fatal("expected game to be won")
	}

	moved := g.MoveUp()
	if moved {
		t.Error("should not allow moves after winning")
	}
	if g.Moves != 1 {
		t.Errorf("expected 1 move, got %d", g.Moves)
	}
}

func TestReset(t *testing.T) {
	g := NewGame()
	g.MoveUp()
	g.MoveDown()

	g.Reset()

	if g.Moves != 0 {
		t.Errorf("expected 0 moves after reset, got %d", g.Moves)
	}
	if g.Won {
		t.Error("expected not won after reset")
	}
	if !IsSolvable(g.Board) {
		t.Error("board should be solvable after reset")
	}
}

func TestCountInversions(t *testing.T) {
	tests := []struct {
		name  string
		flat  []int
		count int
	}{
		{"sorted", []int{1, 2, 3, 4, 5}, 0},
		{"reversed", []int{5, 4, 3, 2, 1}, 10},
		{"one swap", []int{1, 3, 2, 4, 5}, 1},
		{"two inversions", []int{2, 1, 4, 3, 5}, 2},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := countInversions(tt.flat)
			if got != tt.count {
				t.Errorf("countInversions(%v) = %d, want %d", tt.flat, got, tt.count)
			}
		})
	}
}

func TestNewGameFromBoard(t *testing.T) {
	b := Board{
		{1, 2, 3, 4},
		{5, 6, 0, 8},
		{9, 10, 7, 12},
		{13, 14, 15, 11},
	}
	g := NewGameFromBoard(b)

	if g.EmptyRow != 1 || g.EmptyCol != 2 {
		t.Errorf("expected empty at (1,2), got (%d,%d)", g.EmptyRow, g.EmptyCol)
	}
	if g.Moves != 0 {
		t.Errorf("expected 0 moves, got %d", g.Moves)
	}
}
