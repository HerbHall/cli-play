package minesweeper

import "testing"

// testGrid creates a 5x5 grid with mines at specified positions.
// Default layout used by several tests:
//
//	Mines at (0,0), (0,4), (2,2), (4,0), (4,4)
//
//	M 1 0 1 M
//	1 2 1 2 1
//	0 1 M 1 0
//	1 2 1 2 1
//	M 1 0 1 M
func testGrid() *Game {
	mines := [][2]int{{0, 0}, {0, 4}, {2, 2}, {4, 0}, {4, 4}}
	return NewGameWithMines(5, 5, mines)
}

func TestAdjacentCounts(t *testing.T) {
	g := testGrid()

	tests := []struct {
		name string
		row  int
		col  int
		want int
	}{
		{"corner no mine (0,1)", 0, 1, 1},
		{"center of grid (2,2) is mine", 2, 2, 0}, // mine cells don't count
		{"cell (1,1) near 2 mines", 1, 1, 2},
		{"cell (1,2) near 1 mine", 1, 2, 1},
		{"cell (1,3) near 2 mines", 1, 3, 2},
		{"center empty (2,0)", 2, 0, 0},
		{"cell (3,1) near 2 mines", 3, 1, 2},
		{"cell (0,2) zero adjacent", 0, 2, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cell := g.Grid[tt.row][tt.col]
			if cell.Mine {
				return // mine cells have Adjacent=0 by design
			}
			if cell.Adjacent != tt.want {
				t.Errorf("Grid[%d][%d].Adjacent = %d, want %d",
					tt.row, tt.col, cell.Adjacent, tt.want)
			}
		})
	}
}

func TestRevealEmpty(t *testing.T) {
	g := testGrid()

	// Reveal (0,2) which has 0 adjacent -- should flood-fill.
	ok := g.Reveal(0, 2)
	if !ok {
		t.Fatal("Reveal(0,2) returned false, want true")
	}

	// Expected revealed cells from flood-fill starting at (0,2):
	// (0,2) has 0 adjacent -> reveals neighbors
	// (0,1) has 1 -> revealed but stops
	// (0,3) has 1 -> revealed but stops
	// (1,1) has 2 -> revealed but stops
	// (1,2) has 1 -> revealed but stops
	// (1,3) has 2 -> revealed but stops
	wantRevealed := [][2]int{
		{0, 1}, {0, 2}, {0, 3},
		{1, 1}, {1, 2}, {1, 3},
	}

	for _, pos := range wantRevealed {
		cell := g.Grid[pos[0]][pos[1]]
		if cell.State != Revealed {
			t.Errorf("Grid[%d][%d] should be Revealed after flood-fill, got %d",
				pos[0], pos[1], cell.State)
		}
	}

	// Mines and far corners should remain hidden.
	wantHidden := [][2]int{
		{0, 0}, {0, 4}, {2, 2}, {4, 0}, {4, 4},
		{2, 0}, {3, 0}, {4, 1},
	}
	for _, pos := range wantHidden {
		cell := g.Grid[pos[0]][pos[1]]
		if cell.State == Revealed && !cell.Mine {
			t.Errorf("Grid[%d][%d] should be Hidden, got Revealed",
				pos[0], pos[1])
		}
	}
}

func TestRevealMine(t *testing.T) {
	g := testGrid()

	ok := g.Reveal(0, 0)
	if !ok {
		t.Fatal("Reveal(0,0) returned false")
	}
	if g.State != Lost {
		t.Errorf("State = %d, want Lost (%d)", g.State, Lost)
	}

	// All mines should be revealed.
	minePositions := [][2]int{{0, 0}, {0, 4}, {2, 2}, {4, 0}, {4, 4}}
	for _, pos := range minePositions {
		cell := g.Grid[pos[0]][pos[1]]
		if cell.State != Revealed {
			t.Errorf("Mine at (%d,%d) should be Revealed after loss, got %d",
				pos[0], pos[1], cell.State)
		}
	}
}

