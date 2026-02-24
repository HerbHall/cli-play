# cli-play

Terminal game collection built with Go + [Bubbletea](https://github.com/charmbracelet/bubbletea). Quick games to play in VS Code's integrated terminal while waiting for builds, CI, etc.

## Build and Run

```bash
go build -o cli-play.exe ./cmd/cli-play   # build
go run ./cmd/cli-play                      # run directly
go test ./...                              # run all tests
```

## Project Structure

```text
cli-play/
├── cmd/cli-play/          # Entry point
│   └── main.go
├── internal/
│   ├── menu/              # Game launcher menu
│   ├── yahtzee/           # Yahtzee dice game
│   ├── blackjack/         # Blackjack card game
│   ├── wordle/            # Word guessing game
│   ├── minesweeper/       # Grid-based mine sweeper
│   ├── sudoku/            # Number puzzle
│   └── twofortyeight/     # 2048 sliding tile game
└── go.mod
```

## Architecture

Each game is a self-contained Bubbletea `tea.Model` in its own package under `internal/`. The menu package provides the launcher UI. Games return to the menu when finished or when the user quits.

Pattern for each game package:

- `model.go` -- `tea.Model` implementation (Init, Update, View)
- `game.go` -- Game logic (no UI concerns)
- `game_test.go` -- Tests for game logic

## v1 Games

1. Yahtzee -- dice rolling, scoring categories
2. Blackjack -- card rounds
3. Wordle -- 5-letter word guessing
4. Minesweeper -- grid reveal with flagging
5. Sudoku -- number puzzle with difficulty levels
6. 2048 -- slide and merge tiles

## Conventions

- Bubbletea for all TUI rendering and input handling
- Game logic separated from UI (testable without tea.Model)
- Each game is independent -- no shared state between games
- Use [Lip Gloss](https://github.com/charmbracelet/lipgloss) for styling
