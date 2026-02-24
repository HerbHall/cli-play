package rain

import "math/rand/v2"

// TrailChar is a single visible character in a rain column's trail.
type TrailChar struct {
	Y    int
	Char rune
}

// Column is a vertical strip of falling rain characters.
// Movement uses an accumulator so sub-row speeds work smoothly.
type Column struct {
	X            int
	Trail        []TrailChar
	MaxTrailLen  int
	Speed        float64 // rows per second (8.0 - 25.0)
	Accumulator  float64 // fractional row movement
	HeadY        float64 // current head position (can be negative = above screen)
	Draining     bool    // head past bottom, removing tail chars
	MutationRate float64 // chance per char per update to swap character
}

// SpawnColumn creates a new column at the given x coordinate.
// HeadY starts randomly between -(screenHeight/2) and 0 so columns
// stagger their entrance. Speed and trail length are randomised.
func SpawnColumn(x, screenHeight int) *Column {
	half := screenHeight / 2
	startY := -rand.IntN(half + 1) // [-(half) .. 0]

	minTrail := screenHeight / 3
	if minTrail < 4 {
		minTrail = 4
	}
	maxTrail := screenHeight
	trailLen := minTrail + rand.IntN(maxTrail-minTrail+1)

	speed := 8.0 + rand.Float64()*17.0 // [8.0 .. 25.0)

	return &Column{
		X:            x,
		Trail:        nil,
		MaxTrailLen:  trailLen,
		Speed:        speed,
		Accumulator:  0,
		HeadY:        float64(startY),
		Draining:     false,
		MutationRate: 0.02,
	}
}

// Update advances the column by dt seconds.
func (c *Column) Update(dt float64, screenHeight int, pool *CharPool) {
	c.Accumulator += c.Speed * dt

	if c.Draining {
		c.drain()
		return
	}

	c.advance(screenHeight, pool)
	c.trimTrail()
	c.mutate(pool)
}

// advance grows the trail downward while accumulator has whole steps.
func (c *Column) advance(screenHeight int, pool *CharPool) {
	for c.Accumulator >= 1.0 {
		c.Accumulator -= 1.0

		y := int(c.HeadY)
		if y >= 0 && y < screenHeight {
			c.Trail = append(c.Trail, TrailChar{
				Y:    y,
				Char: pool.Random(),
			})
		}

		c.HeadY += 1.0

		if int(c.HeadY) >= screenHeight {
			c.Draining = true
			return
		}
	}
}

// drain removes one tail character per accumulated step.
func (c *Column) drain() {
	for c.Accumulator >= 1.0 && len(c.Trail) > 0 {
		c.Accumulator -= 1.0
		c.Trail = c.Trail[1:] // remove oldest (tail end)
	}
}

// trimTrail enforces the maximum trail length by dropping the oldest chars.
func (c *Column) trimTrail() {
	if len(c.Trail) > c.MaxTrailLen {
		c.Trail = c.Trail[len(c.Trail)-c.MaxTrailLen:]
	}
}

// mutate randomly replaces characters in the trail.
func (c *Column) mutate(pool *CharPool) {
	for i := range c.Trail {
		if rand.Float64() < c.MutationRate {
			c.Trail[i].Char = pool.Random()
		}
	}
}

// IsDead returns true when the column has drained completely.
func (c *Column) IsDead() bool {
	return c.Draining && len(c.Trail) == 0
}