func TestToggleFlag(t *testing.T) {
	g := testGrid()

	// Flag a hidden cell.
	g.ToggleFlag(1, 0)
	if g.Grid[1][0].State != Flagged {
		t.Error("Cell should be Flagged after ToggleFlag")
	}
	if g.FlagsUsed != 1 {
		t.Errorf("FlagsUsed = %d, want 1", g.FlagsUsed)
	}

	// Can't reveal a flagged cell.
	ok := g.Reveal(1, 0)
	if ok {
		t.Error("Reveal on flagged cell should return false")
	}

	// Unflag.
	g.ToggleFlag(1, 0)
	if g.Grid[1][0].State != Hidden {
		t.Error("Cell should be Hidden after second ToggleFlag")
	}
	if g.FlagsUsed != 0 {
		t.Errorf("FlagsUsed = %d, want 0", g.FlagsUsed)
	}

	// Can't flag a revealed cell.
	g.Reveal(1, 0)
	g.ToggleFlag(1, 0)
	if g.Grid[1][0].State != Revealed {
		t.Error("Should not be able to flag a revealed cell")
	}
}

func TestWinDetection(t *testing.T) {
	g := testGrid()

	// Reveal all non-mine cells. 5x5 grid with 5 mines = 20 safe cells.
	for r := 0; r < g.Rows; r++ {
		for c := 0; c < g.Cols; c++ {
			if !g.Grid[r][c].Mine {
				g.Reveal(r, c)
			}
		}
	}

	if g.State != Won {
		t.Errorf("State = %d, want Won (%d)", g.State, Won)
	}
	if g.CellsRevealed != 20 {
		t.Errorf("CellsRevealed = %d, want 20", g.CellsRevealed)
	}
}

func TestFirstClickSafe(t *testing.T) {
	// Create a beginner game and click a bunch of times at various positions.
	// The first click should never hit a mine.
	for i := 0; i < 50; i++ {
		g := NewGame(Beginner)
		row := i % 9
		col := (i * 3) % 9

		g.Reveal(row, col)

		if g.State == Lost {
			t.Errorf("First click at (%d,%d) hit a mine on iteration %d",
				row, col, i)
		}
		if g.FirstClick {
			t.Error("FirstClick should be false after first Reveal")
		}
	}
}

func TestBoundsCheck(t *testing.T) {
	g := testGrid()

	tests := []struct {
		name string
		row  int
		col  int
	}{
		{"negative row", -1, 0},
		{"negative col", 0, -1},
		{"row too large", 5, 0},
		{"col too large", 0, 5},
		{"both negative", -1, -1},
		{"both too large", 10, 10},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if g.Reveal(tt.row, tt.col) {
				t.Errorf("Reveal(%d,%d) should return false for out-of-bounds",
					tt.row, tt.col)
			}
		})
	}
}

func TestNewGameWithMinesComputation(t *testing.T) {
	// Simple 3x3 grid with one mine in the center.
	g := NewGameWithMines(3, 3, [][2]int{{1, 1}})

	// All 8 surrounding cells should have Adjacent=1.
	for r := 0; r < 3; r++ {
		for c := 0; c < 3; c++ {
			if r == 1 && c == 1 {
				if !g.Grid[r][c].Mine {
					t.Error("Center cell should be a mine")
				}
				continue
			}
			if g.Grid[r][c].Adjacent != 1 {
				t.Errorf("Grid[%d][%d].Adjacent = %d, want 1",
					r, c, g.Grid[r][c].Adjacent)
			}
		}
	}
}

func TestRevealAlreadyRevealed(t *testing.T) {
	g := testGrid()

	g.Reveal(1, 0)
	before := g.CellsRevealed

	ok := g.Reveal(1, 0)
	if ok {
		t.Error("Reveal on already-revealed cell should return false")
	}
	if g.CellsRevealed != before {
		t.Error("CellsRevealed should not change on double reveal")
	}
}

func TestRevealAfterGameOver(t *testing.T) {
	g := testGrid()

	// Lose the game.
	g.Reveal(0, 0)
	if g.State != Lost {
		t.Fatal("Expected Lost state")
	}

	// Subsequent reveals should be rejected.
	ok := g.Reveal(1, 0)
	if ok {
		t.Error("Reveal after game over should return false")
	}
}

func TestToggleFlagAfterGameOver(t *testing.T) {
	g := testGrid()

	g.Reveal(0, 0) // lose
	before := g.FlagsUsed

	g.ToggleFlag(1, 0)
	if g.FlagsUsed != before {
		t.Error("Should not be able to flag after game over")
	}
}
