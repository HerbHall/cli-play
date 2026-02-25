package connectfour

import "testing"

func TestNewGame(t *testing.T) {
	g := NewGame()

	board := g.Board()
	for r := 0; r < Rows; r++ {
		for c := 0; c < Cols; c++ {
			if board[r][c] != Empty {
				t.Errorf("board[%d][%d] = %v, want Empty", r, c, board[r][c])
			}
		}
	}

	if g.Current() != Red {
		t.Errorf("current = %v, want Red", g.Current())
	}

	if g.IsOver() {
		t.Error("new game should not be over")
	}
}

func TestDropDiscGravity(t *testing.T) {
	g := NewGame()

	// First disc in column 3 should land at the bottom row
	row, err := g.DropDisc(3)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if row != Rows-1 {
		t.Errorf("first disc row = %d, want %d", row, Rows-1)
	}

	board := g.Board()
	if board[Rows-1][3] != Red {
		t.Errorf("board[%d][3] = %v, want Red", Rows-1, board[Rows-1][3])
	}

	// Second disc (Yellow/AI turn, but we test via DropDisc directly)
	// Switch back to test: drop another Red after AI turn
	// AI will move automatically in the model, but for game logic test,
	// drop as Yellow (current player after Red's move)
	row, err = g.DropDisc(3) // Yellow drops in same column
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if row != Rows-2 {
		t.Errorf("second disc row = %d, want %d", row, Rows-2)
	}

	board = g.Board()
	if board[Rows-2][3] != Yellow {
		t.Errorf("board[%d][3] = %v, want Yellow", Rows-2, board[Rows-2][3])
	}
}

func TestDropDiscFullColumn(t *testing.T) {
	g := NewGame()

	// Fill column 0 completely (alternating Red/Yellow)
	for i := 0; i < Rows; i++ {
		_, err := g.DropDisc(0)
		if err != nil {
			t.Fatalf("drop %d failed: %v", i, err)
		}
		if g.IsOver() {
			// If someone wins before column fills, that's fine for this test
			return
		}
	}

	// Column is now full; next drop should fail
	_, err := g.DropDisc(0)
	if err == nil {
		t.Error("expected error for full column")
	}
}

func TestDropDiscOutOfBounds(t *testing.T) {
	g := NewGame()

	tests := []struct {
		name string
		col  int
	}{
		{"negative", -1},
		{"too large", Cols},
		{"way too large", 100},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := g.DropDisc(tt.col)
			if err == nil {
				t.Errorf("DropDisc(%d) should return error", tt.col)
			}
		})
	}
}

func TestHorizontalWin(t *testing.T) {
	g := NewGame()

	// Set up a horizontal 4-in-a-row for Red at bottom row
	g.board[Rows-1][0] = Red
	g.board[Rows-1][1] = Red
	g.board[Rows-1][2] = Red
	g.board[Rows-1][3] = Red
	g.updateState()

	if g.Winner() != Red {
		t.Errorf("Winner() = %v, want Red", g.Winner())
	}
	if g.GameState() != Won {
		t.Errorf("GameState() = %v, want Won", g.GameState())
	}
}

func TestVerticalWin(t *testing.T) {
	g := NewGame()

	// Set up a vertical 4-in-a-row for Yellow in column 2
	g.board[Rows-1][2] = Yellow
	g.board[Rows-2][2] = Yellow
	g.board[Rows-3][2] = Yellow
	g.board[Rows-4][2] = Yellow
	g.updateState()

	if g.Winner() != Yellow {
		t.Errorf("Winner() = %v, want Yellow", g.Winner())
	}
	if g.GameState() != Lost {
		t.Errorf("GameState() = %v, want Lost", g.GameState())
	}
}

func TestDiagonalWinDownRight(t *testing.T) {
	g := NewGame()

	// Diagonal from top-left to bottom-right
	g.board[0][0] = Red
	g.board[1][1] = Red
	g.board[2][2] = Red
	g.board[3][3] = Red
	g.updateState()

	if g.Winner() != Red {
		t.Errorf("Winner() = %v, want Red", g.Winner())
	}
	if g.GameState() != Won {
		t.Errorf("GameState() = %v, want Won", g.GameState())
	}
}

func TestDiagonalWinDownLeft(t *testing.T) {
	g := NewGame()

	// Diagonal from top-right to bottom-left
	g.board[0][6] = Yellow
	g.board[1][5] = Yellow
	g.board[2][4] = Yellow
	g.board[3][3] = Yellow
	g.updateState()

	if g.Winner() != Yellow {
		t.Errorf("Winner() = %v, want Yellow", g.Winner())
	}
	if g.GameState() != Lost {
		t.Errorf("GameState() = %v, want Lost", g.GameState())
	}
}

