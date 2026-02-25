package twofortyeight

import "math/rand/v2"

// Direction represents a slide direction.
type Direction int

const (
	Up Direction = iota
	Down
	Left
	Right
)

// boardSize is the width and height of the game board.
const boardSize = 4

// Game holds the complete state of a 2048 game.
type Game struct {
	Board     [boardSize][boardSize]int
	Score     int
	Won       bool
	Over      bool
	Continued bool
}

// NewGame creates a fresh game with two random tiles.
func NewGame() *Game {
	g := &Game{}
	g.SpawnTile()
	g.SpawnTile()
	return g
}

// Reset clears the board and spawns two new tiles.
func (g *Game) Reset() {
	g.Board = [boardSize][boardSize]int{}
	g.Score = 0
	g.Won = false
	g.Over = false
	g.Continued = false
	g.SpawnTile()
	g.SpawnTile()
}

// ContinueAfterWin allows play to continue past the 2048 tile.
func (g *Game) ContinueAfterWin() {
	g.Continued = true
}

// Move slides all tiles in the given direction, merges equal adjacent
// pairs (each pair merges at most once per move), adds merge values
// to the score, and spawns a new tile if the board changed. Returns
// true if the board changed.
func (g *Game) Move(dir Direction) bool {
	prev := g.Board

	switch dir {
	case Left:
		for r := range boardSize {
			line := g.Board[r]
			merged := slideLine(line)
			g.Score += merged.score
			g.Board[r] = merged.line
		}
	case Right:
		for r := range boardSize {
			line := reverseLine(g.Board[r])
			merged := slideLine(line)
			g.Score += merged.score
			g.Board[r] = reverseLine(merged.line)
		}
	case Up:
		for c := range boardSize {
			line := extractColumn(g.Board, c)
			merged := slideLine(line)
			g.Score += merged.score
			setColumn(&g.Board, c, merged.line)
		}
	case Down:
		for c := range boardSize {
			line := reverseLine(extractColumn(g.Board, c))
			merged := slideLine(line)
			g.Score += merged.score
			setColumn(&g.Board, c, reverseLine(merged.line))
		}
	}

	changed := prev != g.Board
	if changed {
		g.SpawnTile()
		g.checkState()
	}
	return changed
}

// SpawnTile places a 2 (90% chance) or 4 (10% chance) on a random empty cell.
func (g *Game) SpawnTile() {
	empty := g.emptyCells()
	if len(empty) == 0 {
		return
	}
	cell := empty[rand.IntN(len(empty))]
	val := 2
	if rand.IntN(10) == 0 {
		val = 4
	}
	g.Board[cell.r][cell.c] = val
}

// CanMove returns true if any move is possible: an empty cell exists
// or any two adjacent cells have equal values.
func (g *Game) CanMove() bool {
	for r := range boardSize {
		for c := range boardSize {
			if g.Board[r][c] == 0 {
				return true
			}
			v := g.Board[r][c]
			if c+1 < boardSize && g.Board[r][c+1] == v {
				return true
			}
			if r+1 < boardSize && g.Board[r+1][c] == v {
				return true
			}
		}
	}
	return false
}

type cell struct{ r, c int }

func (g *Game) emptyCells() []cell {
	var cells []cell
	for r := range boardSize {
		for c := range boardSize {
			if g.Board[r][c] == 0 {
				cells = append(cells, cell{r, c})
			}
		}
	}
	return cells
}

func (g *Game) checkState() {
	if !g.Won && !g.Continued {
		for r := range boardSize {
			for c := range boardSize {
				if g.Board[r][c] == 2048 {
					g.Won = true
					return
				}
			}
		}
	}
	if !g.CanMove() {
		g.Over = true
	}
}

// slideResult holds the merged line and the score earned from merges.
type slideResult struct {
	line  [boardSize]int
	score int
}

// slideLine compacts non-zero values to the left, then merges adjacent
// equal pairs left-to-right (each cell merges at most once).
func slideLine(line [boardSize]int) slideResult {
	// Step 1: compact non-zero values to the front.
	var compact [boardSize]int
	idx := 0
	for _, v := range line {
		if v != 0 {
			compact[idx] = v
			idx++
		}
	}

	// Step 2: merge adjacent equal pairs left-to-right.
	var result [boardSize]int
	score := 0
	ri := 0
	for i := 0; i < boardSize; i++ {
		if compact[i] == 0 {
			break
		}
		if i+1 < boardSize && compact[i] == compact[i+1] {
			merged := compact[i] * 2
			result[ri] = merged
			score += merged
			i++ // skip the merged partner
		} else {
			result[ri] = compact[i]
		}
		ri++
	}

	return slideResult{line: result, score: score}
}

func reverseLine(line [boardSize]int) [boardSize]int {
	return [boardSize]int{line[3], line[2], line[1], line[0]}
}

func extractColumn(board [boardSize][boardSize]int, c int) [boardSize]int {
	return [boardSize]int{board[0][c], board[1][c], board[2][c], board[3][c]}
}

func setColumn(board *[boardSize][boardSize]int, c int, col [boardSize]int) {
	for r := range boardSize {
		board[r][c] = col[r]
	}
}
