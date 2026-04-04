package simulation

import (
	"bufio"
	"fmt"
	"io"
	"math"
	"math/rand"
	"sort"
	"strings"
	"time"
	"unicode"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// PhysicsConfig controls spring tuning and y-axis bounce behavior.
type PhysicsConfig struct {
	Gravity            float64 // cells/frame² downward acceleration
	Restitution        float64 // fraction of speed retained on floor bounce
	XFloorFriction     float64 // fraction of target x velocity retained on floor bounce
	EnableCollision    bool    // enables glyph-glyph collisions (extra CPU work)
	EnableDespawn      bool    // enables resting glyph despawn after timeout
	RestThreshold      float64 // cells/frame speed below which a glyph is considered at rest
	RestTimeoutSeconds float64 // seconds at rest before despawn
	SpringFrequency    float64 // spring angular frequency for x-axis follow behavior
	SpringDampingRatio float64 // spring damping ratio for x-axis follow behavior
	LaunchKickMax      float64 // max upward launch speed magnitude for initial glyphs
	SpawnKickMax       float64 // max upward launch speed magnitude for spawned glyphs
	TargetDriftMax     float64 // max absolute x target drift speed
}

// DefaultPhysicsConfig returns the baseline motion tuning.
func DefaultPhysicsConfig() PhysicsConfig {
	return PhysicsConfig{
		Gravity:            0.008,
		Restitution:        0.375,
		XFloorFriction:     0.96,
		EnableCollision:    false,
		EnableDespawn:      false,
		RestThreshold:      0.08,
		RestTimeoutSeconds: 25.0,
		SpringFrequency:    5.0,
		SpringDampingRatio: 0.55,
		LaunchKickMax:      0.6,
		SpawnKickMax:       1.0,
		TargetDriftMax:     0.7,
	}
}

// Glyph represents a glyph in the simulation.
// The X axis uses a spring follower for a drifting, slightly-lagged feel.
// The Y axis uses direct Euler integration so gravity and floor bounces are
// lag-free and land exactly on the floor row.
type Glyph struct {
	x, y       float64 // displayed position
	xVel       float64 // spring internal velocity (x axis)
	yVel       float64 // physics velocity (y axis, cells/frame)
	targetX    float64 // spring equilibrium point (x only)
	targetXVel float64 // wandering velocity of the x target
	holdFor    float64 // seconds to remain pinned before gravity applies
	stillFor   float64 // seconds spent at rest (for despawn)
	char       rune    // display character
	ansi       string  // pre-computed colored cell string
}

// Simulation manages the bouncing glyphs.
type Simulation struct {
	width, height int
	ready         bool
	glyphs        []*Glyph
	count         int
	pendingChars  []rune
	streamCursor  int
	rng           *rand.Rand
	physics       PhysicsConfig
	spring        Spring
	stream        StreamConfig
	frameDuration time.Duration
	spaces        string // s.width spaces, rebuilt on resize
}

type frameMsg time.Time
type streamSpawnMsg time.Time

type stdinGlyphMsg struct {
	char rune
}

// StreamConfig controls stdin-driven glyph spawning.
type StreamConfig struct {
	Reader        io.Reader
	SpawnInterval time.Duration
	DropDelay     time.Duration
}

// These are sampled once per glyph so every instance keeps a stable appearance.
var glyphChars = []rune("ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789!@#$%^&*()-_=+[]{}|;:'\",.<>/?~£€¢¥§±×÷¶©®™✓✕★☆◆◇●○■□▲△✦✧✩✪✫✬✭✮✯✰✶✷")
var glyphColors = []lipgloss.Color{
	// Dracula
	"#FF5555", // red
	"#50FA7B", // green
	"#F1FA8C", // yellow
	"#BD93F9", // purple
	"#FF79C6", // pink
	"#8BE9FD", // cyan
	"#FFB86C", // orange
	"#F8F8F2", // foreground

	// Nord
	"#88C0D0", // frost cyan
	"#81A1C1", // frost blue
	"#5E81AC", // frost deep blue
	"#A3BE8C", // aurora green
	"#EBCB8B", // aurora yellow
	"#D08770", // aurora orange
	"#BF616A", // aurora red
	"#B48EAD", // aurora purple

	// Solarized
	"#268BD2", // blue
	"#2AA198", // cyan
	"#859900", // green
	"#B58900", // yellow
	"#CB4B16", // orange
	"#DC322F", // red
	"#D33682", // magenta
	"#6C71C4", // violet

	// Monokai
	"#F92672", // pink
	"#A6E22E", // green
	"#FD971F", // orange
	"#66D9EF", // cyan
	"#AE81FF", // purple
	"#E6DB74", // yellow
	"#F8F8F2", // foreground
}

// New creates a new Simulation. Width and height are determined from the
// terminal on the first WindowSizeMsg.
func New(count, fps int) *Simulation {
	return NewWithPhysicsConfig(count, fps, DefaultPhysicsConfig())
}

// NewWithPhysicsConfig creates a new Simulation using caller-provided physics tuning.
func NewWithPhysicsConfig(count, fps int, physics PhysicsConfig) *Simulation {
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	if fps < 1 {
		fps = 1
	}
	// The spring is precomputed once because its coefficients only depend on the
	// frame step and configured damping values.
	deltaTime := FPS(fps)
	spring := NewSpring(deltaTime, physics.SpringFrequency, physics.SpringDampingRatio)
	glyphs := make([]*Glyph, count)
	for i := range glyphs {
		color := glyphColors[rng.Intn(len(glyphColors))]
		char := glyphChars[rng.Intn(len(glyphChars))]
		glyphs[i] = &Glyph{
			char: char,
			ansi: lipgloss.NewStyle().Foreground(color).Render(string(char)),
		}
	}
	return &Simulation{
		count:         count,
		rng:           rng,
		glyphs:        glyphs,
		physics:       physics,
		spring:        spring,
		frameDuration: time.Second / time.Duration(fps),
	}
}

// EnableStream configures the simulation to read glyphs from a stream and
// spawn them at a fixed interval.
func (sim *Simulation) EnableStream(reader io.Reader, spawnInterval, dropDelay time.Duration) {
	if reader == nil {
		return
	}
	if spawnInterval <= 0 {
		spawnInterval = 100 * time.Millisecond
	}
	if dropDelay < 0 {
		dropDelay = 0
	}

	sim.stream = StreamConfig{
		Reader:        reader,
		SpawnInterval: spawnInterval,
		DropDelay:     dropDelay,
	}
}

// placeGlyphs launches all glyphs from the top of the screen.
func (sim *Simulation) placeGlyphs() {
	for _, glyph := range sim.glyphs {
		startX := sim.rng.Float64() * float64(sim.width-1)
		// Start each glyph at the top edge, then give it a small upward kick so it
		// immediately falls back into the playfield under gravity.
		glyph.x = startX
		glyph.y = 0
		glyph.xVel = 0
		glyph.yVel = -(sim.rng.Float64() * sim.physics.LaunchKickMax)
		glyph.targetX = startX
		glyph.targetXVel = (sim.rng.Float64()*2 - 1) * sim.physics.TargetDriftMax
		glyph.stillFor = 0
	}
}

// spawnGlyph adds a new glyph at a random x position at the top of the screen.
func (sim *Simulation) spawnGlyph() {
	sim.spawnGlyphWithCharAtX(glyphChars[sim.rng.Intn(len(glyphChars))], sim.rng.Float64()*float64(sim.width-1))
}

func (sim *Simulation) spawnGlyphWithChar(char rune) {
	sim.spawnGlyphWithCharAtX(char, sim.rng.Float64()*float64(sim.width-1))
}

func (sim *Simulation) spawnGlyphWithCharAtX(char rune, x float64) {
	sim.spawnGlyphWithCharAtXAndDelay(char, x, 0)
}

func (sim *Simulation) spawnGlyphWithCharAtXAndDelay(char rune, x float64, holdDelay time.Duration) {
	if sim.width < 1 {
		return
	}
	color := glyphColors[sim.rng.Intn(len(glyphColors))]
	if x < 0 {
		x = 0
	}
	if x > float64(sim.width-1) {
		x = float64(sim.width - 1)
	}
	glyph := &Glyph{
		x:          x,
		y:          0,
		yVel:       -(sim.rng.Float64() * sim.physics.SpawnKickMax),
		targetX:    x,
		targetXVel: (sim.rng.Float64()*2 - 1) * sim.physics.TargetDriftMax,
		holdFor:    holdDelay.Seconds(),
		char:       char,
		ansi:       lipgloss.NewStyle().Foreground(color).Render(string(char)),
	}
	sim.glyphs = append(sim.glyphs, glyph)
}

func (sim *Simulation) animate() tea.Cmd {
	// Bubble Tea drives the simulation by scheduling the next fixed-timestep tick.
	return tea.Tick(sim.frameDuration, func(tickTime time.Time) tea.Msg {
		return frameMsg(tickTime)
	})
}

func (sim *Simulation) streamSpawnTick() tea.Cmd {
	if sim.stream.Reader == nil {
		return nil
	}

	return tea.Tick(sim.stream.SpawnInterval, func(tickTime time.Time) tea.Msg {
		return streamSpawnMsg(tickTime)
	})
}

func (sim *Simulation) readStream(program *tea.Program) {
	reader := bufio.NewReader(sim.stream.Reader)
	for {
		char, _, err := reader.ReadRune()
		if err != nil {
			return
		}
		if !shouldQueueStreamRune(char) {
			continue
		}

		program.Send(stdinGlyphMsg{char: char})
	}
}

func shouldQueueStreamRune(char rune) bool {
	if char == '\n' || char == '\r' || char == '\t' || char == ' ' {
		return true
	}

	return unicode.IsGraphic(char)
}

func (sim *Simulation) advanceStreamCursor(cells int) {
	if sim.width < 1 || cells <= 0 {
		return
	}

	sim.streamCursor += cells
	if sim.streamCursor >= sim.width {
		sim.streamCursor %= sim.width
	}
}

// Run starts the Bubble Tea simulation.
func (sim *Simulation) Run() {
	programOptions := make([]tea.ProgramOption, 0, 1)
	if sim.stream.Reader == nil {
		programOptions = append(programOptions, tea.WithAltScreen())
	} else {
		programOptions = append(programOptions, tea.WithInputTTY())
	}

	program := tea.NewProgram(sim, programOptions...)
	if sim.stream.Reader != nil {
		go sim.readStream(program)
	}
	if _, err := program.Run(); err != nil {
		fmt.Printf("error running simulation: %v\n", err)
	}
}

// Init requests the first animation frame.
func (sim *Simulation) Init() tea.Cmd {
	if sim.stream.Reader != nil {
		return tea.Batch(sim.animate(), sim.streamSpawnTick())
	}

	return sim.animate()
}

// Update advances simulation state on frame ticks and handles quit keys.
func (sim *Simulation) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q", "esc":
			return sim, tea.Quit
		case " ":
			sim.spawnGlyph()
		}
	case stdinGlyphMsg:
		sim.pendingChars = append(sim.pendingChars, msg.char)
	case streamSpawnMsg:
		if sim.ready && len(sim.pendingChars) > 0 {
			streamChar := sim.pendingChars[0]
			switch streamChar {
			case '\r', '\n':
				sim.streamCursor = 0
			case '\t':
				sim.advanceStreamCursor(4)
			case ' ':
				sim.advanceStreamCursor(1)
			default:
				sim.spawnGlyphWithCharAtXAndDelay(streamChar, float64(sim.streamCursor), sim.stream.DropDelay)
				sim.advanceStreamCursor(1)
			}
			sim.pendingChars = sim.pendingChars[1:]
		}
		return sim, sim.streamSpawnTick()
	case tea.WindowSizeMsg:
		if msg.Width > 0 {
			sim.width = msg.Width
		}
		if msg.Height > 0 {
			sim.height = msg.Height
		}
		// Cache a full row of spaces once per resize so View can slice it cheaply.
		sim.spaces = strings.Repeat(" ", sim.width)
		if !sim.ready && sim.width > 0 && sim.height > 0 {
			// Delay initial placement until we know the terminal size.
			sim.placeGlyphs()
			sim.ready = true
		}
	case frameMsg:
		// Advance one simulation step, then immediately schedule the next one.
		sim.update()
		return sim, sim.animate()
	}

	return sim, nil
}

