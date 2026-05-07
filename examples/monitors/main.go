package main

import (
	"fmt"
	"log"

	"github.com/DarlingGoose/gr/monitors"
)

func main() {
	monitors, err := monitors.GetMonitors()
	if err != nil {
		log.Fatal(err)
	}

	for _, mon := range monitors {
		fmt.Printf("%s connected=%v current=%dx%d\n",
			mon.Name,
			mon.Connected,
			mon.CurrentMode.Width,
			mon.CurrentMode.Height,
		)

		for _, mode := range mon.Modes {
			fmt.Printf("  width=%d height=%d\n", mode.Width, mode.Height)
		}
	}
}
