package sudoku

import (
	"errors"
	"math/rand/v2"
)

// Difficulty controls how many given cells a puzzle starts with.
type Difficulty int

const (
	Easy Difficulty = iota
	Medium
	Hard
)

// String returns the display name of the difficulty.
func (d Difficulty) String() string {
	switch d {
	case Easy:
		return "Easy"
	case Medium:
		return "Medium"
	case Hard:
		return "Hard"
	}
	return "Unknown"
}

// targetGivens returns how many pre-filled cells the difficulty provides.
func (d Difficulty) targetGivens() int {
	switch d {
	case Easy:
		return 38
	case Medium:
		return 30
	case Hard:
		return 24
	}
	return 30
}

// Cell represents a single cell on the Sudoku board.
type Cell struct {
	Value       int
	Given       bool
	PencilMarks [9]bool
}

// Game holds the complete state of a Sudoku game.
type Game struct {
	Board      [9][9]Cell
	Solution   [9][9]int
	Difficulty Difficulty
	Won        bool
}

// NewGame generates a new Sudoku puzzle at the given difficulty.
func NewGame(diff Difficulty) *Game {
	full := generate()
	board, solution := removeClues(full, diff)
	return &Game{
		Board:      board,
		Solution:   solution,
		Difficulty: diff,
	}
}

// NewGameWithBoard creates a game from predetermined data (for testing).
func NewGameWithBoard(board, solution [9][9]int, givens [9][9]bool) *Game {
	var cells [9][9]Cell
	for r := 0; r < 9; r++ {
		for c := 0; c < 9; c++ {
			cells[r][c] = Cell{
				Value: board[r][c],
				Given: givens[r][c],
			}
		}
	}
	return &Game{
		Board:    cells,
		Solution: solution,
	}
}

// SetCell places a value in a non-given cell.
func (g *Game) SetCell(row, col, value int) error {
	if row < 0 || row > 8 || col < 0 || col > 8 {
		return errors.New("position out of range")
	}
	if value < 1 || value > 9 {
		return errors.New("value must be 1-9")
	}
	if g.Board[row][col].Given {
		return errors.New("cannot modify a given cell")
	}
	g.Board[row][col].Value = value
	g.Board[row][col].PencilMarks = [9]bool{}
	if g.IsSolved() {
		g.Won = true
	}
	return nil
}

// ClearCell removes the value from a non-given cell.
func (g *Game) ClearCell(row, col int) error {
	if row < 0 || row > 8 || col < 0 || col > 8 {
		return errors.New("position out of range")
	}
	if g.Board[row][col].Given {
		return errors.New("cannot clear a given cell")
	}
	g.Board[row][col].Value = 0
	return nil
}

// TogglePencilMark flips a pencil mark on a non-given, empty cell.
func (g *Game) TogglePencilMark(row, col, num int) error {
	if row < 0 || row > 8 || col < 0 || col > 8 {
		return errors.New("position out of range")
	}
	if num < 1 || num > 9 {
		return errors.New("number must be 1-9")
	}
	if g.Board[row][col].Given {
		return errors.New("cannot pencil mark a given cell")
	}
	g.Board[row][col].PencilMarks[num-1] = !g.Board[row][col].PencilMarks[num-1]
	return nil
}

// HasConflict returns true if the cell's value duplicates a peer
// in the same row, column, or 3x3 box.
func (g *Game) HasConflict(row, col int) bool {
	v := g.Board[row][col].Value
	if v == 0 {
		return false
	}
	// Check row
	for c := 0; c < 9; c++ {
		if c != col && g.Board[row][c].Value == v {
			return true
		}
	}
	// Check column
	for r := 0; r < 9; r++ {
		if r != row && g.Board[r][col].Value == v {
			return true
		}
	}
	// Check 3x3 box
	boxR, boxC := (row/3)*3, (col/3)*3
	for r := boxR; r < boxR+3; r++ {
		for c := boxC; c < boxC+3; c++ {
			if (r != row || c != col) && g.Board[r][c].Value == v {
				return true
			}
		}
	}
	return false
}

// IsSolved returns true when all cells are filled with no conflicts.
func (g *Game) IsSolved() bool {
	for r := 0; r < 9; r++ {
		for c := 0; c < 9; c++ {
			if g.Board[r][c].Value == 0 {
				return false
			}
			if g.HasConflict(r, c) {
				return false
			}
		}
	}
	return true
}

// Hint returns the solution value for the given cell.
func (g *Game) Hint(row, col int) (int, error) {
	if row < 0 || row > 8 || col < 0 || col > 8 {
		return 0, errors.New("position out of range")
	}
	if g.Board[row][col].Given {
		return 0, errors.New("cell is already given")
	}
	return g.Solution[row][col], nil
}

