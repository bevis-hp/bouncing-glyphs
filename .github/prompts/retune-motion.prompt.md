---
description: "Use when retuning ball motion, spring constants, or frame timing in the bouncing-balls simulation"
name: "Retune Motion"
argument-hint: "Describe the motion goal, for example: make the balls feel heavier and less jittery"
agent: "agent"
---
Retune the motion behavior of this workspace's bouncing ball simulation to match this goal: $ARGUMENTS

Constraints:
- Keep `main.go` limited to CLI startup wiring.
- Make motion and rendering changes inside [simulation/simulation.go](./simulation/simulation.go) and [simulation/spring.go](./simulation/spring.go) unless the request clearly requires something else.
- Preserve the Bubble Tea `Init`/`Update`/`View` flow and the existing `frameMsg` tick pattern.
- Avoid changing Charm dependency versions unless the task explicitly asks for it.
- Prefer the smallest parameter or logic changes that achieve the requested feel.

Workflow:
1. Inspect the current tuning points, especially `NewSpring(FPS(fps), 8.0, 0.35)`, target velocity ranges, and frame timing.
2. Explain briefly which parameters or behaviors you will adjust and why they map to the requested feel.
3. Implement the change.
4. Validate with `go test ./...` and `go build ./...`.
5. Summarize what changed, any motion tradeoffs, and any tuning knobs the user may want to adjust next.

Keep the response concise and grounded in the current codebase.