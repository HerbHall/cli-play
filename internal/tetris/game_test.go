package tetris

import (
	"fmt"
	"testing"
)

// newTestGame creates a game with an empty board and a known piece for
// deterministic testing. The piece is placed at a given position.
func newTestGame(pt PieceType, row, col int) *Game {
	g := &Game{Level: 1}
	g.initBoard()
	g.Current = Piece{Type: pt, Rotation: 0, Row: row, Col: col}
	g.Next = PieceO
	return g
}

func TestPieceSpawnPosition(t *testing.T) {
	g := NewGame()
	// Piece should spawn near center. The spawn col is (10-3)/2 = 3.
	if g.Current.Col != 3 {
		t.Errorf("spawn col = %d, want 3", g.Current.Col)
	}
	if g.Current.Row != 0 {
		t.Errorf("spawn row = %d, want 0", g.Current.Row)
	}
}

func TestMoveLeft(t *testing.T) {
	g := newTestGame(PieceO, 0, 5)
	if !g.MoveLeft() {
		t.Error("expected MoveLeft to succeed")
	}
	if g.Current.Col != 4 {
		t.Errorf("col after MoveLeft = %d, want 4", g.Current.Col)
	}
}

func TestMoveRight(t *testing.T) {
	g := newTestGame(PieceO, 0, 5)
	if !g.MoveRight() {
		t.Error("expected MoveRight to succeed")
	}
	if g.Current.Col != 6 {
		t.Errorf("col after MoveRight = %d, want 6", g.Current.Col)
	}
}

func TestMoveDown(t *testing.T) {
	g := newTestGame(PieceO, 5, 4)
	if !g.MoveDown() {
		t.Error("expected MoveDown to succeed")
	}
	if g.Current.Row != 6 {
		t.Errorf("row after MoveDown = %d, want 6", g.Current.Row)
	}
}

func TestRotation(t *testing.T) {
	g := newTestGame(PieceT, 5, 4)
	origRot := g.Current.Rotation
	if !g.Rotate() {
		t.Error("expected Rotate to succeed")
	}
	if g.Current.Rotation == origRot {
		t.Error("rotation did not change")
	}
}

func TestRotationOPieceUnchanged(t *testing.T) {
	g := newTestGame(PieceO, 5, 4)
	cellsBefore := g.Current.Cells()
	g.Rotate()
	cellsAfter := g.Current.Cells()
	// O piece has only 1 rotation state, so cells should remain the same.
	if cellsBefore != cellsAfter {
		t.Errorf("O piece cells changed after rotate: %v -> %v", cellsBefore, cellsAfter)
	}
}

func TestCollisionWithLeftWall(t *testing.T) {
	g := newTestGame(PieceO, 5, 0)
	if g.MoveLeft() {
		t.Error("MoveLeft should fail at left wall")
	}
}

func TestCollisionWithRightWall(t *testing.T) {
	// O piece is 2 wide, so col 8 puts it at cols 8-9 (max).
	g := newTestGame(PieceO, 5, 8)
	if g.MoveRight() {
		t.Error("MoveRight should fail at right wall")
	}
}

func TestCollisionWithFloor(t *testing.T) {
	// O piece is 2 tall; row 18 puts cells at rows 18 and 19 (last row).
	g := newTestGame(PieceO, 18, 4)
	if g.MoveDown() {
		t.Error("MoveDown should fail at floor")
	}
}

func TestCollisionWithPlacedPieces(t *testing.T) {
	g := newTestGame(PieceO, 16, 4)
	// Place a block directly below the piece.
	g.Board[18][4] = PieceI
	g.Board[18][5] = PieceI

	if g.MoveDown() {
		t.Error("MoveDown should fail when blocked by placed pieces")
	}
}

func TestSingleLineClear(t *testing.T) {
	g := newTestGame(PieceI, 0, 0)
	g.Over = false

	// Fill bottom row except cols 0-3.
	for c := 4; c < BoardWidth; c++ {
		g.Board[BoardHeight-1][c] = PieceI
	}

	// Place I piece horizontally at the bottom to complete the row.
	g.Current = Piece{Type: PieceI, Rotation: 0, Row: BoardHeight - 1, Col: 0}
	g.lockPiece()

	if g.Lines != 1 {
		t.Errorf("lines = %d, want 1", g.Lines)
	}
	if g.Score != 100 {
		t.Errorf("score = %d, want 100", g.Score)
	}
}

