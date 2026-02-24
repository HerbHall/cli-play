package rain

import "math/rand/v2"

// CharPool holds a set of runes available for rain characters.
type CharPool struct {
	Chars []rune
}

// NewMatrixPool returns a CharPool with half-width katakana,
// digits, and a handful of symbols -- the classic Matrix look.
func NewMatrixPool() *CharPool {
	var chars []rune

	// Half-width katakana: U+FF66 to U+FF9F (58 chars)
	for r := rune(0xFF66); r <= 0xFF9F; r++ {
		chars = append(chars, r)
	}

	// Digits 0-9
	for r := '0'; r <= '9'; r++ {
		chars = append(chars, r)
	}

	// Symbols
	for _, r := range []rune{'<', '>', '=', '+', '-', '*', ':', '.', '|'} {
		chars = append(chars, r)
	}

	return &CharPool{Chars: chars}
}

// Random returns a random rune from the pool.
func (p *CharPool) Random() rune {
	return p.Chars[rand.IntN(len(p.Chars))]
}
