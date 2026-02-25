package snake

import "testing"

func TestInitialState(t *testing.T) {
	g := NewGame()

	if g.State != StatePlaying {
		t.Errorf("State = %d, want StatePlaying (%d)", g.State, StatePlaying)
	}
	if g.Dir != DirRight {
		t.Errorf("Dir = %d, want DirRight (%d)", g.Dir, DirRight)
	}
	if len(g.Snake) != 3 {
		t.Errorf("Snake length = %d, want 3", len(g.Snake))
	}
	if g.Score != 0 {
		t.Errorf("Score = %d, want 0", g.Score)
	}

	// Snake should be in the center, horizontal, head first.
	centerX := BoardWidth / 2
	centerY := BoardHeight / 2
	expectedHead := Point{X: centerX, Y: centerY}
	if g.Snake[0] != expectedHead {
		t.Errorf("Head = %v, want %v", g.Snake[0], expectedHead)
	}
	if g.Snake[1] != (Point{X: centerX - 1, Y: centerY}) {
		t.Errorf("Body[1] = %v, want (%d,%d)", g.Snake[1], centerX-1, centerY)
	}
	if g.Snake[2] != (Point{X: centerX - 2, Y: centerY}) {
		t.Errorf("Tail = %v, want (%d,%d)", g.Snake[2], centerX-2, centerY)
	}
}

func TestMovementRight(t *testing.T) {
	g := NewGame()
	head := g.Snake[0]

	g.Tick()

	if g.State != StatePlaying {
		t.Fatal("Expected game to still be playing")
	}
	newHead := g.Snake[0]
	if newHead.X != head.X+1 || newHead.Y != head.Y {
		t.Errorf("After moving right: head = %v, want (%d,%d)", newHead, head.X+1, head.Y)
	}
}

func TestMovementUp(t *testing.T) {
	g := NewGame()
	g.SetDirection(DirUp)
	head := g.Snake[0]

	g.Tick()

	newHead := g.Snake[0]
	if newHead.X != head.X || newHead.Y != head.Y-1 {
		t.Errorf("After moving up: head = %v, want (%d,%d)", newHead, head.X, head.Y-1)
	}
}

func TestMovementDown(t *testing.T) {
	g := NewGame()
	g.SetDirection(DirDown)
	head := g.Snake[0]

	g.Tick()

	newHead := g.Snake[0]
	if newHead.X != head.X || newHead.Y != head.Y+1 {
		t.Errorf("After moving down: head = %v, want (%d,%d)", newHead, head.X, head.Y+1)
	}
}

func TestMovementLeft(t *testing.T) {
	// First move up (can't go left when moving right initially).
	g := NewGame()
	g.SetDirection(DirUp)
	g.Tick()
	g.SetDirection(DirLeft)
	head := g.Snake[0]

	g.Tick()

	newHead := g.Snake[0]
	if newHead.X != head.X-1 || newHead.Y != head.Y {
		t.Errorf("After moving left: head = %v, want (%d,%d)", newHead, head.X-1, head.Y)
	}
}

func TestEatingFood(t *testing.T) {
	g := NewGame()
	// Place food directly ahead of the snake.
	head := g.Snake[0]
	g.Food = Point{X: head.X + 1, Y: head.Y}
	lengthBefore := len(g.Snake)
	scoreBefore := g.Score

	g.Tick()

	if g.Score != scoreBefore+1 {
		t.Errorf("Score = %d, want %d", g.Score, scoreBefore+1)
	}
	if len(g.Snake) != lengthBefore+1 {
		t.Errorf("Snake length = %d, want %d", len(g.Snake), lengthBefore+1)
	}
}

func TestWallCollisionRight(t *testing.T) {
	g := NewGame()
	// Move snake head to the right edge.
	g.Snake[0] = Point{X: BoardWidth - 1, Y: BoardHeight / 2}
	g.Dir = DirRight

	g.Tick()

	if g.State != StateGameOver {
		t.Error("Expected game over after hitting right wall")
	}
}

func TestWallCollisionLeft(t *testing.T) {
	g := NewGame()
	g.Snake[0] = Point{X: 0, Y: BoardHeight / 2}
	g.Snake[1] = Point{X: 1, Y: BoardHeight / 2}
	g.Snake[2] = Point{X: 2, Y: BoardHeight / 2}
	g.Dir = DirLeft

	g.Tick()

	if g.State != StateGameOver {
		t.Error("Expected game over after hitting left wall")
	}
}

func TestWallCollisionTop(t *testing.T) {
	g := NewGame()
	g.Snake[0] = Point{X: BoardWidth / 2, Y: 0}
	g.Dir = DirUp

	g.Tick()

	if g.State != StateGameOver {
		t.Error("Expected game over after hitting top wall")
	}
}

