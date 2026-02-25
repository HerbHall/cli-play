package twofortyeight

import "testing"

// setBoard directly sets the game board for deterministic testing.
func setBoard(g *Game, board [4][4]int) {
	g.Board = board
}

func TestSlideLeft(t *testing.T) {
	tests := []struct {
		name      string
		board     [4][4]int
		wantBoard [4][4]int
		wantScore int
	}{
		{
			"merge pair",
			[4][4]int{
				{2, 2, 0, 0},
				{0, 0, 0, 0},
				{0, 0, 0, 0},
				{0, 0, 0, 0},
			},
			[4][4]int{
				{4, 0, 0, 0},
				{0, 0, 0, 0},
				{0, 0, 0, 0},
				{0, 0, 0, 0},
			},
			4,
		},
		{
			"no double merge",
			[4][4]int{
				{2, 2, 2, 2},
				{0, 0, 0, 0},
				{0, 0, 0, 0},
				{0, 0, 0, 0},
			},
			[4][4]int{
				{4, 4, 0, 0},
				{0, 0, 0, 0},
				{0, 0, 0, 0},
				{0, 0, 0, 0},
			},
			8,
		},
		{
			"slide with gap",
			[4][4]int{
				{2, 0, 0, 2},
				{0, 0, 0, 0},
				{0, 0, 0, 0},
				{0, 0, 0, 0},
			},
			[4][4]int{
				{4, 0, 0, 0},
				{0, 0, 0, 0},
				{0, 0, 0, 0},
				{0, 0, 0, 0},
			},
			4,
		},
		{
			"no merge different values",
			[4][4]int{
				{2, 4, 8, 16},
				{0, 0, 0, 0},
				{0, 0, 0, 0},
				{0, 0, 0, 0},
			},
			[4][4]int{
				{2, 4, 8, 16},
				{0, 0, 0, 0},
				{0, 0, 0, 0},
				{0, 0, 0, 0},
			},
			0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := &Game{}
			setBoard(g, tt.board)

			// Test slideLine directly on row 0 for determinism (no spawn).
			result := slideLine(g.Board[0])
			g.Board[0] = result.line

			if g.Board[0] != tt.wantBoard[0] {
				t.Errorf("board row 0 = %v, want %v", g.Board[0], tt.wantBoard[0])
			}
			if result.score != tt.wantScore {
				t.Errorf("score = %d, want %d", result.score, tt.wantScore)
			}
		})
	}
}

func TestSlideRight(t *testing.T) {
	tests := []struct {
		name    string
		line    [4]int
		want    [4]int
		wantScr int
	}{
		{"merge right", [4]int{0, 0, 2, 2}, [4]int{0, 0, 0, 4}, 4},
		{"slide to right", [4]int{2, 0, 0, 0}, [4]int{0, 0, 0, 2}, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rev := reverseLine(tt.line)
			result := slideLine(rev)
			got := reverseLine(result.line)
			if got != tt.want {
				t.Errorf("got %v, want %v", got, tt.want)
			}
			if result.score != tt.wantScr {
				t.Errorf("score = %d, want %d", result.score, tt.wantScr)
			}
		})
	}
}

func TestSlideUp(t *testing.T) {
	g := &Game{}
	setBoard(g, [4][4]int{
		{2, 0, 0, 0},
		{2, 0, 0, 0},
		{0, 0, 0, 0},
		{0, 0, 0, 0},
	})

	col := extractColumn(g.Board, 0)
	result := slideLine(col)
	setColumn(&g.Board, 0, result.line)

	if g.Board[0][0] != 4 {
		t.Errorf("top cell = %d, want 4", g.Board[0][0])
	}
	if g.Board[1][0] != 0 {
		t.Errorf("second cell = %d, want 0", g.Board[1][0])
	}
	if result.score != 4 {
		t.Errorf("score = %d, want 4", result.score)
	}
}

func TestSlideDown(t *testing.T) {
	g := &Game{}
	setBoard(g, [4][4]int{
		{0, 0, 0, 0},
		{0, 0, 0, 0},
		{2, 0, 0, 0},
		{2, 0, 0, 0},
	})

	col := reverseLine(extractColumn(g.Board, 0))
	result := slideLine(col)
	setColumn(&g.Board, 0, reverseLine(result.line))

	if g.Board[3][0] != 4 {
		t.Errorf("bottom cell = %d, want 4", g.Board[3][0])
	}
	if g.Board[2][0] != 0 {
		t.Errorf("second-to-bottom = %d, want 0", g.Board[2][0])
	}
	if result.score != 4 {
		t.Errorf("score = %d, want 4", result.score)
	}
}

