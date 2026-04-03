package simulation

import (
	"fmt"
	"math"
	"math/rand"
	"sort"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Physics constants for the energy-loss/gravity system.
const (
	gravity        = 0.008 // cells/frame² downward acceleration (tuned for default 60 fps)
	restitution    = 0.75  // fraction of speed retained on floor bounce
	xFloorFriction = 0.96  // fraction of horizontal speed retained on floor bounce
	restThreshold  = 0.08  // cells/frame speed below which a glyph is considered at rest
	restTimeout    = 5.0   // seconds at rest before despawn
)

// Glyph represents a glyph in the simulation.
// The X axis uses a spring follower for a drifting, slightly-lagged feel.
// The Y axis uses direct Euler integration so gravity and floor bounces are
// lag-free and land exactly on the floor row.
type Glyph struct {
	x, y     float64 // displayed position
	vx       float64 // spring internal velocity (x axis)
	vy       float64 // physics velocity (y axis, cells/frame)
	targetX  float64 // spring equilibrium point (x only)
	targetVX float64 // wandering velocity of the x target
	stillFor float64 // seconds spent at rest (for despawn)
	char     rune    // display character
	ansi     string  // pre-computed colored cell string
}

// Simulation manages the bouncing glyphs.
type Simulation struct {
	width, height int
	ready         bool
	glyphs        []*Glyph
	count         int
	rng           *rand.Rand
	spring        Spring
	frameDuration time.Duration
	spaces        string // s.width spaces, rebuilt on resize
}

type frameMsg time.Time

var glyphChars = []rune("ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789!@#$%^&*()-_=+[]{}|;:'\",.<>/?~£€¢¥§±×÷¶©®™✓✕★☆◆◇●○■□▲△✦✧✩✪✫✬✭✮✯✰✶✷")
var glyphColors = []lipgloss.Color{
	"#DC3C3C",
	"#3CB44B",
	"#FFE119",
	"#0082C8",
	"#F58330",
	"#911EB4",
	"#46F0F0",
	"#F032E6",
	"#D2F53C",
	"#FABFD4",
	"#AA6E28",
	"#008080",
}

// New creates a new Simulation. Width and height are determined from the
// terminal on the first WindowSizeMsg.
func New(count, fps int) *Simulation {
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	if fps < 1 {
		fps = 1
	}
	deltaTime := FPS(fps)
	spring := NewSpring(deltaTime, 5.0, 0.55)
	glyphs := make([]*Glyph, count)
	for i := range glyphs {
		c := glyphColors[rng.Intn(len(glyphColors))]
		ch := glyphChars[rng.Intn(len(glyphChars))]
		glyphs[i] = &Glyph{
			char: ch,
			ansi: lipgloss.NewStyle().Foreground(c).Render(string(ch)),
		}
	}
	return &Simulation{
		count:         count,
		rng:           rng,
		glyphs:        glyphs,
		spring:        spring,
		frameDuration: time.Second / time.Duration(fps),
	}
}

// placeGlyphs launches all glyphs from the top of the screen.
func (s *Simulation) placeGlyphs() {
	for _, g := range s.glyphs {
		startX := s.rng.Float64() * float64(s.width-1)
		g.x = startX
		g.y = 0
		g.vx = 0
		g.vy = -(s.rng.Float64() * 0.6) // small upward kick; gravity brings them down
		g.targetX = startX
		g.targetVX = (s.rng.Float64()*2 - 1) * 0.7
		g.stillFor = 0
	}
}

// spawnGlyph adds a new glyph at a random x position at the top of the screen.
func (s *Simulation) spawnGlyph() {
	if s.width < 1 {
		return
	}
	c := glyphColors[s.rng.Intn(len(glyphColors))]
	ch := glyphChars[s.rng.Intn(len(glyphChars))]
	x := s.rng.Float64() * float64(s.width-1)
	g := &Glyph{
		x:        x,
		y:        0,
		vy:       -(s.rng.Float64() * 1.0), // upward kick; gravity brings it down
		targetX:  x,
		targetVX: (s.rng.Float64()*2 - 1) * 0.7,
		char:     ch,
		ansi:     lipgloss.NewStyle().Foreground(c).Render(string(ch)),
	}
	s.glyphs = append(s.glyphs, g)
}

func (s *Simulation) animate() tea.Cmd {
	return tea.Tick(s.frameDuration, func(t time.Time) tea.Msg {
		return frameMsg(t)
	})
}

// Run starts the Bubble Tea simulation.
func (s *Simulation) Run() {
	p := tea.NewProgram(s, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Printf("error running simulation: %v\n", err)
	}
}

// Init requests the first animation frame.
func (s *Simulation) Init() tea.Cmd {
	return s.animate()
}

// Update advances simulation state on frame ticks and handles quit keys.
func (s *Simulation) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q", "esc":
			return s, tea.Quit
		case " ":
			s.spawnGlyph()
		}
	case tea.WindowSizeMsg:
		if msg.Width > 0 {
			s.width = msg.Width
		}
		if msg.Height > 0 {
			s.height = msg.Height
		}
		s.spaces = strings.Repeat(" ", s.width)
		if !s.ready && s.width > 0 && s.height > 0 {
			s.placeGlyphs()
			s.ready = true
		}
	case frameMsg:
		s.update()
		return s, s.animate()
	}

	return s, nil
}

