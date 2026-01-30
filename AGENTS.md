# Repository Guidelines

## Project Structure & Module Organization
- `cmd/roastgit/`: CLI entry point (`main.go`).
- `internal/`: core packages, split by concern:
  - `internal/git/`: git command execution + parsing (streaming log/numstat).
  - `internal/analyze/`: metrics, scoring, offenders.
  - `internal/render/`: text + JSON output.
  - `internal/roast/`: phrasing, intensity, censoring.
  - `internal/model/`: shared data structs.
  - `internal/util/`: helpers (sampling, spinner, string utilities).
- `internal/git/fixtures/`: test fixtures for parsers.
- `README.md`: usage and examples.

## Build, Test, and Development Commands
- `go build ./cmd/roastgit` — build the CLI binary.
- `go run ./cmd/roastgit` — run locally against the current repo.
- `go test ./...` — run all unit tests.

## Coding Style & Naming Conventions
- Go 1.22+; follow idiomatic Go formatting (`gofmt`).
- Indentation: tabs (Go standard).
- Package names are short and lowercase (`analyze`, `render`).
- Keep exported symbols minimal and documented only when needed.

## Testing Guidelines
- Framework: standard `testing` package.
- Tests live alongside code as `*_test.go`.
- Parser tests use fixtures in `internal/git/fixtures/` and avoid real git repos.
- Run: `go test ./...`.

## Commit & Pull Request Guidelines
- Current history contains only an "Initial commit"; no established convention yet.
- Suggested: imperative subject lines (e.g., "Add JSON renderer").
- PRs should include: purpose, key changes, and test command(s) run.
- For user-facing output changes, include a sample snippet in the PR description.

## Security & Configuration Notes
- Tool works offline; no network calls.
- Requires local `git` executable on PATH.
- Use `--deep` thoughtfully on large repos (slower by design).
