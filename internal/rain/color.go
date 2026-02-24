package rain

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
)

// Highlight and deposited colours used by the splash/menu renderer.
var (
	ColorHighlight = lipgloss.Color("#FFD700") // gold for highlight chars
	ColorDeposited = lipgloss.Color("#FFFFFF") // white for revealed menu chars
)

// 4-stop gradient from head (brightest) to tail (dimmest).
type colorStop struct {
	pos     float64
	r, g, b uint8
}

var trailStops = [4]colorStop{
	{0.00, 0xDC, 0xFF, 0xDC}, // head:       #DCFFDC (white-green)
	{0.15, 0x00, 0xE6, 0x32}, // body bright: #00E632
	{0.50, 0x00, 0x96, 0x1E}, // body mid:    #00961E
	{1.00, 0x00, 0x3C, 0x0F}, // tail:        #003C0F
}

// TrailColor returns an interpolated lipgloss.Color for a position
// in the trail. 0.0 is the head (brightest), 1.0 is the tail (dimmest).
func TrailColor(position float64) lipgloss.Color {
	if position <= 0 {
		return lipgloss.Color("#DCFFDC")
	}
	if position >= 1 {
		return lipgloss.Color("#003C0F")
	}

	// Find the two surrounding stops.
	var lo, hi colorStop
	for i := 0; i < len(trailStops)-1; i++ {
		if position >= trailStops[i].pos && position <= trailStops[i+1].pos {
			lo = trailStops[i]
			hi = trailStops[i+1]
			break
		}
	}

	// Normalise position within the segment [lo.pos .. hi.pos].
	span := hi.pos - lo.pos
	if span == 0 {
		return hexColor(lo.r, lo.g, lo.b)
	}
	t := (position - lo.pos) / span

	r := lerp8(lo.r, hi.r, t)
	g := lerp8(lo.g, hi.g, t)
	b := lerp8(lo.b, hi.b, t)

	return hexColor(r, g, b)
}

// lerp8 linearly interpolates between two uint8 values.
func lerp8(a, b uint8, t float64) uint8 {
	return uint8(float64(a)*(1-t) + float64(b)*t)
}

// hexColor formats RGB components as a lipgloss hex colour string.
func hexColor(r, g, b uint8) lipgloss.Color {
	return lipgloss.Color(fmt.Sprintf("#%02X%02X%02X", r, g, b))
}