func TestDrawDetection(t *testing.T) {
	g := NewGame()

	// Fill the board in a pattern that produces no winner.
	// Column pattern: R Y R R Y R Y (repeated with offset per row pair)
	pattern := [Rows][Cols]Cell{
		{Red, Yellow, Red, Yellow, Red, Yellow, Red},
		{Red, Yellow, Red, Yellow, Red, Yellow, Red},
		{Yellow, Red, Yellow, Red, Yellow, Red, Yellow},
		{Red, Yellow, Red, Yellow, Red, Yellow, Red},
		{Red, Yellow, Red, Yellow, Red, Yellow, Red},
		{Yellow, Red, Yellow, Red, Yellow, Red, Yellow},
	}
	g.board = pattern

	// Verify no winner exists in this pattern
	if checkWinner(g.board) != Empty {
		t.Fatal("test pattern has an unexpected winner; adjust the pattern")
	}

	g.updateState()

	if g.GameState() != Draw {
		t.Errorf("GameState() = %v, want Draw", g.GameState())
	}
}

func TestAIBlocksPlayerWin(t *testing.T) {
	g := NewGame()

	// Red has 3 in a row at bottom, col 0-2. Column 3 is open.
	// AI (Yellow) must block at column 3.
	g.board[Rows-1][0] = Red
	g.board[Rows-1][1] = Red
	g.board[Rows-1][2] = Red
	// Put some Yellow discs elsewhere so the board is plausible
	g.board[Rows-1][4] = Yellow
	g.board[Rows-1][5] = Yellow
	g.board[Rows-2][4] = Yellow
	g.current = Yellow

	col := g.AIMove()
	if col != 3 {
		t.Errorf("AI chose column %d, want 3 (block player win)", col)
	}

	board := g.Board()
	if board[Rows-1][3] != Yellow {
		t.Errorf("board[%d][3] = %v, want Yellow", Rows-1, board[Rows-1][3])
	}
}

func TestAITakesWinningMove(t *testing.T) {
	g := NewGame()

	// Yellow has 3 in a row at bottom, col 1-3. Column 4 and 0 are open.
	// AI should take the win at column 4 (or 0).
	g.board[Rows-1][1] = Yellow
	g.board[Rows-1][2] = Yellow
	g.board[Rows-1][3] = Yellow
	// Red has some discs to make the board plausible
	g.board[Rows-1][5] = Red
	g.board[Rows-1][6] = Red
	g.board[Rows-2][5] = Red
	g.current = Yellow

	g.AIMove()

	// AI should have won
	if g.GameState() != Lost {
		board := g.Board()
		t.Errorf("AI should have won; GameState() = %v\n%s", g.GameState(), boardString(board))
	}
}

func TestAIPrefersCenter(t *testing.T) {
	g := NewGame()

	// Empty board, Yellow's turn. AI should prefer column 3 (center).
	g.current = Yellow

	col := g.AIMove()
	if col != 3 {
		t.Errorf("AI chose column %d on empty board, want 3 (center)", col)
	}
}

func TestDropDiscAfterGameOver(t *testing.T) {
	g := NewGame()

	g.board[Rows-1][0] = Red
	g.board[Rows-1][1] = Red
	g.board[Rows-1][2] = Red
	g.board[Rows-1][3] = Red
	g.updateState()

	_, err := g.DropDisc(4)
	if err == nil {
		t.Error("expected error for drop after game over")
	}
}

func TestAIMoveDoesNothingWhenGameOver(t *testing.T) {
	g := NewGame()

	g.board[Rows-1][0] = Red
	g.board[Rows-1][1] = Red
	g.board[Rows-1][2] = Red
	g.board[Rows-1][3] = Red
	g.updateState()

	boardBefore := g.Board()
	g.AIMove()
	boardAfter := g.Board()

	if boardBefore != boardAfter {
		t.Error("AI should not move when game is over")
	}
}

func TestWinnerReturnsEmpty(t *testing.T) {
	g := NewGame()

	if g.Winner() != Empty {
		t.Errorf("Winner() = %v on new game, want Empty", g.Winner())
	}
}

// boardString returns a text representation for test failure messages.
func boardString(board [Rows][Cols]Cell) string {
	rows := make([]string, 0, Rows)
	for r := 0; r < Rows; r++ {
		row := ""
		for c := 0; c < Cols; c++ {
			row += board[r][c].String()
			if c < Cols-1 {
				row += "|"
			}
		}
		rows = append(rows, row)
	}
	result := ""
	for i, row := range rows {
		result += row
		if i < len(rows)-1 {
			result += "\n"
		}
	}
	return result
}
