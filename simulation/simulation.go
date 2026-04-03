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

// Ball represents a ball in the simulation.
type Ball struct {
	x, y               float64 // displayed position
	vx, vy             float64 // spring velocity
	targetX, targetY   float64 // spring equilibrium point
	targetVX, targetVY float64 // target velocity
	char               rune    // display character
	ansi               string  // pre-computed colored cell string
}

// Simulation manages the bouncing balls.
type Simulation struct {
	width, height int
	ready         bool
	balls         []*Ball
	count         int
	rng           *rand.Rand
	spring        Spring
	frameDuration time.Duration
	spaces        string // s.width spaces, rebuilt on resize
}

type frameMsg time.Time

var ballChars = []rune{'●', '○', '◉', '◎', '◆'}
var ballColors = []lipgloss.Color{
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
	spring := NewSpring(deltaTime, 8.0, 0.35)
	balls := make([]*Ball, count)
	for i := range balls {
		c := ballColors[rng.Intn(len(ballColors))]
		ch := ballChars[i%len(ballChars)]
		balls[i] = &Ball{
			char: ch,
			ansi: lipgloss.NewStyle().Foreground(c).Render(string(ch)),
		}
	}
	return &Simulation{
		count:         count,
		rng:           rng,
		balls:         balls,
		spring:        spring,
		frameDuration: time.Second / time.Duration(fps),
	}
}

// placeBalls randomises ball positions within the current bounds.
func (s *Simulation) placeBalls() {
	for _, b := range s.balls {
		startX := s.rng.Float64() * float64(s.width-1)
		startY := s.rng.Float64() * float64(s.height-1)
		b.x = startX
		b.y = startY
		b.vx = 0
		b.vy = 0
		b.targetX = startX
		b.targetY = startY
		b.targetVX = (s.rng.Float64()*2 - 1) * 1.2
		b.targetVY = (s.rng.Float64()*2 - 1) * 0.8
	}
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
			s.placeBalls()
			s.ready = true
		}
	case frameMsg:
		s.update()
		return s, s.animate()
	}

	return s, nil
}

func (s *Simulation) update() {
	for _, b := range s.balls {
		b.targetX += b.targetVX
		b.targetY += b.targetVY

		if b.targetX < 0 {
			b.targetX = 0
			b.targetVX = -b.targetVX
		} else if b.targetX >= float64(s.width-1) {
			b.targetX = float64(s.width - 1)
			b.targetVX = -b.targetVX
		}

		if b.targetY < 0 {
			b.targetY = 0
			b.targetVY = -b.targetVY
		} else if b.targetY >= float64(s.height-1) {
			b.targetY = float64(s.height - 1)
			b.targetVY = -b.targetVY
		}

		b.x, b.vx = s.spring.Update(b.x, b.vx, b.targetX)
		b.y, b.vy = s.spring.Update(b.y, b.vy, b.targetY)
	}
}

// placed is a ball position resolved to integer grid coordinates.
type placed struct {
	row, col int
	content  string
}

func (s *Simulation) View() string {
	if s.width < 1 || s.height < 1 {
		return ""
	}

	// Collect ball positions (exclude footer row so the hint is always visible).
	positions := make([]placed, 0, len(s.balls))
	for _, b := range s.balls {
		col := int(math.Round(b.x))
		row := int(math.Round(b.y))
		if row >= 0 && row < s.height-1 && col >= 0 && col < s.width {
			positions = append(positions, placed{row, col, b.ansi})
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
			const footer = "(Press q, esc, or ctrl+c to quit)"
			if len(footer) <= s.width {
				out.WriteString(footer)
				out.WriteString(s.spaces[:s.width-len(footer)])
			} else {
				out.WriteString(footer[:s.width])
			}
			break
		}

		// Write spaces and balls for this row using a left-to-right cursor.
		col := 0
		for pi < len(positions) && positions[pi].row == row {
			p := positions[pi]
			pi++
			if p.col < col {
				continue // two balls in the same cell; skip the later one
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
