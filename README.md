# Bouncing Glyphs

Terminal physics toy of bouncing glyphs built with Go and the brilliant TUI framework from Charm - BubbleTea

![Demo](demo/demo.gif)

## Features

- Fixed-timestep terminal animation loop
- Responsive to resizing of terminal
- Gravity + Coefficient of Restitution + Floor Friction
- Spring-following horizontal drift for smooth lateral motion
- Random glyph characters and colours
- Interactive controls to spawn more glyphs
- NEW: Collision flag added 
- NEW: Streaming glyph intake from piped stdin

## Requirements

- Go 1.26+
- A terminal with ANSI color support

## Install

```bash
go install github.com/bevis-hp/bouncing-glyphs@latest
```

Then run:

```bash
bouncing-glyphs
```

## Run

```bash
go run main.go
```

Pipe stdout from another command into the simulation and keep the shell command visible above the animation:

```bash
grep -o . README.md | go run main.go -count 0
```

## CLI Reference (Auto-generated)

This section is refreshed by `scripts/update_readme.sh`.

<!-- BEGIN AUTO-CLI -->
```text
Usage of bouncing-glyphs:
  -collision
    	enable glyph-glyph collisions (higher CPU cost)
  -count int
    	number of glyphs to simulate (default 10)
  -despawn
    	despawn resting glyphs after rest-timeout
  -fps int
    	frames per second (default 60)
  -gravity float
    	downward acceleration in cells/frame^2 (default 0.008)
  -launch-kick-max float
    	max upward launch speed for initial glyphs (default 0.6)
  -rest-threshold float
    	speed below which glyphs are considered resting (default 0.08)
  -rest-timeout float
    	seconds at rest before glyph despawns (default 5)
  -restitution float
	    	bounce speed retention fraction (default 0.375)
  -spawn-kick-max float
    	max upward launch speed for spawned glyphs (default 1)
  -spring-damping float
    	x-axis spring damping ratio (default 0.55)
  -spring-frequency float
    	x-axis spring angular frequency (default 5)
  -stdin-interval-ms int
    	milliseconds between glyph spawns from piped stdin (default 100)
  -stdin-drop-delay-ms int
    	milliseconds stdin glyphs wait at top before dropping (default 200)
  -target-drift-max float
    	max x-target drift speed magnitude (default 0.7)
  -x-floor-friction float
    	horizontal drift retention on floor bounce (default 0.96)
```
<!-- END AUTO-CLI -->

## Controls

- `Space`: spawn a new glyph
- `q`, `esc`, or `ctrl+c`: quit

When stdin is piped in, Bubble Tea runs in the normal screen buffer instead of the alternate screen so the original shell command remains above the animation. Keyboard input is read from the controlling TTY, while piped stdin is buffered and released into the simulation one glyph at a time.

## Development

Build:

```bash
go build ./...
```

Test:

```bash
go test ./...
```

Install repo hooks once for this clone:

```bash
git config core.hooksPath .githooks
```

What the pre-push hook does:

- Regenerates the README CLI section from `go run . -h`
- Stages `README.md` and blocks push if that section changed (commit, then push again)
- Syncs GitHub description/topics from `.github/repo-metadata.env`

## License

This project is licensed under the GNU General Public License v3.0. See [LICENSE](LICENSE) for details.
