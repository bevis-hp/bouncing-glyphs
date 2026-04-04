package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/bevis-hp/glyphfall/simulation"
)

func stdinHasStream() bool {
	stdinInfo, err := os.Stdin.Stat()
	if err != nil {
		return false
	}

	return stdinInfo.Mode()&os.ModeCharDevice == 0
}

func main() {
	physics := simulation.PhysicsConfig{
		Gravity:            0.008,
		Restitution:        0.6,
		XFloorFriction:     0.96,
		RestThreshold:      0.08,
		RestTimeoutSeconds: 5.0,
		SpringFrequency:    5.0,
		SpringDampingRatio: 0.5,
		LaunchKickMax:      0.6,
		SpawnKickMax:       1.0,
		TargetDriftMax:     0.7,
	}

	count := flag.Int("count", 10, "number of glyphs to simulate")
	fps := flag.Int("fps", 60, "frames per second")
	stdinIntervalMS := flag.Int("stdin-interval-ms", 100, "milliseconds between glyph spawns from piped stdin")
	stdinDropDelayMS := flag.Int("stdin-drop-delay-ms", 200, "milliseconds stdin glyphs wait at top before dropping")
	flag.Float64Var(&physics.Gravity, "gravity", physics.Gravity, "downward acceleration in cells/frame^2")
	flag.Float64Var(&physics.Restitution, "restitution", physics.Restitution, "bounce speed retention fraction")
	flag.Float64Var(&physics.XFloorFriction, "x-floor-friction", physics.XFloorFriction, "horizontal drift retention on floor bounce")
	flag.Float64Var(&physics.RestThreshold, "rest-threshold", physics.RestThreshold, "speed below which glyphs are considered resting")
	flag.Float64Var(&physics.RestTimeoutSeconds, "rest-timeout", physics.RestTimeoutSeconds, "seconds at rest before glyph despawns")
	flag.Float64Var(&physics.SpringFrequency, "spring-frequency", physics.SpringFrequency, "x-axis spring angular frequency")
	flag.Float64Var(&physics.SpringDampingRatio, "spring-damping", physics.SpringDampingRatio, "x-axis spring damping ratio")
	flag.Float64Var(&physics.LaunchKickMax, "launch-kick-max", physics.LaunchKickMax, "max upward launch speed for initial glyphs")
	flag.Float64Var(&physics.SpawnKickMax, "spawn-kick-max", physics.SpawnKickMax, "max upward launch speed for spawned glyphs")
	flag.Float64Var(&physics.TargetDriftMax, "target-drift-max", physics.TargetDriftMax, "max x-target drift speed magnitude")
	flag.BoolVar(&physics.EnableDespawn, "despawn", false, "despawn resting glyphs after rest-timeout")
	flag.BoolVar(&physics.EnableCollision, "collision", false, "enable glyph-glyph collisions (higher CPU cost)")
	flag.Parse()

	hasStreamInput := stdinHasStream()
	isCountSet := false
	flag.Visit(func(f *flag.Flag) {
		if f.Name == "count" {
			isCountSet = true
		}
	})
	if hasStreamInput && !isCountSet {
		*count = 0
	}

	if *count < 0 {
		fmt.Fprintln(os.Stderr, "error: count must be at least 0")
		os.Exit(1)
	}
	if *fps < 1 {
		fmt.Fprintln(os.Stderr, "error: fps must be at least 1")
		os.Exit(1)
	}
	if *stdinIntervalMS < 1 {
		fmt.Fprintln(os.Stderr, "error: stdin-interval-ms must be at least 1")
		os.Exit(1)
	}
	if *stdinDropDelayMS < 0 {
		fmt.Fprintln(os.Stderr, "error: stdin-drop-delay-ms must be >= 0")
		os.Exit(1)
	}
	if physics.Gravity < 0 {
		fmt.Fprintln(os.Stderr, "error: gravity must be >= 0")
		os.Exit(1)
	}
	if physics.Restitution < 0 {
		fmt.Fprintln(os.Stderr, "error: restitution must be >= 0")
		os.Exit(1)
	}
	if physics.XFloorFriction < 0 {
		fmt.Fprintln(os.Stderr, "error: x-floor-friction must be >= 0")
		os.Exit(1)
	}
	if physics.RestThreshold < 0 {
		fmt.Fprintln(os.Stderr, "error: rest-threshold must be >= 0")
		os.Exit(1)
	}
	if physics.RestTimeoutSeconds < 0 {
		fmt.Fprintln(os.Stderr, "error: rest-timeout must be >= 0")
		os.Exit(1)
	}
	if physics.SpringFrequency < 0 {
		fmt.Fprintln(os.Stderr, "error: spring-frequency must be >= 0")
		os.Exit(1)
	}
	if physics.SpringDampingRatio < 0 {
		fmt.Fprintln(os.Stderr, "error: spring-damping must be >= 0")
		os.Exit(1)
	}
	if physics.LaunchKickMax < 0 {
		fmt.Fprintln(os.Stderr, "error: launch-kick-max must be >= 0")
		os.Exit(1)
	}
	if physics.SpawnKickMax < 0 {
		fmt.Fprintln(os.Stderr, "error: spawn-kick-max must be >= 0")
		os.Exit(1)
	}
	if physics.TargetDriftMax < 0 {
		fmt.Fprintln(os.Stderr, "error: target-drift-max must be >= 0")
		os.Exit(1)
	}

	sim := simulation.NewWithPhysicsConfig(*count, *fps, physics)
	if hasStreamInput {
		sim.EnableStream(
			os.Stdin,
			time.Duration(*stdinIntervalMS)*time.Millisecond,
			time.Duration(*stdinDropDelayMS)*time.Millisecond,
		)
	}
	sim.Run()
}
