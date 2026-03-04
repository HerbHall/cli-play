---
applyTo: "**/*.go"
---

# Go Coding Instructions

## Error Handling

Always wrap errors with context using `fmt.Errorf`:

```go
// GOOD: wraps error with context
if err != nil {
    return fmt.Errorf("fetch device %s: %w", id, err)
}

// BAD: discards context
if err != nil {
    return err
}
```

Never return `(result, nil)` when `err != nil`. Propagate both:

```go
// GOOD: nilerr compliant
return &Result{Message: runErr.Error()}, fmt.Errorf("run: %w", runErr)
```

## Named Returns

When golangci-lint flags `unnamedResult`, add named returns. After adding names,
change `:=` to `=` for those variables and remove redundant `var` declarations:

```go
// GOOD: named returns, = not :=
func compute(data []float64) (score float64, detail string) {
    switch {
    case len(data) == 0:
        detail = "no data"
    default:
        score = average(data)
        detail = fmt.Sprintf("n=%d", len(data))
    }
    return score, detail
}
```

## Lint Patterns (golangci-lint v2)

Fix these before committing:

- **rangeValCopy**: `for _, v := range slice` on large structs.
  Use `for i := range slice { slice[i]... }`
- **prealloc**: `var result []T` in a loop.
  Use `make([]T, 0, len(source))`
- **appendCombine**: two consecutive `append()` to same slice.
  Combine into one call
- **emptyStringTest**: `len(s) > 0`. Use `s != ""`
- **httpNoBody**: use `http.NoBody` instead of `nil` body
- **bodyclose**: always close `resp.Body`. Deferred:
  `defer func() { _ = resp.Body.Close() }()`
- **exhaustive**: switch on enum types must list ALL cases
- **noctx**: use `ExecContext`, `QueryContext`,
  `QueryRowContext` instead of context-less variants
- **paramTypeCombine**: `(a int, b int)` becomes `(a, b int)`
- **builtinShadow**: don't shadow `new`, `make`, `len`,
  `cap`, `close`, `delete`, `copy`, `append`, `min`, `max`,
  `clear` as parameter names
- **gosec G101**: credential-adjacent constants need
  `//nolint:gosec // G101: <reason>`
- **preferFprint**: use `fmt.Fprintf(&b, ...)` instead of
  `b.WriteString(fmt.Sprintf(...))`
- **dupBranchBody**: identical `if`/`else` branches.
  Remove the conditional, keep just the body
- **sloppyReassign**: `if err = f(); err != nil` with named
  return `err`. Use `:=` to shadow instead

## Testing

Use table-driven tests with `t.Run`:

```go
func TestSomething(t *testing.T) {
    tests := []struct {
        name    string
        input   string
        want    int
        wantErr bool
    }{
        {name: "valid input", input: "abc", want: 3},
        {name: "empty input", input: "", wantErr: true},
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got, err := Something(tt.input)
            if (err != nil) != tt.wantErr {
                t.Errorf("error = %v, wantErr %v", err, tt.wantErr)
            }
            if got != tt.want {
                t.Errorf("got %v, want %v", got, tt.want)
            }
        })
    }
}
```

## HTTP Handlers (Go 1.22+)

Use Go 1.22 enhanced `ServeMux` patterns:

```go
mux.HandleFunc("GET /items/{id}", h.getItem)
mux.HandleFunc("POST /items", h.createItem)
```

Always name `*http.Request` parameter even if unused.
Expanding stubs later requires it.

## Imports

Group imports in three blocks separated by blank lines:

```go
import (
    "context"
    "fmt"

    "github.com/external/pkg"

    "github.com/yourorg/project/internal/models"
)
```
