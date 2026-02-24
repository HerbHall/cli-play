package transition

import "time"

// Phase tracks which stage of the transition animation is active.
type Phase int

const (
	PhaseDissolve Phase = iota // Rain starts, eats splash text
	PhaseReveal                // Rain deposits menu characters
	PhaseDrain                 // Rain fades out
	PhaseDone                  // Hand off to menu
)

const (
	// TickInterval is the animation frame duration (~30 FPS).
	TickInterval = 33 * time.Millisecond

	// SpawnRate is the base probability per empty column per frame
	// of spawning a new rain column.
	SpawnRate = 0.15
)
