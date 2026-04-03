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

## Flags

- `-count` (default `10`): initial number of glyphs
- `-fps` (default `60`): simulation frame rate

Example:

```bash
go run main.go -count 20 -fps 75
```

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

## License

This project is licensed under the GNU General Public License v3.0. See [LICENSE](LICENSE) for details.
