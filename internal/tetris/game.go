package tetris

import "math/rand/v2"

// Board dimensions.
const (
	BoardWidth  = 10
	BoardHeight = 20
)

// PieceType identifies a tetromino shape.
type PieceType int

const (
	PieceI PieceType = iota
	PieceO
	PieceT
	PieceS
	PieceZ
	PieceJ
	PieceL
	PieceNone // empty cell
)

// pieceCount is the number of distinct tetromino types.
const pieceCount = 7

// Point represents a row,col coordinate on the board.
type Point struct {
	Row, Col int
}

// rotation holds the cell offsets for one rotation state.
type rotation [4]Point

// pieceRotations defines all rotation states for each piece type.
// Each rotation is defined as 4 cell offsets relative to the piece origin.
var pieceRotations = [pieceCount][]rotation{
	// I
	{
		{Point{0, 0}, Point{0, 1}, Point{0, 2}, Point{0, 3}},
		{Point{0, 0}, Point{1, 0}, Point{2, 0}, Point{3, 0}},
		{Point{0, 0}, Point{0, 1}, Point{0, 2}, Point{0, 3}},
		{Point{0, 0}, Point{1, 0}, Point{2, 0}, Point{3, 0}},
	},
	// O
	{
		{Point{0, 0}, Point{0, 1}, Point{1, 0}, Point{1, 1}},
	},
	// T
	{
		{Point{0, 0}, Point{0, 1}, Point{0, 2}, Point{1, 1}},
		{Point{0, 0}, Point{1, 0}, Point{2, 0}, Point{1, 1}},
		{Point{1, 0}, Point{1, 1}, Point{1, 2}, Point{0, 1}},
		{Point{0, 1}, Point{1, 1}, Point{2, 1}, Point{1, 0}},
	},
	// S
	{
		{Point{0, 1}, Point{0, 2}, Point{1, 0}, Point{1, 1}},
		{Point{0, 0}, Point{1, 0}, Point{1, 1}, Point{2, 1}},
		{Point{0, 1}, Point{0, 2}, Point{1, 0}, Point{1, 1}},
		{Point{0, 0}, Point{1, 0}, Point{1, 1}, Point{2, 1}},
	},
	// Z
	{
		{Point{0, 0}, Point{0, 1}, Point{1, 1}, Point{1, 2}},
		{Point{0, 1}, Point{1, 0}, Point{1, 1}, Point{2, 0}},
		{Point{0, 0}, Point{0, 1}, Point{1, 1}, Point{1, 2}},
		{Point{0, 1}, Point{1, 0}, Point{1, 1}, Point{2, 0}},
	},
	// J
	{
		{Point{0, 0}, Point{1, 0}, Point{1, 1}, Point{1, 2}},
		{Point{0, 0}, Point{0, 1}, Point{1, 0}, Point{2, 0}},
		{Point{0, 0}, Point{0, 1}, Point{0, 2}, Point{1, 2}},
		{Point{0, 0}, Point{1, 0}, Point{2, 0}, Point{2, -1}},
	},
	// L
	{
		{Point{0, 2}, Point{1, 0}, Point{1, 1}, Point{1, 2}},
		{Point{0, 0}, Point{1, 0}, Point{2, 0}, Point{2, 1}},
		{Point{0, 0}, Point{0, 1}, Point{0, 2}, Point{1, 0}},
		{Point{0, 0}, Point{0, 1}, Point{1, 1}, Point{2, 1}},
	},
}

// Piece represents the currently falling tetromino.
type Piece struct {
	Type     PieceType
	Rotation int
	Row      int // top-left origin row on the board
	Col      int // top-left origin col on the board
}

// Cells returns the absolute board coordinates of the piece's four cells.
func (p Piece) Cells() [4]Point {
	rots := pieceRotations[p.Type]
	rot := rots[p.Rotation%len(rots)]
	var cells [4]Point
	for i, off := range rot {
		cells[i] = Point{p.Row + off.Row, p.Col + off.Col}
	}
	return cells
}

// Game holds the complete state of a Tetris game.
type Game struct {
	Board   [BoardHeight][BoardWidth]PieceType
	Current Piece
	Next    PieceType
	Score   int
	Lines   int
	Level   int
	Over    bool
}

// NewGame creates a fresh Tetris game with the first piece spawned.
func NewGame() *Game {
	g := &Game{
		Level: 1,
	}
	g.initBoard()
	g.Next = g.randomPiece()
	g.spawnPiece()
	return g
}

// Reset restarts the game from scratch.
func (g *Game) Reset() {
	g.initBoard()
	g.Score = 0
	g.Lines = 0
	g.Level = 1
	g.Over = false
	g.Next = g.randomPiece()
	g.spawnPiece()
}

func (g *Game) initBoard() {
	for r := range BoardHeight {
		for c := range BoardWidth {
			g.Board[r][c] = PieceNone
		}
	}
}

func (g *Game) randomPiece() PieceType {
	return PieceType(rand.IntN(pieceCount))
}

