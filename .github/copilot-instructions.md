# Project Guidelines

## Architecture
- Keep `main.go` limited to CLI flag parsing and startup wiring.
- Keep Bubble Tea model flow and terminal rendering in `simulation/simulation.go`.
- Keep spring math and fixed-step physics helpers in `simulation/spring.go`.
- Prefer small, focused changes that preserve the existing package boundary: `main` starts the app, `simulation` owns behavior.

## Build and Test
- Use `go run main.go` for an interactive sanity check. The app supports `-count` and `-fps` flags.
- Use `go build ./...` to verify the workspace compiles after changes.
- Use `go test ./...` as the default validation command even though the repo currently has no test files.

## Conventions
- Preserve the Bubble Tea `Init`/`Update`/`View` structure and the `frameMsg` tick pattern unless the task specifically requires a different event flow.
- Keep rendering allocation-aware: this code precomputes colored ball cells and reuses a cached spaces string instead of rebuilding styles every frame.
- Treat the Charm dependency versions as coupled. The current `bubbletea`, `lipgloss`, `x/ansi`, and `x/term` versions in `go.mod` are a known-good set; do not pin or upgrade one in isolation.
- Be cautious when changing spring constants or frame timing. Motion behavior depends on the current `NewSpring(FPS(fps), 8.0, 0.35)` tuning.