func TestWallCollisionBottom(t *testing.T) {
	g := NewGame()
	g.Snake[0] = Point{X: BoardWidth / 2, Y: BoardHeight - 1}
	g.Dir = DirDown

	g.Tick()

	if g.State != StateGameOver {
		t.Error("Expected game over after hitting bottom wall")
	}
}

func TestSelfCollision(t *testing.T) {
	g := NewGame()
	// Create a snake that will collide with itself:
	// Shape: going right, body wraps back on itself.
	g.Snake = []Point{
		{X: 5, Y: 5}, // head
		{X: 5, Y: 4}, // body
		{X: 6, Y: 4}, // body
		{X: 6, Y: 5}, // body
		{X: 6, Y: 6}, // body
	}
	g.Dir = DirRight // head at (5,5) moves right to (6,5) which is body

	g.Tick()

	if g.State != StateGameOver {
		t.Error("Expected game over after self-collision")
	}
}

func TestCannotReverseDirection(t *testing.T) {
	tests := []struct {
		name    string
		initial Direction
		attempt Direction
	}{
		{"right cannot go left", DirRight, DirLeft},
		{"left cannot go right", DirLeft, DirRight},
		{"up cannot go down", DirUp, DirDown},
		{"down cannot go up", DirDown, DirUp},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewGame()
			g.Dir = tt.initial
			g.SetDirection(tt.attempt)
			if g.Dir != tt.initial {
				t.Errorf("Direction changed to %d after reverse attempt, want %d",
					g.Dir, tt.initial)
			}
		})
	}
}

func TestFoodSpawnsOnEmptyCell(t *testing.T) {
	g := NewGame()

	// Run several food spawns and verify food is never on the snake.
	for range 50 {
		g.spawnFood()
		for i := range g.Snake {
			if g.Snake[i] == g.Food {
				t.Fatalf("Food at %v is on snake segment %d", g.Food, i)
			}
		}
	}
}

func TestFoodWithinBounds(t *testing.T) {
	g := NewGame()

	for range 50 {
		g.spawnFood()
		if g.Food.X < 0 || g.Food.X >= g.Width {
			t.Fatalf("Food X=%d out of bounds [0,%d)", g.Food.X, g.Width)
		}
		if g.Food.Y < 0 || g.Food.Y >= g.Height {
			t.Fatalf("Food Y=%d out of bounds [0,%d)", g.Food.Y, g.Height)
		}
	}
}

func TestGameOverState(t *testing.T) {
	g := NewGame()
	g.Snake[0] = Point{X: BoardWidth - 1, Y: BoardHeight / 2}
	g.Dir = DirRight

	g.Tick()

	if g.State != StateGameOver {
		t.Fatal("Expected StateGameOver")
	}

	// Tick should be a no-op when game is over.
	snakeBefore := make([]Point, len(g.Snake))
	copy(snakeBefore, g.Snake)
	g.Tick()
	for i := range g.Snake {
		if g.Snake[i] != snakeBefore[i] {
			t.Error("Tick should not move snake after game over")
			break
		}
	}

	// SetDirection should be a no-op when game is over.
	g.SetDirection(DirUp)
	if g.Dir != DirRight {
		t.Error("SetDirection should not change direction after game over")
	}
}

func TestReset(t *testing.T) {
	g := NewGame()
	// Advance the game.
	g.Tick()
	g.Tick()
	g.Score = 5

	g.Reset()

	if g.Score != 0 {
		t.Errorf("Score after reset = %d, want 0", g.Score)
	}
	if g.State != StatePlaying {
		t.Errorf("State after reset = %d, want StatePlaying", g.State)
	}
	if len(g.Snake) != 3 {
		t.Errorf("Snake length after reset = %d, want 3", len(g.Snake))
	}
	if g.Dir != DirRight {
		t.Errorf("Direction after reset = %d, want DirRight", g.Dir)
	}
}

func TestSnakeLengthPreservedOnNormalMove(t *testing.T) {
	g := NewGame()
	// Place food far away so no eating happens.
	g.Food = Point{X: 0, Y: 0}
	initialLen := len(g.Snake)

	g.Tick()

	if len(g.Snake) != initialLen {
		t.Errorf("Snake length changed from %d to %d without eating", initialLen, len(g.Snake))
	}
}

func TestDirectionChangeAllowed(t *testing.T) {
	tests := []struct {
		name    string
		initial Direction
		attempt Direction
	}{
		{"right can go up", DirRight, DirUp},
		{"right can go down", DirRight, DirDown},
		{"left can go up", DirLeft, DirUp},
		{"left can go down", DirLeft, DirDown},
		{"up can go left", DirUp, DirLeft},
		{"up can go right", DirUp, DirRight},
		{"down can go left", DirDown, DirLeft},
		{"down can go right", DirDown, DirRight},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewGame()
			g.Dir = tt.initial
			g.SetDirection(tt.attempt)
			if g.Dir != tt.attempt {
				t.Errorf("Direction = %d after valid change, want %d",
					g.Dir, tt.attempt)
			}
		})
	}
}