// spawnPiece places the next piece at the top-center of the board.
// Returns false (and sets Over=true) if the spawn position is blocked.
func (g *Game) spawnPiece() bool {
	g.Current = Piece{
		Type:     g.Next,
		Rotation: 0,
		Row:      0,
		Col:      (BoardWidth - 3) / 2,
	}
	g.Next = g.randomPiece()

	if !g.isValid(g.Current) {
		g.Over = true
		return false
	}
	return true
}

// isValid checks whether the given piece fits on the board without
// overlapping walls, floor, or placed cells.
func (g *Game) isValid(p Piece) bool {
	cells := p.Cells()
	for _, c := range cells {
		if c.Col < 0 || c.Col >= BoardWidth {
			return false
		}
		if c.Row < 0 || c.Row >= BoardHeight {
			return false
		}
		if g.Board[c.Row][c.Col] != PieceNone {
			return false
		}
	}
	return true
}

// MoveLeft shifts the current piece one column left if possible.
func (g *Game) MoveLeft() bool {
	if g.Over {
		return false
	}
	candidate := g.Current
	candidate.Col--
	if g.isValid(candidate) {
		g.Current = candidate
		return true
	}
	return false
}

// MoveRight shifts the current piece one column right if possible.
func (g *Game) MoveRight() bool {
	if g.Over {
		return false
	}
	candidate := g.Current
	candidate.Col++
	if g.isValid(candidate) {
		g.Current = candidate
		return true
	}
	return false
}

// MoveDown shifts the current piece one row down. If it cannot move,
// the piece is locked in place and a new piece spawns.
// Returns true if the piece actually moved down.
func (g *Game) MoveDown() bool {
	if g.Over {
		return false
	}
	candidate := g.Current
	candidate.Row++
	if g.isValid(candidate) {
		g.Current = candidate
		return true
	}
	g.lockPiece()
	return false
}

// Rotate rotates the current piece clockwise if possible.
func (g *Game) Rotate() bool {
	if g.Over {
		return false
	}
	candidate := g.Current
	rots := pieceRotations[candidate.Type]
	candidate.Rotation = (candidate.Rotation + 1) % len(rots)
	if g.isValid(candidate) {
		g.Current = candidate
		return true
	}
	// Basic wall kick: try shifting left and right by 1.
	for _, offset := range []int{-1, 1, -2, 2} {
		kicked := candidate
		kicked.Col += offset
		if g.isValid(kicked) {
			g.Current = kicked
			return true
		}
	}
	return false
}

// HardDrop instantly drops the piece to its lowest valid position
// and locks it. Returns the number of rows dropped.
func (g *Game) HardDrop() int {
	if g.Over {
		return 0
	}
	rows := 0
	for {
		candidate := g.Current
		candidate.Row++
		if !g.isValid(candidate) {
			break
		}
		g.Current = candidate
		rows++
	}
	g.lockPiece()
	return rows
}

// GhostRow returns the row position where the current piece would land
// if hard-dropped. Used for rendering the ghost piece.
func (g *Game) GhostRow() int {
	ghost := g.Current
	for {
		candidate := ghost
		candidate.Row++
		if !g.isValid(candidate) {
			break
		}
		ghost = candidate
	}
	return ghost.Row
}

// lockPiece writes the current piece into the board, clears full lines,
// updates score, and spawns the next piece.
func (g *Game) lockPiece() {
	cells := g.Current.Cells()
	for _, c := range cells {
		if c.Row >= 0 && c.Row < BoardHeight && c.Col >= 0 && c.Col < BoardWidth {
			g.Board[c.Row][c.Col] = g.Current.Type
		}
	}

	cleared := g.clearLines()
	g.Lines += cleared
	g.Score += lineScore(cleared)
	g.Level = g.Lines/10 + 1

	g.spawnPiece()
}

// clearLines removes all completed rows and shifts rows above downward.
// Returns the number of lines cleared.
func (g *Game) clearLines() int {
	cleared := 0
	dst := BoardHeight - 1
	for src := BoardHeight - 1; src >= 0; src-- {
		if g.isRowFull(src) {
			cleared++
			continue
		}
		if dst != src {
			g.Board[dst] = g.Board[src]
		}
		dst--
	}
	// Fill remaining top rows with empty cells.
	for dst >= 0 {
		for c := range BoardWidth {
			g.Board[dst][c] = PieceNone
		}
		dst--
	}
	return cleared
}

func (g *Game) isRowFull(row int) bool {
	for c := range BoardWidth {
		if g.Board[row][c] == PieceNone {
			return false
		}
	}
	return true
}

// lineScore returns the score awarded for clearing n lines at once.
func lineScore(n int) int {
	switch n {
	case 1:
		return 100
	case 2:
		return 300
	case 3:
		return 500
	case 4:
		return 800
	default:
		return 0
	}
}

// TickInterval returns the gravity interval in milliseconds for the
// current level. Starts at 500ms and decreases by 40ms per level,
// with a minimum of 60ms.
func (g *Game) TickInterval() int {
	ms := 500 - (g.Level-1)*40
	if ms < 60 {
		ms = 60
	}
	return ms
}