func (sim *Simulation) update() {
	frameSeconds := sim.frameDuration.Seconds()
	floor := float64(sim.height - 2) // last playfield row (height-1 is the footer)

	surviving := sim.glyphs[:0]
	for _, glyph := range sim.glyphs {
		if glyph.holdFor > 0 {
			glyph.holdFor -= frameSeconds
			if glyph.holdFor > 0 {
				surviving = append(surviving, glyph)
				continue
			}
			glyph.holdFor = 0
			glyph.yVel = 0
		}

		// Y axis: direct Euler integration — no spring lag, bounces land exactly on the floor row.
		glyph.yVel += sim.physics.Gravity
		glyph.y += glyph.yVel

		if glyph.y < 0 {
			glyph.y = 0
			glyph.yVel = math.Abs(glyph.yVel) // bounce off ceiling, ensure downward
		}
		if glyph.y >= floor {
			glyph.y = floor
			glyph.yVel = -glyph.yVel * sim.physics.Restitution
			glyph.targetXVel *= sim.physics.XFloorFriction // bleed off horizontal wander on bounce
		}

		// X axis: spring follower tracking a wandering target.
		glyph.targetX += glyph.targetXVel
		if glyph.targetX < 0 {
			glyph.targetX = 0
			glyph.targetXVel = -glyph.targetXVel
		} else if glyph.targetX >= float64(sim.width-1) {
			glyph.targetX = float64(sim.width - 1)
			glyph.targetXVel = -glyph.targetXVel
		}
		glyph.x, glyph.xVel = sim.spring.Update(glyph.x, glyph.xVel, glyph.targetX)

		if sim.physics.EnableDespawn {
			// Accumulate rest time; despawn after restTimeout seconds.
			if math.Abs(glyph.yVel) < sim.physics.RestThreshold && glyph.y >= floor-0.5 {
				glyph.stillFor += frameSeconds
			} else {
				glyph.stillFor = 0
			}
			if glyph.stillFor < sim.physics.RestTimeoutSeconds {
				surviving = append(surviving, glyph)
			}
		} else {
			glyph.stillFor = 0
			surviving = append(surviving, glyph)
		}
	}
	if sim.physics.EnableCollision && len(surviving) > 1 {
		sim.resolveCollisions(surviving, floor)
	}
	sim.glyphs = surviving
}

