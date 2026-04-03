# Bouncing Glyphs

A terminal animation built with Go and Bubble Tea where colorful glyphs bounce under gravity, drift horizontally with spring-like motion, and eventually despawn after coming to rest.

## Features

- Fixed-timestep terminal animation loop
- Gravity + restitution floor bounces
- Spring-following horizontal drift for smooth lateral motion
- Random glyph characters and colors
- Interactive controls to spawn more glyphs and quit cleanly

## Requirements

- Go 1.26+
- A terminal with ANSI color support

## Run

```bash
go run main.go
```

## CLI Reference (Auto-generated)

This section is refreshed by `scripts/update_readme.sh`.

<!-- BEGIN AUTO-CLI -->
```text
Usage of bouncing-glyphs:
  -count int
    	number of glyphs to simulate (default 10)
  -fps int
    	frames per second (default 60)
```
<!-- END AUTO-CLI -->

## Controls

- `Space`: spawn a new glyph
- `q`, `esc`, or `ctrl+c`: quit

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
