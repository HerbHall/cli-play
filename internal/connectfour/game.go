package connectfour

import "errors"

const (
	Rows = 6
	Cols = 7
)

// Cell represents the state of a board position.
type Cell int

const (
	Empty  Cell = iota
	Red         // Human player
	Yellow      // AI player
)

// String returns a display character for the cell.
func (c Cell) String() string {
	switch c {
	case Red:
		return "R"
	case Yellow:
		return "Y"
	default:
		return "."
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

// Game holds the complete state of a Connect Four game.
type Game struct {
	board   [Rows][Cols]Cell
	current Cell
	state   State
}

// NewGame creates a fresh game. Red (human) goes first.
func NewGame() *Game {
	return &Game{
		current: Red,
		state:   Playing,
	}
}

// Board returns a copy of the current board.
func (g *Game) Board() [Rows][Cols]Cell {
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

// DropDisc places the current player's disc in the given column.
// The disc falls to the lowest empty row. Returns the row it landed on.
func (g *Game) DropDisc(col int) (int, error) {
	if g.IsOver() {
		return -1, errors.New("game is over")
	}
	if col < 0 || col >= Cols {
		return -1, errors.New("column out of bounds")
	}
	if g.board[0][col] != Empty {
		return -1, errors.New("column is full")
	}

	row := g.lowestEmptyRow(col)
	g.board[row][col] = g.current
	g.updateState()

	if !g.IsOver() {
		g.switchTurn()
	}
	return row, nil
}

// AIMove makes the AI (Yellow) play using a heuristic strategy.
// Returns the column chosen.
func (g *Game) AIMove() int {
	if g.IsOver() || g.current != Yellow {
		return -1
	}

	col := g.bestAIColumn()
	if col < 0 {
		return -1
	}

	row := g.lowestEmptyRow(col)
	g.board[row][col] = Yellow
	g.updateState()

	if !g.IsOver() {
		g.switchTurn()
	}
	return col
}

// bestAIColumn selects the best column for the AI using a priority strategy:
// 1. Win if possible
// 2. Block opponent's winning move
// 3. Prefer center, then adjacent columns
// 4. Avoid moves that give opponent a win next turn
func (g *Game) bestAIColumn() int {
	// 1. Check for winning move
	for c := 0; c < Cols; c++ {
		if g.canDrop(c) {
			r := g.lowestEmptyRow(c)
			g.board[r][c] = Yellow
			if checkWinner(g.board) == Yellow {
				g.board[r][c] = Empty
				return c
			}
			g.board[r][c] = Empty
		}
	}

	// 2. Block opponent's winning move
	for c := 0; c < Cols; c++ {
		if g.canDrop(c) {
			r := g.lowestEmptyRow(c)
			g.board[r][c] = Red
			if checkWinner(g.board) == Red {
				g.board[r][c] = Empty
				return c
			}
			g.board[r][c] = Empty
		}
	}

	// 3. Prefer center column, then adjacent, avoiding moves that give opponent a win
	preferred := [Cols]int{3, 2, 4, 1, 5, 0, 6}
	fallback := -1

	for _, c := range preferred {
		if !g.canDrop(c) {
			continue
		}
		if !g.givesOpponentWin(c) {
			return c
		}
		if fallback < 0 {
			fallback = c
		}
	}

	// 4. All moves give opponent a win; pick the first available
	if fallback >= 0 {
		return fallback
	}
	return -1
}

// givesOpponentWin returns true if dropping Yellow in col allows Red to win
// on the next move.
func (g *Game) givesOpponentWin(col int) bool {
	r := g.lowestEmptyRow(col)
	g.board[r][col] = Yellow

	for c := 0; c < Cols; c++ {
		if g.board[0][c] != Empty && (r != 0 || c != col) {
			continue
		}
		opRow := g.lowestEmptyRow(c)
		if opRow < 0 {
			continue
		}
		g.board[opRow][c] = Red
		wins := checkWinner(g.board) == Red
		g.board[opRow][c] = Empty
		if wins {
			g.board[r][col] = Empty
			return true
		}
	}

	g.board[r][col] = Empty
	return false
}

func (g *Game) canDrop(col int) bool {
	return col >= 0 && col < Cols && g.board[0][col] == Empty
}

func (g *Game) lowestEmptyRow(col int) int {
	for r := Rows - 1; r >= 0; r-- {
		if g.board[r][col] == Empty {
			return r
		}
	}
	return -1
}

func (g *Game) switchTurn() {
	switch g.current {
	case Red:
		g.current = Yellow
	case Yellow:
		g.current = Red
	}
}

func (g *Game) updateState() {
	winner := checkWinner(g.board)
	switch winner {
	case Red:
		g.state = Won
	case Yellow:
		g.state = Lost
	default:
		if g.isBoardFull() {
			g.state = Draw
		}
	}
}

func (g *Game) isBoardFull() bool {
	for c := 0; c < Cols; c++ {
		if g.board[0][c] == Empty {
			return false
		}
	}
	return true
}

// checkWinner checks all possible lines of 4 on the board.
// Returns Red or Yellow if there is a winner, or Empty otherwise.
func checkWinner(board [Rows][Cols]Cell) Cell {
	// Horizontal
	for r := 0; r < Rows; r++ {
		for c := 0; c <= Cols-4; c++ {
			if board[r][c] != Empty &&
				board[r][c] == board[r][c+1] &&
				board[r][c+1] == board[r][c+2] &&
				board[r][c+2] == board[r][c+3] {
				return board[r][c]
			}
		}
	}

	// Vertical
	for c := 0; c < Cols; c++ {
		for r := 0; r <= Rows-4; r++ {
			if board[r][c] != Empty &&
				board[r][c] == board[r+1][c] &&
				board[r+1][c] == board[r+2][c] &&
				board[r+2][c] == board[r+3][c] {
				return board[r][c]
			}
		}
	}

	// Diagonal down-right
	for r := 0; r <= Rows-4; r++ {
		for c := 0; c <= Cols-4; c++ {
			if board[r][c] != Empty &&
				board[r][c] == board[r+1][c+1] &&
				board[r+1][c+1] == board[r+2][c+2] &&
				board[r+2][c+2] == board[r+3][c+3] {
				return board[r][c]
			}
		}
	}

	// Diagonal down-left
	for r := 0; r <= Rows-4; r++ {
		for c := 3; c < Cols; c++ {
			if board[r][c] != Empty &&
				board[r][c] == board[r+1][c-1] &&
				board[r+1][c-1] == board[r+2][c-2] &&
				board[r+2][c-2] == board[r+3][c-3] {
				return board[r][c]
			}
		}
	}

	return Empty
}

// Winner returns the winning cell, or Empty if no winner yet.
func (g *Game) Winner() Cell {
	return checkWinner(g.board)
}
