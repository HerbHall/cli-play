package minesweeper

import "math/rand/v2"

// Difficulty represents a minesweeper difficulty preset.
type Difficulty int

const (
	Beginner     Difficulty = iota
	Intermediate
	Expert
)

// DifficultyConfig holds the grid dimensions and mine count for a difficulty.
type DifficultyConfig struct {
	Rows  int
	Cols  int
	Mines int
}

var difficulties = map[Difficulty]DifficultyConfig{
	Beginner:     {Rows: 9, Cols: 9, Mines: 10},
	Intermediate: {Rows: 16, Cols: 16, Mines: 40},
	Expert:       {Rows: 16, Cols: 30, Mines: 99},
}

// GetConfig returns the configuration for a difficulty level.
func GetConfig(d Difficulty) DifficultyConfig {
	return difficulties[d]
}

// CellState represents the visibility state of a cell.
type CellState int

const (
	Hidden   CellState = iota
	Revealed
	Flagged
)

// Cell represents a single cell on the minesweeper grid.
type Cell struct {
	Mine     bool
	State    CellState
	Adjacent int
}

// GameState represents the overall state of the game.
type GameState int

const (
	Playing GameState = iota
	Won
	Lost
)

// Game holds the complete state of a minesweeper game.
type Game struct {
	Grid          [][]Cell
	Rows          int
	Cols          int
	TotalMines    int
	FlagsUsed     int
	CellsRevealed int
	State         GameState
	FirstClick    bool
}

// NewGame creates a new game with mines not yet placed (placed on first click).
func NewGame(diff Difficulty) *Game {
	cfg := difficulties[diff]
	grid := make([][]Cell, cfg.Rows)
	for r := range grid {
		grid[r] = make([]Cell, cfg.Cols)
	}
	return &Game{
		Grid:       grid,
		Rows:       cfg.Rows,
		Cols:       cfg.Cols,
		TotalMines: cfg.Mines,
		FirstClick: true,
	}
}

// NewGameWithMines creates a game with mines at specific positions (for testing).
// Sets FirstClick to false since mines are already placed.
func NewGameWithMines(rows, cols int, mines [][2]int) *Game {
	grid := make([][]Cell, rows)
	for r := range grid {
		grid[r] = make([]Cell, cols)
	}
	g := &Game{
		Grid:       grid,
		Rows:       rows,
		Cols:       cols,
		TotalMines: len(mines),
		FirstClick: false,
	}
	for _, pos := range mines {
		g.Grid[pos[0]][pos[1]].Mine = true
	}
	g.computeAdjacent()
	return g
}

// placeMines randomly places mines on the grid, excluding the safe cell and
// its 8 neighbors. Called on the first Reveal.
func (g *Game) placeMines(safeRow, safeCol int) {
	excluded := make(map[[2]int]bool)
	for _, n := range g.neighbors(safeRow, safeCol) {
		excluded[n] = true
	}
	excluded[[2]int{safeRow, safeCol}] = true

	placed := 0
	for placed < g.TotalMines {
		r := rand.IntN(g.Rows)
		c := rand.IntN(g.Cols)
		pos := [2]int{r, c}
		if excluded[pos] || g.Grid[r][c].Mine {
			continue
		}
		g.Grid[r][c].Mine = true
		placed++
	}
	g.computeAdjacent()
}

// computeAdjacent calculates the adjacent mine count for every cell.
func (g *Game) computeAdjacent() {
	for r := 0; r < g.Rows; r++ {
		for c := 0; c < g.Cols; c++ {
			if g.Grid[r][c].Mine {
				continue
			}
			count := 0
			for _, n := range g.neighbors(r, c) {
				if g.Grid[n[0]][n[1]].Mine {
					count++
				}
			}
			g.Grid[r][c].Adjacent = count
		}
	}
}

// Reveal uncovers a cell. Returns false if the cell cannot be revealed
// (out of bounds, already revealed, or flagged). On first click, mines
// are placed avoiding the clicked cell. Hitting a mine ends the game.
// Revealing a zero-adjacent cell flood-fills neighboring cells.
func (g *Game) Reveal(row, col int) bool {
	if !g.inBounds(row, col) {
		return false
	}
	cell := &g.Grid[row][col]
	if cell.State == Revealed || cell.State == Flagged {
		return false
	}
	if g.State != Playing {
		return false
	}

	if g.FirstClick {
		g.placeMines(row, col)
		g.FirstClick = false
	}

	if cell.Mine {
		g.State = Lost
		g.revealAllMines()
		return true
	}

	g.floodReveal(row, col)
	g.checkWin()
	return true
}

// floodReveal uses BFS to reveal a cell and, if it has zero adjacent mines,
// continues revealing neighbors until hitting numbered cells.
func (g *Game) floodReveal(row, col int) {
	type pos struct{ r, c int }
	queue := []pos{{row, col}}

	for len(queue) > 0 {
		p := queue[0]
		queue = queue[1:]

		cell := &g.Grid[p.r][p.c]
		if cell.State == Revealed {
			continue
		}
		if cell.State == Flagged {
			continue
		}
		if cell.Mine {
			continue
		}

		cell.State = Revealed
		g.CellsRevealed++

		if cell.Adjacent == 0 {
			for _, n := range g.neighbors(p.r, p.c) {
				if g.Grid[n[0]][n[1]].State == Hidden {
					queue = append(queue, pos{n[0], n[1]})
				}
			}
		}
	}
}

// revealAllMines shows all mine locations (called on game loss).
func (g *Game) revealAllMines() {
	for r := 0; r < g.Rows; r++ {
		for c := 0; c < g.Cols; c++ {
			if g.Grid[r][c].Mine {
				g.Grid[r][c].State = Revealed
			}
		}
	}
}

// ToggleFlag toggles the flag state on a hidden cell.
func (g *Game) ToggleFlag(row, col int) {
	if !g.inBounds(row, col) || g.State != Playing {
		return
	}
	cell := &g.Grid[row][col]
	switch cell.State {
	case Hidden:
		cell.State = Flagged
		g.FlagsUsed++
	case Flagged:
		cell.State = Hidden
		g.FlagsUsed--
	}
}

// checkWin sets the game state to Won if all non-mine cells are revealed.
func (g *Game) checkWin() {
	if g.CellsRevealed == g.Rows*g.Cols-g.TotalMines {
		g.State = Won
	}
}

// inBounds returns true if the coordinates are within the grid.
func (g *Game) inBounds(row, col int) bool {
	return row >= 0 && row < g.Rows && col >= 0 && col < g.Cols
}

// neighbors returns the valid neighboring coordinates for a cell.
func (g *Game) neighbors(row, col int) [][2]int {
	var result [][2]int
	for dr := -1; dr <= 1; dr++ {
		for dc := -1; dc <= 1; dc++ {
			if dr == 0 && dc == 0 {
				continue
			}
			nr, nc := row+dr, col+dc
			if g.inBounds(nr, nc) {
				result = append(result, [2]int{nr, nc})
			}
		}
	}
	return result
}