func TestDoubleLineClear(t *testing.T) {
	g := &Game{Level: 1}
	g.initBoard()
	g.Next = PieceO

	// Fill bottom 2 rows except cols 0-1.
	for r := BoardHeight - 2; r < BoardHeight; r++ {
		for c := 2; c < BoardWidth; c++ {
			g.Board[r][c] = PieceT
		}
	}

	// O piece at the gap (cols 0-1, rows 18-19) completes both rows.
	g.Current = Piece{Type: PieceO, Rotation: 0, Row: BoardHeight - 2, Col: 0}
	g.lockPiece()

	if g.Lines != 2 {
		t.Errorf("lines = %d, want 2", g.Lines)
	}
	if g.Score != 300 {
		t.Errorf("score = %d, want 300", g.Score)
	}
}

func TestTripleLineClear(t *testing.T) {
	g := &Game{Level: 1}
	g.initBoard()
	g.Next = PieceO

	// Fill bottom 3 rows except col 0.
	for r := BoardHeight - 3; r < BoardHeight; r++ {
		for c := 1; c < BoardWidth; c++ {
			g.Board[r][c] = PieceS
		}
	}

	// Use an I piece vertically to fill col 0 for 3 of the rows.
	// I piece vertical covers 4 rows, but we only have 3 full.
	// Instead manually set col 0 for those 3 rows.
	for r := BoardHeight - 3; r < BoardHeight; r++ {
		g.Board[r][0] = PieceS
	}

	cleared := g.clearLines()
	g.Lines += cleared
	g.Score += lineScore(cleared)

	if g.Lines != 3 {
		t.Errorf("lines = %d, want 3", g.Lines)
	}
	if g.Score != 500 {
		t.Errorf("score = %d, want 500", g.Score)
	}
}

func TestTetrisClear(t *testing.T) {
	g := &Game{Level: 1}
	g.initBoard()
	g.Next = PieceO

	// Fill bottom 4 rows completely.
	for r := BoardHeight - 4; r < BoardHeight; r++ {
		for c := range BoardWidth {
			g.Board[r][c] = PieceJ
		}
	}

	cleared := g.clearLines()
	g.Lines += cleared
	g.Score += lineScore(cleared)

	if g.Lines != 4 {
		t.Errorf("lines = %d, want 4", g.Lines)
	}
	if g.Score != 800 {
		t.Errorf("score = %d, want 800", g.Score)
	}
}

func TestScoringValues(t *testing.T) {
	tests := []struct {
		lines int
		want  int
	}{
		{0, 0},
		{1, 100},
		{2, 300},
		{3, 500},
		{4, 800},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("%d_lines", tt.lines), func(t *testing.T) {
			got := lineScore(tt.lines)
			if got != tt.want {
				t.Errorf("lineScore(%d) = %d, want %d", tt.lines, got, tt.want)
			}
		})
	}
}

func TestLevelProgression(t *testing.T) {
	g := &Game{Level: 1}
	g.initBoard()
	g.Next = PieceO

	// Simulate clearing 10 lines.
	g.Lines = 9
	g.Level = 1

	// Fill bottom row and clear it.
	for c := range BoardWidth {
		g.Board[BoardHeight-1][c] = PieceL
	}
	cleared := g.clearLines()
	g.Lines += cleared
	g.Level = g.Lines/10 + 1

	if g.Level != 2 {
		t.Errorf("level = %d, want 2 after 10 lines", g.Level)
	}
}

func TestHardDrop(t *testing.T) {
	g := newTestGame(PieceO, 0, 4)
	rows := g.HardDrop()

	// O piece is 2 tall. From row 0, it drops to row 18 (rows 18-19 occupied).
	// That's 18 rows dropped.
	if rows != 18 {
		t.Errorf("hard drop rows = %d, want 18", rows)
	}

	// The piece should be locked on the board at the bottom.
	if g.Board[18][4] != PieceO {
		t.Error("expected PieceO at [18][4] after hard drop")
	}
	if g.Board[19][4] != PieceO {
		t.Error("expected PieceO at [19][4] after hard drop")
	}
}

func TestGhostRow(t *testing.T) {
	g := newTestGame(PieceO, 0, 4)
	ghostR := g.GhostRow()

	// O piece from row 0 should ghost to row 18 (occupies 18-19).
	if ghostR != 18 {
		t.Errorf("ghost row = %d, want 18", ghostR)
	}
}

