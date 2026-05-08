package autorunner

import (
	"errors"

	"github.com/DarlingGoose/gr"
	"github.com/DarlingGoose/gr/gamescope"
	"github.com/DarlingGoose/gr/monitors"
	"github.com/DarlingGoose/gr/wine"
)

func NewRunner(winePrefix string) (gr.Runner, error) {
	deps := CheckDependencies(true)
	return newRunner(winePrefix, deps)
}

func newRunner(winePrefix string, deps DependencyStatus) (gr.Runner, error) {
	if !deps.WineInstalled() {
		return nil, errors.New("wine is not installed")
	}

	switch {
	case deps.GamescopeInstalled():
		var w, h int
		m, err := monitors.GetMonitors()
		if err == nil && len(m) > 0 {
			w = m[0].CurrentMode.Width
			h = m[0].CurrentMode.Height
		} else {
			w = 1280
			h = 720
		}

		return gamescope.New(
			gamescope.WithWine(true),
			gamescope.WithDefaultWinePrefix(winePrefix),
			gamescope.WithResolution(w, h),
			gamescope.WithFullscreen(false),
			gamescope.WithExposeWayland(monitors.IsWayland()),
		), nil
	default:
		return wine.New(wine.WithDefaultPrefix(winePrefix)), nil
	}
}
