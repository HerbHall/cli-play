package memory

import "math/rand/v2"

const (
	rows     = 4
	cols     = 4
	numPairs = rows * cols / 2
)

// CardState represents the visibility state of a card.
type CardState int

const (
	FaceDown CardState = iota
	FaceUp
	Matched
)

// Card represents a single card on the board.
type Card struct {
	Symbol byte
	State  CardState
}

// Game holds the complete state of a memory game.
type Game struct {
	Board      [rows][cols]Card
	FirstPick  [2]int // row, col of first flipped card (-1 if none)
	SecondPick [2]int // row, col of second flipped card (-1 if none)
	HasFirst   bool
	HasSecond  bool
	Moves      int
	PairsFound int
	GameOver   bool
}

// NewGame creates a fresh game with shuffled cards.
func NewGame() *Game {
	g := &Game{
		FirstPick:  [2]int{-1, -1},
		SecondPick: [2]int{-1, -1},
	}
	g.shuffle()
	return g
}

// NewGameWithBoard creates a game with a specific board layout (for testing).
func NewGameWithBoard(board [rows][cols]byte) *Game {
	g := &Game{
		FirstPick:  [2]int{-1, -1},
		SecondPick: [2]int{-1, -1},
	}
	for r := 0; r < rows; r++ {
		for c := 0; c < cols; c++ {
			g.Board[r][c] = Card{Symbol: board[r][c], State: FaceDown}
		}
	}
	return g
}

// shuffle places pairs of symbols A-H randomly on the board.
func (g *Game) shuffle() {
	symbols := make([]byte, 0, rows*cols)
	for i := 0; i < numPairs; i++ {
		ch := byte('A' + i)
		symbols = append(symbols, ch, ch)
	}
	rand.Shuffle(len(symbols), func(i, j int) {
		symbols[i], symbols[j] = symbols[j], symbols[i]
	})
	idx := 0
	for r := 0; r < rows; r++ {
		for c := 0; c < cols; c++ {
			g.Board[r][c] = Card{Symbol: symbols[idx], State: FaceDown}
			idx++
		}
	}
}

// FlipCard attempts to flip a card at the given position.
// Returns true if the card was flipped.
func (g *Game) FlipCard(row, col int) bool {
	if row < 0 || row >= rows || col < 0 || col >= cols {
		return false
	}
	if g.GameOver || g.HasSecond {
		return false
	}

	card := &g.Board[row][col]
	if card.State != FaceDown {
		return false
	}

	card.State = FaceUp

	if !g.HasFirst {
		g.FirstPick = [2]int{row, col}
		g.HasFirst = true
		return true
	}

	// Second card flipped
	g.SecondPick = [2]int{row, col}
	g.HasSecond = true
	g.Moves++
	return true
}

// CheckMatch evaluates whether the two flipped cards match.
// Returns true if they match, false otherwise.
// Requires both cards to be flipped (HasSecond == true).
func (g *Game) CheckMatch() bool {
	if !g.HasSecond {
		return false
	}

	first := g.Board[g.FirstPick[0]][g.FirstPick[1]]
	second := g.Board[g.SecondPick[0]][g.SecondPick[1]]

	return first.Symbol == second.Symbol
}

// ResolveMatch marks the two flipped cards as matched and resets picks.
func (g *Game) ResolveMatch() {
	if !g.HasSecond {
		return
	}
	g.Board[g.FirstPick[0]][g.FirstPick[1]].State = Matched
	g.Board[g.SecondPick[0]][g.SecondPick[1]].State = Matched
	g.PairsFound++
	g.resetPicks()
	if g.PairsFound == numPairs {
		g.GameOver = true
	}
}

// ResolveNoMatch flips the two non-matching cards back face-down and resets picks.
func (g *Game) ResolveNoMatch() {
	if !g.HasSecond {
		return
	}
	g.Board[g.FirstPick[0]][g.FirstPick[1]].State = FaceDown
	g.Board[g.SecondPick[0]][g.SecondPick[1]].State = FaceDown
	g.resetPicks()
}

// resetPicks clears the current pick state.
func (g *Game) resetPicks() {
	g.FirstPick = [2]int{-1, -1}
	g.SecondPick = [2]int{-1, -1}
	g.HasFirst = false
	g.HasSecond = false
}

// TotalPairs returns the number of pairs in the game.
func (g *Game) TotalPairs() int {
	return numPairs
}

// Rows returns the number of rows on the board.
func (g *Game) Rows() int {
	return rows
}

// Cols returns the number of columns on the board.
func (g *Game) Cols() int {
	return cols
}