func (s *Simulation) update() {
	dt := s.frameDuration.Seconds()
	floor := float64(s.height - 2) // last playfield row (height-1 is the footer)

	surviving := s.glyphs[:0]
	for _, g := range s.glyphs {
		// Y axis: direct Euler integration — no spring lag, bounces land exactly
		// on the floor row.
		g.vy += gravity
		g.y += g.vy

		if g.y < 0 {
			g.y = 0
			g.vy = math.Abs(g.vy) // bounce off ceiling, ensure downward
		}
		if g.y >= floor {
			g.y = floor
			g.vy = -g.vy * restitution
			g.targetVX *= xFloorFriction // bleed off horizontal wander on bounce
		}

		// X axis: spring follower tracking a wandering target.
		g.targetX += g.targetVX
		if g.targetX < 0 {
			g.targetX = 0
			g.targetVX = -g.targetVX
		} else if g.targetX >= float64(s.width-1) {
			g.targetX = float64(s.width - 1)
			g.targetVX = -g.targetVX
		}
		g.x, g.vx = s.spring.Update(g.x, g.vx, g.targetX)

		// Accumulate rest time; despawn after restTimeout seconds.
		if math.Abs(g.vy) < restThreshold && g.y >= floor-0.5 {
			g.stillFor += dt
		} else {
			g.stillFor = 0
		}
		if g.stillFor < restTimeout {
			surviving = append(surviving, g)
		}
	}
	s.glyphs = surviving
}

// placed is a glyph position resolved to integer grid coordinates.
type placed struct {
	row, col int
	content  string
}

func (s *Simulation) View() string {
	if s.width < 1 || s.height < 1 {
		return ""
	}

	// Collect glyph positions (exclude footer row so the hint is always visible).
	positions := make([]placed, 0, len(s.glyphs))
	for _, g := range s.glyphs {
		col := int(math.Round(g.x))
		row := int(math.Round(g.y))
		if row >= 0 && row < s.height-1 && col >= 0 && col < s.width {
			positions = append(positions, placed{row, col, g.ansi})
		}
	}
	sort.Slice(positions, func(i, j int) bool {
		if positions[i].row != positions[j].row {
			return positions[i].row < positions[j].row
		}
		return positions[i].col < positions[j].col
	})

	var out strings.Builder
	out.Grow(s.height * (s.width + 1))

	pi := 0
	for row := 0; row < s.height; row++ {
		if row == s.height-1 {
			// Footer row: fixed hint text, padded to full width.
			const footer = "(Space: add glyph · q/esc/ctrl+c: quit)"
			if len(footer) <= s.width {
				out.WriteString(footer)
				out.WriteString(s.spaces[:s.width-len(footer)])
			} else {
				out.WriteString(footer[:s.width])
			}
			break
		}

		// Write spaces and glyphs for this row using a left-to-right cursor.
		col := 0
		for pi < len(positions) && positions[pi].row == row {
			p := positions[pi]
			pi++
			if p.col < col {
				continue // two glyphs in the same cell; skip the later one
			}
			if p.col > col {
				out.WriteString(s.spaces[:p.col-col])
			}
			out.WriteString(p.content)
			col = p.col + 1
		}
		// Pad remainder of row with spaces.
		if col < s.width {
			out.WriteString(s.spaces[:s.width-col])
		}
		out.WriteByte('\n')
	}

	return out.String()
}