func (sim *Simulation) resolveCollisions(glyphs []*Glyph, floor float64) {
	const minDistance = 1.0
	const minDistanceSq = minDistance * minDistance
	const collisionRestitutionScale = 0.125
	const maxCollisionImpulse = 0.025
	const minApproachSpeed = -0.001
	const separationPercent = 0.5
	const separationSlop = 0.05

	type cell struct {
		x int
		y int
	}

	// Bucket glyphs into integer grid cells so each glyph only checks nearby
	// neighbors instead of every other glyph on screen.
	buckets := make(map[cell][]int, len(glyphs))
	for i, glyph := range glyphs {
		cellX := int(math.Floor(glyph.x))
		cellY := int(math.Floor(glyph.y))

		for dy := -1; dy <= 1; dy++ {
			for dx := -1; dx <= 1; dx++ {
				key := cell{x: cellX + dx, y: cellY + dy}
				for _, j := range buckets[key] {
					otherGlyph := glyphs[j]
					deltaX := glyph.x - otherGlyph.x
					deltaY := glyph.y - otherGlyph.y
					// Use squared distance first to avoid an unnecessary sqrt for most pairs.
					distSq := deltaX*deltaX + deltaY*deltaY
					if distSq <= 0 || distSq >= minDistanceSq {
						continue
					}

					dist := math.Sqrt(distSq)
					normalX := deltaX / dist
					normalY := deltaY / dist

					relXVel := glyph.xVel - otherGlyph.xVel
					relYVel := glyph.yVel - otherGlyph.yVel
					velAlongNormal := relXVel*normalX + relYVel*normalY
					// Only apply an impulse when the pair is moving toward each other.
					if velAlongNormal < minApproachSpeed {
						effectiveRestitution := math.Min(1, math.Max(0, sim.physics.Restitution*collisionRestitutionScale))
						impulse := -(1 + effectiveRestitution) * velAlongNormal / 2
						if impulse > maxCollisionImpulse {
							impulse = maxCollisionImpulse
						}
						impulseX := impulse * normalX
						impulseY := impulse * normalY
						glyph.xVel += impulseX
						glyph.yVel += impulseY
						otherGlyph.xVel -= impulseX
						otherGlyph.yVel -= impulseY
					}

					// Push overlapping glyphs apart even if their closing speed is tiny.
					separation := math.Max(minDistance-dist-separationSlop, 0) * separationPercent / 2
					glyph.x += normalX * separation
					glyph.y += normalY * separation
					otherGlyph.x -= normalX * separation
					otherGlyph.y -= normalY * separation

					// Re-clamp after separation so collision resolution cannot push glyphs
					// outside the playable area.
					if glyph.x < 0 {
						glyph.x = 0
					}
					if glyph.x > float64(sim.width-1) {
						glyph.x = float64(sim.width - 1)
					}
					if otherGlyph.x < 0 {
						otherGlyph.x = 0
					}
					if otherGlyph.x > float64(sim.width-1) {
						otherGlyph.x = float64(sim.width - 1)
					}

					// Clamp vertically too, since separation can push glyphs above the
					// ceiling or slightly below the resolved floor.
					if glyph.y < 0 {
						glyph.y = 0
					}
					if glyph.y > floor {
						glyph.y = floor
					}
					if otherGlyph.y < 0 {
						otherGlyph.y = 0
					}
					if otherGlyph.y > floor {
						otherGlyph.y = floor
					}
				}
			}
		}

		// Register this glyph after processing neighbors so each pair is resolved once.
		buckets[cell{x: cellX, y: cellY}] = append(buckets[cell{x: cellX, y: cellY}], i)
	}
}