func TestScoring(t *testing.T) {
	g := &Game{}
	setBoard(g, [4][4]int{
		{2, 2, 4, 4},
		{0, 0, 0, 0},
		{0, 0, 0, 0},
		{0, 0, 0, 0},
	})

	// Slide left: 2+2=4 (score +4), 4+4=8 (score +8). Total = 12.
	g.Move(Left)
	if g.Score != 12 {
		t.Errorf("Score = %d, want 12", g.Score)
	}
}

func TestWinDetection(t *testing.T) {
	g := &Game{}
	setBoard(g, [4][4]int{
		{1024, 1024, 0, 0},
		{0, 0, 0, 0},
		{0, 0, 0, 0},
		{0, 0, 0, 0},
	})

	g.Move(Left)

	if !g.Won {
		t.Error("expected Won=true after reaching 2048")
	}
}

func TestGameOver(t *testing.T) {
	g := &Game{}
	// Full board, no adjacent equals.
	setBoard(g, [4][4]int{
		{2, 4, 8, 16},
		{16, 8, 4, 2},
		{2, 4, 8, 16},
		{16, 8, 4, 2},
	})

	if g.CanMove() {
		t.Error("expected CanMove()=false on full board with no merges")
	}
}

func TestCanMove(t *testing.T) {
	tests := []struct {
		name  string
		board [4][4]int
		want  bool
	}{
		{
			"empty cell exists",
			[4][4]int{
				{2, 4, 8, 16},
				{16, 8, 4, 2},
				{2, 4, 8, 16},
				{16, 8, 4, 0},
			},
			true,
		},
		{
			"adjacent horizontal equal",
			[4][4]int{
				{2, 2, 8, 16},
				{16, 8, 4, 2},
				{2, 4, 8, 16},
				{16, 8, 4, 32},
			},
			true,
		},
		{
			"adjacent vertical equal",
			[4][4]int{
				{2, 4, 8, 16},
				{16, 4, 32, 2},
				{2, 64, 8, 16},
				{16, 8, 4, 32},
			},
			true,
		},
		{
			"no moves",
			[4][4]int{
				{2, 4, 8, 16},
				{16, 8, 4, 2},
				{2, 4, 8, 16},
				{16, 8, 4, 2},
			},
			false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := &Game{}
			setBoard(g, tt.board)
			if got := g.CanMove(); got != tt.want {
				t.Errorf("CanMove() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSpawnTile(t *testing.T) {
	g := &Game{}
	// Board with exactly one empty cell.
	setBoard(g, [4][4]int{
		{2, 4, 8, 16},
		{16, 8, 4, 2},
		{2, 4, 8, 16},
		{16, 8, 4, 0},
	})

	g.SpawnTile()

	val := g.Board[3][3]
	if val != 2 && val != 4 {
		t.Errorf("spawned value = %d, want 2 or 4", val)
	}

	// Existing tiles should be unchanged.
	if g.Board[0][0] != 2 {
		t.Errorf("existing tile changed: [0][0] = %d, want 2", g.Board[0][0])
	}
}

func TestSpawnDoesNotOverwrite(t *testing.T) {
	g := &Game{}
	setBoard(g, [4][4]int{
		{2, 4, 8, 16},
		{16, 8, 4, 2},
		{2, 4, 8, 16},
		{16, 8, 4, 2},
	})

	// Full board -- spawn should be a no-op.
	g.SpawnTile()

	want := [4][4]int{
		{2, 4, 8, 16},
		{16, 8, 4, 2},
		{2, 4, 8, 16},
		{16, 8, 4, 2},
	}
	if g.Board != want {
		t.Error("SpawnTile modified a full board")
	}
}

func TestContinueAfterWin(t *testing.T) {
	g := &Game{}
	setBoard(g, [4][4]int{
		{1024, 1024, 0, 0},
		{0, 0, 0, 0},
		{0, 0, 0, 0},
		{0, 0, 0, 0},
	})

	g.Move(Left)
	if !g.Won {
		t.Fatal("expected Won=true")
	}

	g.ContinueAfterWin()
	if !g.Continued {
		t.Error("expected Continued=true")
	}

	// Further moves should not re-trigger Won.
	g.Board[0][1] = 1024
	g.Board[0][2] = 1024
	g.Move(Left)
	// Should still be playing (Continued is set).
	if g.Over {
		t.Error("game should not be over after continuing")
	}
}

func TestReset(t *testing.T) {
	g := NewGame()
	g.Score = 500
	g.Won = true
	g.Over = true

	g.Reset()

	if g.Score != 0 {
		t.Errorf("Score = %d after reset, want 0", g.Score)
	}
	if g.Won {
		t.Error("Won should be false after reset")
	}
	if g.Over {
		t.Error("Over should be false after reset")
	}

	// Should have exactly 2 tiles.
	count := 0
	for r := range boardSize {
		for c := range boardSize {
			if g.Board[r][c] != 0 {
				count++
			}
		}
	}
	if count != 2 {
		t.Errorf("tile count = %d after reset, want 2", count)
	}
}