// FilledCount returns how many cells have a value.
func (g *Game) FilledCount() int {
	count := 0
	for r := 0; r < 9; r++ {
		for c := 0; c < 9; c++ {
			if g.Board[r][c].Value != 0 {
				count++
			}
		}
	}
	return count
}

// GivenCount returns how many cells are pre-filled givens.
func (g *Game) GivenCount() int {
	count := 0
	for r := 0; r < 9; r++ {
		for c := 0; c < 9; c++ {
			if g.Board[r][c].Given {
				count++
			}
		}
	}
	return count
}

// --- Puzzle generation ---

// generate creates a complete valid 9x9 Sudoku board.
func generate() [9][9]int {
	var board [9][9]int

	// Fill the three diagonal 3x3 boxes independently.
	// These boxes don't constrain each other.
	for b := 0; b < 3; b++ {
		fillBox(&board, b*3, b*3)
	}

	// Solve the rest using backtracking with randomized candidates.
	solveRandom(&board)
	return board
}

// fillBox fills a 3x3 box starting at (startR, startC) with a random
// permutation of 1-9 values that aren't already placed.
func fillBox(board *[9][9]int, startR, startC int) {
	nums := rand.Perm(9)
	idx := 0
	for r := startR; r < startR+3; r++ {
		for c := startC; c < startC+3; c++ {
			board[r][c] = nums[idx] + 1
			idx++
		}
	}
}

// solveRandom fills empty cells using backtracking with shuffled candidates.
// Used only during generation to produce a random complete board.
func solveRandom(board *[9][9]int) bool {
	row, col, found := findEmpty(board)
	if !found {
		return true // board is complete
	}
	candidates := rand.Perm(9)
	for _, ci := range candidates {
		num := ci + 1
		if isValidPlacement(board, row, col, num) {
			board[row][col] = num
			if solveRandom(board) {
				return true
			}
			board[row][col] = 0
		}
	}
	return false
}

// solve counts solutions up to countLimit using ordered backtracking.
// Returns the number of solutions found (capped at countLimit).
func solve(board *[9][9]int, countLimit int) int {
	return solveCount(board, countLimit, 0)
}

func solveCount(board *[9][9]int, countLimit, count int) int {
	row, col, found := findEmpty(board)
	if !found {
		return count + 1 // found a solution
	}
	for num := 1; num <= 9; num++ {
		if isValidPlacement(board, row, col, num) {
			board[row][col] = num
			count = solveCount(board, countLimit, count)
			if count >= countLimit {
				return count // leave board filled when limit reached
			}
			board[row][col] = 0
		}
	}
	return count
}

// findEmpty returns the first empty cell (value 0) scanning left-to-right,
// top-to-bottom. Returns (row, col, true) or (0, 0, false) if none.
func findEmpty(board *[9][9]int) (row, col int, found bool) {
	for r := 0; r < 9; r++ {
		for c := 0; c < 9; c++ {
			if board[r][c] == 0 {
				return r, c, true
			}
		}
	}
	return 0, 0, false
}

// isValidPlacement checks whether placing num at (row, col) violates
// any Sudoku constraints.
func isValidPlacement(board *[9][9]int, row, col, num int) bool {
	// Check row
	for c := 0; c < 9; c++ {
		if board[row][c] == num {
			return false
		}
	}
	// Check column
	for r := 0; r < 9; r++ {
		if board[r][col] == num {
			return false
		}
	}
	// Check 3x3 box
	boxR, boxC := (row/3)*3, (col/3)*3
	for r := boxR; r < boxR+3; r++ {
		for c := boxC; c < boxC+3; c++ {
			if board[r][c] == num {
				return false
			}
		}
	}
	return true
}

// removeClues removes values from a complete board to create a puzzle,
// ensuring a unique solution. Returns the puzzle board and solution.
func removeClues(full [9][9]int, diff Difficulty) (puzzle [9][9]Cell, solution [9][9]int) {
	solution = full
	target := diff.targetGivens()

	// Start with all cells as givens.
	for r := 0; r < 9; r++ {
		for c := 0; c < 9; c++ {
			puzzle[r][c] = Cell{Value: full[r][c], Given: true}
		}
	}

	// Create a random order of all 81 positions.
	positions := rand.Perm(81)

	currentGivens := 81
	for _, pos := range positions {
		if currentGivens <= target {
			break
		}
		r, c := pos/9, pos%9
		saved := puzzle[r][c].Value

		// Temporarily remove the value and check uniqueness.
		var check [9][9]int
		for ri := 0; ri < 9; ri++ {
			for ci := 0; ci < 9; ci++ {
				if puzzle[ri][ci].Given && !(ri == r && ci == c) {
					check[ri][ci] = puzzle[ri][ci].Value
				}
			}
		}

		if solve(&check, 2) == 1 {
			puzzle[r][c].Value = 0
			puzzle[r][c].Given = false
			currentGivens--
		} else {
			puzzle[r][c].Value = saved
		}
	}

	return puzzle, solution
}
