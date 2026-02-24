# CLI Play

A collection of classic terminal games built with Go and the
[Charm](https://charm.sh/) ecosystem. Features an animated digital rain
splash screen and a game selection menu.

## Games

| Game | Status |
|------|--------|
| Yahtzee | Playable |
| Blackjack | Playable |
| Wordle | Coming soon |
| Minesweeper | Coming soon |
| Sudoku | Coming soon |
| 2048 | Coming soon |

## Install

```bash
go install github.com/herbhall/cli-play/cmd/cli-play@latest
```

## Usage

```bash
cli-play
```

Or run from source:

```bash
go run ./cmd/cli-play
```

## Build

```bash
go build -o cli-play ./cmd/cli-play
```

## Requirements

- Go 1.23+
- A terminal with 256-color support (most modern terminals)

## Credits

Built by [Herb Hall](https://github.com/herbhall) and
[Claude Code](https://claude.ai/claude-code) using the
[Charm](https://charm.sh/) ecosystem
([Bubble Tea](https://github.com/charmbracelet/bubbletea),
[Lip Gloss](https://github.com/charmbracelet/lipgloss)).
