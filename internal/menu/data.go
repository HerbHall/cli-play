package menu

// gameIcon maps a game index (from the Games slice) to a display icon.
var gameIcon = map[int]string{
	0:  "\u2684",       // Yahtzee: die face 5
	1:  "\u2660",       // Blackjack: spade
	2:  "Aa",           // Wordle
	3:  "\u2731",       // Minesweeper: heavy asterisk
	4:  "#",            // Sudoku
	5:  "\u229e",       // 2048: squared plus
	6:  "\u2620",       // Hangman: skull
	7:  "\u2715",       // Tic-Tac-Toe: multiplication X
	8:  "\u25cf",       // Mastermind: black circle
	9:  "\u2666\u2666", // Memory: diamonds
	10: "\u25c9",       // Connect Four: fisheye
	11: "\u22a1",       // Fifteen Puzzle: squared dot
	12: "~",            // Snake
	13: "\u259f",       // Tetris: quadrant lower right
	14: "\u2663",       // Solitaire: club
	15: "\u2328",       // Typing Test: keyboard
}

// category groups games by theme. Each category references game indices
// from the original Games slice (preserving launch indices).
type category struct {
	Name    string
	Icon    string
	Indices []int
}

var categories = []category{
	{Name: "Classic", Icon: "\U0001f3b2", Indices: []int{0, 1, 14}},          // Yahtzee, Blackjack, Solitaire
	{Name: "Puzzles", Icon: "\U0001f9e9", Indices: []int{2, 3, 4, 5, 11, 8}}, // Wordle, Minesweeper, Sudoku, 2048, Fifteen Puzzle, Mastermind
	{Name: "Action", Icon: "\U0001f3ae", Indices: []int{12, 13}},             // Snake, Tetris
	{Name: "Strategy", Icon: "\U0001f0cf", Indices: []int{6, 7, 10, 9}},      // Hangman, Tic-Tac-Toe, Connect Four, Memory
	{Name: "Skills", Icon: "\U0001f3af", Indices: []int{15}}, // Typing Test (dart/target -- consistent 2-cell emoji)
}

// gamePreview holds the info panel text for each game.
type gamePreview struct {
	Rules    string
	Controls string
}

var previews = map[int]gamePreview{
	0: {
		Rules:    "Roll 5 dice up to 3 times per turn.\nFill 13 scoring categories.\nHighest total score wins.",
		Controls: "Space: roll | 1-5: hold dice | Enter: score",
	},
	1: {
		Rules:    "Get closer to 21 than the dealer\nwithout going over. Aces are 1 or 11.\nBlackjack (21 in 2 cards) pays extra.",
		Controls: "H: hit | S: stand | D: double down",
	},
	2: {
		Rules:    "Guess a 5-letter word in 6 tries.\nGreen = right letter, right spot.\nYellow = right letter, wrong spot.",
		Controls: "Type letters | Enter: submit | Backspace",
	},
	3: {
		Rules:    "Uncover all safe cells without\nhitting a mine. Numbers show how\nmany mines are adjacent.",
		Controls: "Arrows: move | Space: reveal | F: flag",
	},
	4: {
		Rules:    "Fill every row, column, and 3x3 box\nwith digits 1-9. No repeats allowed\nin any group.",
		Controls: "Arrows: move | 1-9: place | 0: clear",
	},
	5: {
		Rules:    "Slide tiles to merge matching\nnumbers. Reach 2048 to win.\nGame over when no moves remain.",
		Controls: "Arrow keys to slide all tiles",
	},
	6: {
		Rules:    "Guess the hidden word one letter\nat a time. Too many wrong guesses\nand it's game over.",
		Controls: "A-Z: guess a letter",
	},
	7: {
		Rules:    "Place X's on a 3x3 grid. Get three\nin a row to win. Block the AI's\nattempts to do the same.",
		Controls: "Arrows: move | Enter: place mark",
	},
	8: {
		Rules:    "Deduce the secret 4-color code.\nBlack peg = right color, right spot.\nWhite peg = right color, wrong spot.",
		Controls: "1-6: pick color | Enter: submit guess",
	},
	9: {
		Rules:    "Flip cards to find matching pairs.\nRemember positions to clear the\nboard in the fewest moves.",
		Controls: "Arrows: move | Enter/Space: flip card",
	},
	10: {
		Rules:    "Drop colored discs into columns.\nFirst to connect 4 in a row\n(horizontal, vertical, diagonal) wins.",
		Controls: "Left/Right: choose column | Enter: drop",
	},
	11: {
		Rules:    "Slide numbered tiles to arrange\nthem 1-15 in order. Use the empty\nspace to maneuver.",
		Controls: "Arrow keys to slide adjacent tile",
	},
	12: {
		Rules:    "Guide the snake to eat food and\ngrow longer. Don't crash into walls\nor your own tail.",
		Controls: "Arrow keys: direction | P: pause",
	},
	13: {
		Rules:    "Rotate and place falling blocks to\ncomplete rows. Completed rows are\ncleared for points.",
		Controls: "Arrows: move | Up: rotate | Space: drop",
	},
	14: {
		Rules:    "Move cards between tableau piles and\nfoundations. Build foundations up by\nsuit from Ace to King.",
		Controls: "Arrows: move | Enter: select/place",
	},
	15: {
		Rules:    "Type the displayed text as fast and\naccurately as you can. Your WPM\nand accuracy are measured.",
		Controls: "Type the highlighted text",
	},
}

// tips shown in the rotating ticker at the bottom of the menu.
var tips = []string{
	"Tip: Press P to pause in Snake and Tetris",
	"Tip: In Wordle, green = correct position, yellow = wrong position",
	"Tip: In Minesweeper, flag suspected mines with F",
	"Tip: 2048 strategy -- keep your largest tile in a corner",
	"Tip: Sudoku -- scan rows, columns, and boxes to eliminate candidates",
	"Tip: In Blackjack, stand on 17+ against a dealer showing 2-6",
	"Tip: Yahtzee -- go for the bonus by scoring 63+ in the upper section",
	"Tip: Tic-Tac-Toe -- take the center square first",
	"Tip: In Mastermind, use your first guesses to narrow down colors",
	"Tip: Memory -- flip cards methodically, row by row",
	"Tip: Connect Four -- control the center columns for more options",
	"Tip: Typing Test -- focus on accuracy first, speed follows",
	"Tip: In Hangman, guess vowels first to reveal the word structure",
	"Tip: Press number keys 1-9 or 0/a-f to quick-select games",
}

// shortcutKeys maps the display position (0-based) within the flattened
// category list to a shortcut label. Games 1-9 use "1"-"9", game 10
// uses "0", games 11-16 use "a"-"f".
func shortcutLabel(displayIndex int) string {
	if displayIndex < 9 {
		return string(rune('1' + displayIndex))
	}
	if displayIndex == 9 {
		return "0"
	}
	return string(rune('a' + displayIndex - 10))
}
