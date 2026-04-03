package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/bevis-hp/bouncing-glyphs/simulation"
)

func main() {
	count := flag.Int("count", 10, "number of glyphs to simulate")
	fps := flag.Int("fps", 60, "frames per second")
	gravity := flag.Float64("gravity", 0.008, "downward acceleration in cells/frame^2")
	restitution := flag.Float64("restitution", 0.75, "bounce speed retention fraction")
	xFloorFriction := flag.Float64("x-floor-friction", 0.96, "horizontal drift retention on floor bounce")
	restThreshold := flag.Float64("rest-threshold", 0.08, "speed below which glyphs are considered resting")
	restTimeout := flag.Float64("rest-timeout", 5.0, "seconds at rest before glyph despawns")
	springFrequency := flag.Float64("spring-frequency", 5.0, "x-axis spring angular frequency")
	springDamping := flag.Float64("spring-damping", 0.55, "x-axis spring damping ratio")
	launchKickMax := flag.Float64("launch-kick-max", 0.6, "max upward launch speed for initial glyphs")
	spawnKickMax := flag.Float64("spawn-kick-max", 1.0, "max upward launch speed for spawned glyphs")
	targetDriftMax := flag.Float64("target-drift-max", 0.7, "max x-target drift speed magnitude")
	flag.Parse()

	if *count < 1 {
		fmt.Fprintln(os.Stderr, "error: count must be at least 1")
		os.Exit(1)
	}
	if *fps < 1 {
		fmt.Fprintln(os.Stderr, "error: fps must be at least 1")
		os.Exit(1)
	}
	if *gravity < 0 {
		fmt.Fprintln(os.Stderr, "error: gravity must be >= 0")
		os.Exit(1)
	}
	if *restitution < 0 {
		fmt.Fprintln(os.Stderr, "error: restitution must be >= 0")
		os.Exit(1)
	}
	if *xFloorFriction < 0 {
		fmt.Fprintln(os.Stderr, "error: x-floor-friction must be >= 0")
		os.Exit(1)
	}
	if *restThreshold < 0 {
		fmt.Fprintln(os.Stderr, "error: rest-threshold must be >= 0")
		os.Exit(1)
	}
	if *restTimeout < 0 {
		fmt.Fprintln(os.Stderr, "error: rest-timeout must be >= 0")
		os.Exit(1)
	}
	if *springFrequency < 0 {
		fmt.Fprintln(os.Stderr, "error: spring-frequency must be >= 0")
		os.Exit(1)
	}
	if *springDamping < 0 {
		fmt.Fprintln(os.Stderr, "error: spring-damping must be >= 0")
		os.Exit(1)
	}
	if *launchKickMax < 0 {
		fmt.Fprintln(os.Stderr, "error: launch-kick-max must be >= 0")
		os.Exit(1)
	}
	if *spawnKickMax < 0 {
		fmt.Fprintln(os.Stderr, "error: spawn-kick-max must be >= 0")
		os.Exit(1)
	}
	if *targetDriftMax < 0 {
		fmt.Fprintln(os.Stderr, "error: target-drift-max must be >= 0")
		os.Exit(1)
	}

	physics := simulation.PhysicsConfig{
		Gravity:            *gravity,
		Restitution:        *restitution,
		XFloorFriction:     *xFloorFriction,
		RestThreshold:      *restThreshold,
		RestTimeoutSeconds: *restTimeout,
		SpringFrequency:    *springFrequency,
		SpringDampingRatio: *springDamping,
		LaunchKickMax:      *launchKickMax,
		SpawnKickMax:       *spawnKickMax,
		TargetDriftMax:     *targetDriftMax,
	}

	sim := simulation.NewWithPhysicsConfig(*count, *fps, physics)
	sim.Run()
}
