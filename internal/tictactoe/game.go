package tictactoe

import "errors"

// Cell represents the state of a board position.
type Cell int

const (
	Empty Cell = iota
	X          // Human player
	O          // AI player
)

// String returns a display character for the cell.
func (c Cell) String() string {
	switch c {
	case X:
		return "X"
	case O:
		return "O"
	default:
		return " "
	}
}

// State represents the current game state.
type State int

const (
	Playing State = iota
	Won
	Lost
	Draw
)

// Game holds the complete state of a Tic-Tac-Toe game.
type Game struct {
	board   [3][3]Cell
	current Cell
	state   State
}

// NewGame creates a fresh game with an empty board. X (human) goes first.
func NewGame() *Game {
	return &Game{
		current: X,
		state:   Playing,
	}
}

// Board returns a copy of the current board.
func (g *Game) Board() [3][3]Cell {
	return g.board
}

// Current returns whose turn it is.
func (g *Game) Current() Cell {
	return g.current
}

// GameState returns the current game state.
func (g *Game) GameState() State {
	return g.state
}

// IsOver returns true if the game has ended.
func (g *Game) IsOver() bool {
	return g.state != Playing
}

// Winner returns the winning cell (X or O), or Empty if no winner yet.
func (g *Game) Winner() Cell {
	return checkWinner(g.board)
}

// IsDraw returns true if the board is full with no winner.
func (g *Game) IsDraw() bool {
	if checkWinner(g.board) != Empty {
		return false
	}
	for r := 0; r < 3; r++ {
		for c := 0; c < 3; c++ {
			if g.board[r][c] == Empty {
				return false
			}
		}
	}
	return true
}

// Move places the human player's mark (X) at the given position.
func (g *Game) Move(row, col int) error {
	if g.IsOver() {
		return errors.New("game is over")
	}
	if row < 0 || row > 2 || col < 0 || col > 2 {
		return errors.New("position out of bounds")
	}
	if g.board[row][col] != Empty {
		return errors.New("cell is occupied")
	}
	if g.current != X {
		return errors.New("not human's turn")
	}

	g.board[row][col] = X
	g.updateState()
	if !g.IsOver() {
		g.current = O
	}
	return nil
}

// AIMove makes the AI (O) play using minimax. It selects the optimal move.
func (g *Game) AIMove() {
	if g.IsOver() || g.current != O {
		return
	}

	bestScore := -1000
	bestRow, bestCol := -1, -1

	for r := 0; r < 3; r++ {
		for c := 0; c < 3; c++ {
			if g.board[r][c] == Empty {
				g.board[r][c] = O
				score := minimax(g.board, false)
				g.board[r][c] = Empty
				if score > bestScore {
					bestScore = score
					bestRow = r
					bestCol = c
				}
			}
		}
	}

	if bestRow >= 0 {
		g.board[bestRow][bestCol] = O
		g.updateState()
		if !g.IsOver() {
			g.current = X
		}
	}
}

func (g *Game) updateState() {
	winner := checkWinner(g.board)
	switch winner {
	case X:
		g.state = Won
	case O:
		g.state = Lost
	default:
		if g.IsDraw() {
			g.state = Draw
		}
	}
}

// checkWinner returns X or O if there is a winner, or Empty otherwise.
func checkWinner(board [3][3]Cell) Cell {
	// Check rows
	for r := 0; r < 3; r++ {
		if board[r][0] != Empty && board[r][0] == board[r][1] && board[r][1] == board[r][2] {
			return board[r][0]
		}
	}
	// Check columns
	for c := 0; c < 3; c++ {
		if board[0][c] != Empty && board[0][c] == board[1][c] && board[1][c] == board[2][c] {
			return board[0][c]
		}
	}
	// Check diagonals
	if board[0][0] != Empty && board[0][0] == board[1][1] && board[1][1] == board[2][2] {
		return board[0][0]
	}
	if board[0][2] != Empty && board[0][2] == board[1][1] && board[1][1] == board[2][0] {
		return board[0][2]
	}
	return Empty
}

// minimax evaluates all positions and returns the best score for the current player.
// isMaximizing=true means it's O's turn (AI), false means X's turn (human).
func minimax(board [3][3]Cell, isMaximizing bool) int {
	winner := checkWinner(board)
	if winner == O {
		return 1
	}
	if winner == X {
		return -1
	}
	if isBoardFull(board) {
		return 0
	}

	if isMaximizing {
		best := -1000
		for r := 0; r < 3; r++ {
			for c := 0; c < 3; c++ {
				if board[r][c] == Empty {
					board[r][c] = O
					score := minimax(board, false)
					board[r][c] = Empty
					if score > best {
						best = score
					}
				}
			}
		}
		return best
	}

	best := 1000
	for r := 0; r < 3; r++ {
		for c := 0; c < 3; c++ {
			if board[r][c] == Empty {
				board[r][c] = X
				score := minimax(board, true)
				board[r][c] = Empty
				if score < best {
					best = score
				}
			}
		}
	}
	return best
}

func isBoardFull(board [3][3]Cell) bool {
	for r := 0; r < 3; r++ {
		for c := 0; c < 3; c++ {
			if board[r][c] == Empty {
				return false
			}
		}
	}
	return true
}
