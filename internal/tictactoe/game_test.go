package tictactoe

import "testing"

func TestNewGame(t *testing.T) {
	g := NewGame()

	board := g.Board()
	for r := 0; r < 3; r++ {
		for c := 0; c < 3; c++ {
			if board[r][c] != Empty {
				t.Errorf("board[%d][%d] = %v, want Empty", r, c, board[r][c])
			}
		}
	}

	if g.Current() != X {
		t.Errorf("current = %v, want X", g.Current())
	}

	if g.IsOver() {
		t.Error("new game should not be over")
	}
}

func TestValidMove(t *testing.T) {
	g := NewGame()

	err := g.Move(1, 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	board := g.Board()
	if board[1][1] != X {
		t.Errorf("board[1][1] = %v, want X", board[1][1])
	}
}

func TestInvalidMoveOccupied(t *testing.T) {
	g := NewGame()

	// Place X at center
	err := g.Move(1, 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// AI moves, then it's human's turn again
	// Try to place on center (occupied by X)
	err = g.Move(1, 1)
	if err == nil {
		t.Error("expected error for occupied cell")
	}
}

func TestInvalidMoveOutOfBounds(t *testing.T) {
	g := NewGame()

	tests := []struct {
		name     string
		row, col int
	}{
		{"negative row", -1, 0},
		{"negative col", 0, -1},
		{"row too large", 3, 0},
		{"col too large", 0, 3},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := g.Move(tt.row, tt.col)
			if err == nil {
				t.Errorf("Move(%d, %d) should return error", tt.row, tt.col)
			}
		})
	}
}

func TestHorizontalWin(t *testing.T) {
	tests := []struct {
		name string
		row  int
	}{
		{"top row", 0},
		{"middle row", 1},
		{"bottom row", 2},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewGame()
			// Force a horizontal win for X by setting board directly
			g.board[tt.row][0] = X
			g.board[tt.row][1] = X
			g.board[tt.row][2] = X
			g.updateState()

			if g.Winner() != X {
				t.Errorf("Winner() = %v, want X", g.Winner())
			}
			if g.GameState() != Won {
				t.Errorf("GameState() = %v, want Won", g.GameState())
			}
		})
	}
}

func TestVerticalWin(t *testing.T) {
	tests := []struct {
		name string
		col  int
	}{
		{"left column", 0},
		{"center column", 1},
		{"right column", 2},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewGame()
			g.board[0][tt.col] = O
			g.board[1][tt.col] = O
			g.board[2][tt.col] = O
			g.updateState()

			if g.Winner() != O {
				t.Errorf("Winner() = %v, want O", g.Winner())
			}
			if g.GameState() != Lost {
				t.Errorf("GameState() = %v, want Lost", g.GameState())
			}
		})
	}
}

func TestDiagonalWin(t *testing.T) {
	t.Run("main diagonal", func(t *testing.T) {
		g := NewGame()
		g.board[0][0] = X
		g.board[1][1] = X
		g.board[2][2] = X
		g.updateState()

		if g.Winner() != X {
			t.Errorf("Winner() = %v, want X", g.Winner())
		}
	})

	t.Run("anti diagonal", func(t *testing.T) {
		g := NewGame()
		g.board[0][2] = X
		g.board[1][1] = X
		g.board[2][0] = X
		g.updateState()

		if g.Winner() != X {
			t.Errorf("Winner() = %v, want X", g.Winner())
		}
	})
}

func TestDraw(t *testing.T) {
	g := NewGame()
	// Set up a drawn board:
	// X | O | X
	// X | O | O
	// O | X | X
	g.board = [3][3]Cell{
		{X, O, X},
		{X, O, O},
		{O, X, X},
	}
	g.updateState()

	if !g.IsDraw() {
		t.Error("expected draw")
	}
	if g.Winner() != Empty {
		t.Errorf("Winner() = %v, want Empty", g.Winner())
	}
	if g.GameState() != Draw {
		t.Errorf("GameState() = %v, want Draw", g.GameState())
	}
}

func TestAIMoveExists(t *testing.T) {
	g := NewGame()

	// Human moves first
	err := g.Move(0, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// After human moves, AI should have responded (Move triggers turn switch)
	// AI's turn now -- call AIMove
	g.AIMove()

	// Count non-empty cells: should be 2 (one X, one O)
	board := g.Board()
	count := 0
	for r := 0; r < 3; r++ {
		for c := 0; c < 3; c++ {
			if board[r][c] != Empty {
				count++
			}
		}
	}
	if count != 2 {
		t.Errorf("expected 2 pieces on board, got %d", count)
	}
}

func TestAIBlocksWin(t *testing.T) {
	g := NewGame()
	// Set up board where X has two in a row and AI must block:
	// X | X | _
	// O | _ | _
	// _ | _ | _
	g.board[0][0] = X
	g.board[0][1] = X
	g.board[1][0] = O
	g.current = O
	g.state = Playing

	g.AIMove()

	board := g.Board()
	// AI must place O at (0,2) to block X's horizontal win
	if board[0][2] != O {
		t.Errorf("AI should block at (0,2), got board:\n%s", boardString(board))
	}
}

func TestMoveAfterGameOver(t *testing.T) {
	g := NewGame()
	g.board[0][0] = X
	g.board[0][1] = X
	g.board[0][2] = X
	g.updateState()

	err := g.Move(1, 0)
	if err == nil {
		t.Error("expected error for move after game over")
	}
}

func TestAIMoveDoesNothingWhenGameOver(t *testing.T) {
	g := NewGame()
	g.board[0][0] = X
	g.board[0][1] = X
	g.board[0][2] = X
	g.updateState()

	boardBefore := g.Board()
	g.AIMove()
	boardAfter := g.Board()

	if boardBefore != boardAfter {
		t.Error("AI should not move when game is over")
	}
}

// boardString returns a simple text representation for test failure messages.
func boardString(board [3][3]Cell) string {
	var rows []string
	for r := 0; r < 3; r++ {
		row := ""
		for c := 0; c < 3; c++ {
			switch board[r][c] {
			case X:
				row += "X"
			case O:
				row += "O"
			default:
				row += "."
			}
			if c < 2 {
				row += "|"
			}
		}
		rows = append(rows, row)
	}
	return rows[0] + "\n" + rows[1] + "\n" + rows[2]
}
