---
description: "Use when editing simulation package files, spring tuning, Bubble Tea frame updates, or rendering logic in simulation/**/*.go"
name: "Simulation Guidelines"
applyTo: "simulation/**/*.go"
---

# Simulation Guidelines

- Keep the Bubble Tea model flow centered in `Simulation.Init`, `Simulation.Update`, and `Simulation.View`; do not move simulation behavior into `main.go`.
- Preserve the fixed-step `frameMsg` tick pattern unless the task explicitly requires a different event flow.
- Treat `NewSpring(FPS(fps), 8.0, 0.35)` as the current motion baseline. Prefer small, explainable tuning changes over broader physics rewrites.
- Keep rendering allocation-aware: reuse cached strings and precomputed ball styling instead of rebuilding styles or large buffers every frame.
- When changing drawing logic, preserve the footer row behavior and avoid introducing per-frame work that scales poorly with terminal size.
- Keep spring math numerically stable. If you change damping or coefficient logic in `spring.go`, validate that all damping regimes still behave correctly.
- Do not change Charm dependency versions from simulation work unless the task is explicitly about dependency maintenance.