// placed is a glyph position resolved to integer grid coordinates.
type placed struct {
	row, col int
	content  string
}

func (sim *Simulation) View() string {
	if sim.width < 1 || sim.height < 1 {
		return ""
	}

	// Collect glyph positions (exclude footer row so the hint is always visible).
	positions := make([]placed, 0, len(sim.glyphs))
	for _, glyph := range sim.glyphs {
		col := int(math.Round(glyph.x))
		row := int(math.Round(glyph.y))
		// The footer owns the last terminal row, so skip any glyph that would land there.
		if row >= 0 && row < sim.height-1 && col >= 0 && col < sim.width {
			positions = append(positions, placed{row, col, glyph.ansi})
		}
	}
	// Sort once so rendering can stream rows left-to-right without random access.
	sort.Slice(positions, func(i, j int) bool {
		if positions[i].row != positions[j].row {
			return positions[i].row < positions[j].row
		}
		return positions[i].col < positions[j].col
	})

	// out accumulates the full terminal frame before returning it to Bubble Tea.
	// Pre-growing it avoids repeated reallocations while rows are appended.
	var out strings.Builder
	out.Grow(sim.height * (sim.width + 1))

	positionIndex := 0
	for row := 0; row < sim.height; row++ {
		if row == sim.height-1 {
			// Footer row: fixed hint text, padded to full width.
			footer := "(Space: add glyph · q/esc/ctrl+c: quit)"
			if sim.stream.Reader != nil {
				footer = fmt.Sprintf("(stdin queued: %d · Space: add glyph · q/esc/ctrl+c: quit)", len(sim.pendingChars))
			}
			if len(footer) <= sim.width {
				out.WriteString(footer)
				out.WriteString(sim.spaces[:sim.width-len(footer)])
			} else {
				out.WriteString(footer[:sim.width])
			}
			break
		}

		// Write spaces and glyphs for this row using a left-to-right cursor.
		col := 0
		for positionIndex < len(positions) && positions[positionIndex].row == row {
			position := positions[positionIndex]
			positionIndex++
			if position.col < col {
				continue // two glyphs in the same cell; skip the later one
			}
			if position.col > col {
				out.WriteString(sim.spaces[:position.col-col])
			}
			out.WriteString(position.content)
			col = position.col + 1
		}
		// Pad remainder of row with spaces.
		if col < sim.width {
			out.WriteString(sim.spaces[:sim.width-col])
		}
		out.WriteByte('\n')
	}

	return out.String()
}
