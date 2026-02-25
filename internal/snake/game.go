package snake

import "math/rand/v2"

// Direction represents the snake's movement direction.
type Direction int

const (
	DirUp Direction = iota
	DirDown
	DirLeft
	DirRight
)

// Point represents a coordinate on the board.
type Point struct {
	X int
	Y int
}

// Board dimensions.
const (
	BoardWidth  = 20
	BoardHeight = 15
)

// GameState represents the overall state of the game.
type GameState int

const (
	StatePlaying GameState = iota
	StateGameOver
)

// Game holds the complete state of a Snake game.
type Game struct {
	Snake     []Point
	Dir       Direction
	Food      Point
	Score     int
	State     GameState
	Width     int
	Height    int
	randFunc  func(int) int
}

// NewGame creates a fresh Snake game with the snake in the center.
func NewGame() *Game {
	g := &Game{
		Width:    BoardWidth,
		Height:   BoardHeight,
		Dir:      DirRight,
		State:    StatePlaying,
		randFunc: rand.IntN,
	}

	// Snake starts in the center, length 3, moving right.
	centerX := BoardWidth / 2
	centerY := BoardHeight / 2
	g.Snake = []Point{
		{X: centerX, Y: centerY},
		{X: centerX - 1, Y: centerY},
		{X: centerX - 2, Y: centerY},
	}

	g.spawnFood()
	return g
}

// SetDirection changes the snake's direction, preventing 180-degree reversals.
func (g *Game) SetDirection(d Direction) {
	if g.State != StatePlaying {
		return
	}
	switch d {
	case DirUp:
		if g.Dir == DirDown {
			return
		}
	case DirDown:
		if g.Dir == DirUp {
			return
		}
	case DirLeft:
		if g.Dir == DirRight {
			return
		}
	case DirRight:
		if g.Dir == DirLeft {
			return
		}
	}
	g.Dir = d
}

// Tick advances the game by one step: moves the snake, checks collisions,
// and handles food eating.
func (g *Game) Tick() {
	if g.State != StatePlaying {
		return
	}

	head := g.Snake[0]
	var next Point

	switch g.Dir {
	case DirUp:
		next = Point{X: head.X, Y: head.Y - 1}
	case DirDown:
		next = Point{X: head.X, Y: head.Y + 1}
	case DirLeft:
		next = Point{X: head.X - 1, Y: head.Y}
	case DirRight:
		next = Point{X: head.X + 1, Y: head.Y}
	}

	// Check wall collision.
	if next.X < 0 || next.X >= g.Width || next.Y < 0 || next.Y >= g.Height {
		g.State = StateGameOver
		return
	}

	// Check self-collision (exclude tail since it will move away, unless eating).
	if g.isSelfCollision(next) {
		g.State = StateGameOver
		return
	}

	// Move snake forward.
	g.Snake = append([]Point{next}, g.Snake...)

	// Check food.
	if next == g.Food {
		g.Score++
		g.spawnFood()
	} else {
		// Remove tail if not eating.
		g.Snake = g.Snake[:len(g.Snake)-1]
	}
}

// isSelfCollision checks if the point collides with any part of the snake body.
// Excludes the last segment since the tail moves away on a normal tick.
func (g *Game) isSelfCollision(p Point) bool {
	// Check all body segments except the tail (which will be removed).
	for i := 0; i < len(g.Snake)-1; i++ {
		if g.Snake[i] == p {
			return true
		}
	}
	return false
}

// spawnFood places food on a random empty cell.
func (g *Game) spawnFood() {
	occupied := make(map[Point]bool, len(g.Snake))
	for i := range g.Snake {
		occupied[g.Snake[i]] = true
	}

	// Collect empty cells.
	empty := make([]Point, 0, g.Width*g.Height-len(g.Snake))
	for y := 0; y < g.Height; y++ {
		for x := 0; x < g.Width; x++ {
			p := Point{X: x, Y: y}
			if !occupied[p] {
				empty = append(empty, p)
			}
		}
	}

	if len(empty) == 0 {
		return
	}

	g.Food = empty[g.randFunc(len(empty))]
}

// Reset restarts the game.
func (g *Game) Reset() {
	rf := g.randFunc
	*g = *NewGame()
	g.randFunc = rf
	g.spawnFood()
}

// IsOccupied returns true if the given point is part of the snake.
func (g *Game) IsOccupied(p Point) bool {
	for i := range g.Snake {
		if g.Snake[i] == p {
			return true
		}
	}
	return false
}
