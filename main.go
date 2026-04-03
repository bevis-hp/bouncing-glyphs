package main

import (
	"bouncing-glyphs/simulation"
	"flag"
	"fmt"
	"os"
)

func main() {
	count := flag.Int("count", 10, "number of glyphs to simulate")
	fps := flag.Int("fps", 60, "frames per second")
	flag.Parse()

	if *count < 1 {
		fmt.Fprintln(os.Stderr, "error: count must be at least 1")
		os.Exit(1)
	}

	sim := simulation.New(*count, *fps)
	sim.Run()
}
