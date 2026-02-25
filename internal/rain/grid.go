package rain

// CellState tracks what a grid cell is currently doing.
type CellState int

const (
	CellEmpty      CellState = iota
	CellSplash               // Splash text character
	CellRain                 // Active rain character
	CellDissolving           // Splash char being eaten by rain
	CellDeposited            // Menu char locked in place
)

// Cell is a single position in the terminal grid.
type Cell struct {
	Char       rune
	State      CellState
	FadeStep   int  // animation progress counter
	TargetChar rune // for deposited: the menu char to reveal
}

// CellGrid is a full-screen grid of cells indexed [y][x].
type CellGrid struct {
	Width  int
	Height int
	Cells  [][]Cell
}

// NewCellGrid allocates a grid of the given dimensions.
// Every cell starts as CellEmpty.
func NewCellGrid(width, height int) *CellGrid {
	g := &CellGrid{
		Width:  width,
		Height: height,
		Cells:  allocCells(width, height),
	}
	return g
}

// Get returns the cell at (x, y).
// Out-of-bounds coordinates return a zero-value CellEmpty cell.
func (g *CellGrid) Get(x, y int) Cell {
	if x < 0 || x >= g.Width || y < 0 || y >= g.Height {
		return Cell{}
	}
	return g.Cells[y][x]
}

// Set writes a cell at (x, y).
// Out-of-bounds coordinates are silently ignored.
func (g *CellGrid) Set(x, y int, c Cell) {
	if x < 0 || x >= g.Width || y < 0 || y >= g.Height {
		return
	}
	g.Cells[y][x] = c
}

// Resize reallocates the grid to new dimensions, preserving cells
// that fit within both old and new bounds.
func (g *CellGrid) Resize(width, height int) {
	newCells := allocCells(width, height)

	// Copy the overlapping region.
	copyW := g.Width
	if width < copyW {
		copyW = width
	}
	copyH := g.Height
	if height < copyH {
		copyH = height
	}
	for y := 0; y < copyH; y++ {
		copy(newCells[y][:copyW], g.Cells[y][:copyW])
	}

	g.Width = width
	g.Height = height
	g.Cells = newCells
}

// Clear resets every cell to CellEmpty.
func (g *CellGrid) Clear() {
	for y := range g.Cells {
		for x := range g.Cells[y] {
			g.Cells[y][x] = Cell{}
		}
	}
}

// allocCells creates a height x width slice of empty cells.
func allocCells(width, height int) [][]Cell {
	cells := make([][]Cell, height)
	for y := range cells {
		cells[y] = make([]Cell, width)
	}
	return cells
}
