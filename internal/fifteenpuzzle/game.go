package fifteenpuzzle

import "math/rand/v2"

const boardSize = 4

// Board is a 4x4 grid. 0 represents the empty space.
type Board [boardSize][boardSize]int

// Game holds the complete state of a Fifteen Puzzle game.
type Game struct {
	Board    Board
	EmptyRow int
	EmptyCol int
	Moves    int
	Won      bool
}

// NewGame creates a fresh solvable puzzle.
func NewGame() *Game {
	g := &Game{}
	g.shuffle()
	return g
}

// NewGameFromBoard creates a game with a specific board state (for testing).
func NewGameFromBoard(b Board) *Game {
	g := &Game{Board: b}
	for r := range boardSize {
		for c := range boardSize {
			if b[r][c] == 0 {
				g.EmptyRow = r
				g.EmptyCol = c
			}
		}
	}
	return g
}

// shuffle generates a random solvable board.
func (g *Game) shuffle() {
	// Create solved state: 1..15, 0
	tiles := make([]int, 0, boardSize*boardSize)
	for i := 1; i < boardSize*boardSize; i++ {
		tiles = append(tiles, i)
	}
	tiles = append(tiles, 0)

	// Shuffle until solvable
	for {
		rand.Shuffle(len(tiles), func(i, j int) {
			tiles[i], tiles[j] = tiles[j], tiles[i]
		})

		// Place tiles on board
		idx := 0
		for r := range boardSize {
			for c := range boardSize {
				g.Board[r][c] = tiles[idx]
				if tiles[idx] == 0 {
					g.EmptyRow = r
					g.EmptyCol = c
				}
				idx++
			}
		}

		if IsSolvable(g.Board) && !g.isSolved() {
			break
		}
	}

	g.Moves = 0
	g.Won = false
}

// MoveUp slides the tile below the empty space up into it.
func (g *Game) MoveUp() bool {
	return g.move(g.EmptyRow+1, g.EmptyCol)
}

// MoveDown slides the tile above the empty space down into it.
func (g *Game) MoveDown() bool {
	return g.move(g.EmptyRow-1, g.EmptyCol)
}

// MoveLeft slides the tile to the right of the empty space left into it.
func (g *Game) MoveLeft() bool {
	return g.move(g.EmptyRow, g.EmptyCol+1)
}

// MoveRight slides the tile to the left of the empty space right into it.
func (g *Game) MoveRight() bool {
	return g.move(g.EmptyRow, g.EmptyCol-1)
}

// move swaps the tile at (r, c) with the empty space if adjacent.
func (g *Game) move(r, c int) bool {
	if g.Won {
		return false
	}
	if r < 0 || r >= boardSize || c < 0 || c >= boardSize {
		return false
	}

	g.Board[g.EmptyRow][g.EmptyCol] = g.Board[r][c]
	g.Board[r][c] = 0
	g.EmptyRow = r
	g.EmptyCol = c
	g.Moves++

	if g.isSolved() {
		g.Won = true
	}

	return true
}

// isSolved checks if tiles are in order: 1..15, 0.
func (g *Game) isSolved() bool {
	expected := 1
	for r := range boardSize {
		for c := range boardSize {
			if r == boardSize-1 && c == boardSize-1 {
				return g.Board[r][c] == 0
			}
			if g.Board[r][c] != expected {
				return false
			}
			expected++
		}
	}
	return true
}

// IsSolvable returns true if the board configuration can be solved.
// For an even-width grid (4x4), a position is solvable when
// (inversion count + row of empty from bottom, 1-indexed) is even.
// Equivalently: inversions is even when empty is on an odd row from
// bottom (1, 3), or inversions is odd when empty is on an even row (2, 4).
func IsSolvable(b Board) bool {
	flat := make([]int, 0, boardSize*boardSize)
	emptyRowFromBottom := 0

	for r := range boardSize {
		for c := range boardSize {
			if b[r][c] == 0 {
				emptyRowFromBottom = boardSize - r
			} else {
				flat = append(flat, b[r][c])
			}
		}
	}

	inversions := countInversions(flat)

	if emptyRowFromBottom%2 == 1 {
		// Empty on odd row from bottom (1st or 3rd): need even inversions.
		return inversions%2 == 0
	}
	// Empty on even row from bottom (2nd or 4th): need odd inversions.
	return inversions%2 == 1
}

// countInversions counts the number of pairs (i, j) where i < j
// but flat[i] > flat[j].
func countInversions(flat []int) int {
	count := 0
	for i := 0; i < len(flat); i++ {
		for j := i + 1; j < len(flat); j++ {
			if flat[i] > flat[j] {
				count++
			}
		}
	}
	return count
}

// Reset creates a new shuffled board.
func (g *Game) Reset() {
	g.shuffle()
}
