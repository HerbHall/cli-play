# cli-play -- Copilot Instructions

Interactive CLI playground showcasing terminal UI patterns with Bubbletea and Lipgloss.

## Tech Stack

- **Go** 1.25 -- primary language
- **Bubbletea** -- TUI framework (Elm architecture for terminals)
- **Lipgloss** -- terminal styling (colors, borders, layout)
- **Bubbles** -- reusable TUI components (spinners, text inputs, viewports)

## Project Structure

```text
cli-play/
├── cmd/              - Entry points and CLI wiring
├── internal/         - Private application packages
├── scripts/          - Build and utility scripts
├── .github/          - CI workflows and Copilot config
└── CLAUDE.md         - Claude Code instructions
```

## Code Style

- Conventional commits: `feat:`, `fix:`, `refactor:`, `docs:`, `test:`, `chore:`
- Co-author tag: `Co-Authored-By: GitHub Copilot <noreply@github.com>`
- Errors wrapped with context: `fmt.Errorf("operation: %w", err)`
- Table-driven tests with `t.Run` and descriptive names
- All lint checks must pass before committing (golangci-lint v2)

## Coding Guidelines

- Fix errors immediately -- never classify them as pre-existing
- Build, test, and lint must pass before any commit
- Never skip hooks (`--no-verify`) or force-push main
- Remove unused code completely; no backwards-compatibility hacks

## Available Resources

```bash
make build        # Compile the project
make test         # Run all tests
make lint         # Run golangci-lint
make run          # Run the application
go build ./...    # Direct build verification
go test ./...     # Direct test run
```

## Do NOT

- Add `//nolint` directives without fixing the root cause first
- Commit generated files without regenerating them first
- Add dependencies without updating the lock file
- Use `panic` in library code; return errors instead
- Store secrets, tokens, or credentials in code or config files
- Mark work as complete when known errors remain