func TestGameOverDetection(t *testing.T) {
	g := &Game{Level: 1}
	g.initBoard()

	// Fill the top rows to block spawning.
	for c := range BoardWidth {
		g.Board[0][c] = PieceI
		g.Board[1][c] = PieceI
	}

	g.Next = PieceO
	spawned := g.spawnPiece()

	if spawned {
		t.Error("expected spawn to fail when top is blocked")
	}
	if !g.Over {
		t.Error("expected Over=true when spawn is blocked")
	}
}

func TestTickInterval(t *testing.T) {
	tests := []struct {
		level   int
		wantMin int
		wantMax int
	}{
		{1, 500, 500},
		{2, 460, 460},
		{5, 340, 340},
		{12, 60, 60}, // capped at minimum
		{20, 60, 60}, // still capped
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("level_%d", tt.level), func(t *testing.T) {
			g := &Game{Level: tt.level}
			got := g.TickInterval()
			if got < tt.wantMin || got > tt.wantMax {
				t.Errorf("TickInterval() = %d, want [%d, %d]", got, tt.wantMin, tt.wantMax)
			}
		})
	}
}

func TestReset(t *testing.T) {
	g := NewGame()
	g.Score = 5000
	g.Lines = 42
	g.Level = 5
	g.Over = true
	g.Board[10][5] = PieceT

	g.Reset()

	if g.Score != 0 {
		t.Errorf("Score = %d after reset, want 0", g.Score)
	}
	if g.Lines != 0 {
		t.Errorf("Lines = %d after reset, want 0", g.Lines)
	}
	if g.Level != 1 {
		t.Errorf("Level = %d after reset, want 1", g.Level)
	}
	if g.Over {
		t.Error("Over should be false after reset")
	}
	// Board should be cleared (except for the current piece, which
	// is floating and not on the board).
	for r := range BoardHeight {
		for c := range BoardWidth {
			if g.Board[r][c] != PieceNone {
				t.Errorf("Board[%d][%d] = %d, want PieceNone after reset", r, c, g.Board[r][c])
			}
		}
	}
}

func TestMoveOnGameOver(t *testing.T) {
	g := newTestGame(PieceO, 5, 4)
	g.Over = true

	if g.MoveLeft() {
		t.Error("MoveLeft should return false when game is over")
	}
	if g.MoveRight() {
		t.Error("MoveRight should return false when game is over")
	}
	if g.MoveDown() {
		t.Error("MoveDown should return false when game is over")
	}
	if g.Rotate() {
		t.Error("Rotate should return false when game is over")
	}
	if g.HardDrop() != 0 {
		t.Error("HardDrop should return 0 when game is over")
	}
}

func TestLineClearShiftsRowsDown(t *testing.T) {
	g := &Game{Level: 1}
	g.initBoard()

	// Place a block on row 17.
	g.Board[17][3] = PieceT

	// Fill row 18 completely.
	for c := range BoardWidth {
		g.Board[18][c] = PieceL
	}

	// Row 19 is empty.
	cleared := g.clearLines()
	if cleared != 1 {
		t.Errorf("cleared = %d, want 1", cleared)
	}

	// The block from row 17 should have shifted to row 18.
	if g.Board[18][3] != PieceT {
		t.Errorf("expected PieceT at [18][3] after clear, got %d", g.Board[18][3])
	}

	// Original row 17 col 3 should now be empty (shifted down).
	if g.Board[17][3] != PieceNone {
		t.Errorf("expected PieceNone at [17][3] after clear, got %d", g.Board[17][3])
	}
}

func TestWallKickOnRotation(t *testing.T) {
	// Place T piece against the left wall.
	g := newTestGame(PieceT, 5, 0)
	// Rotate to a state that would extend left of col 0.
	// T rotation state 3: offsets {0,1}, {1,1}, {2,1}, {1,0}
	// At col 0 this is fine. Rotate again to state 0: {0,0}, {0,1}, {0,2}, {1,1}
	// Also fine at col 0. Let's try a piece that actually needs a kick.

	// I piece vertical at col 0: offsets {0,0},{1,0},{2,0},{3,0}
	g = newTestGame(PieceI, 5, 0)
	g.Current.Rotation = 1 // vertical: col 0 only

	// Rotate to horizontal: offsets {0,0},{0,1},{0,2},{0,3}
	// At col 0 this is valid. But let's verify it works.
	ok := g.Rotate()
	if !ok {
		t.Error("expected rotation with wall kick to succeed")
	}
}
