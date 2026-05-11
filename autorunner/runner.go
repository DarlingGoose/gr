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
	return newRunner(winePrefix, deps, nil, nil)
}
func NewRunnerWithOptions(winePrefix string, wineOpts []wine.Option, gamescopeOptions []gamescope.Option) (gr.Runner, error) {
	deps := CheckDependencies(true)
	return newRunner(winePrefix, deps, wineOpts, gamescopeOptions)
}

func newRunner(winePrefix string, deps DependencyStatus, wineOpts []wine.Option, gamescopeOptions []gamescope.Option) (gr.Runner, error) {
	if !deps.WineInstalled() {
		return nil, errors.New("wine is not installed")
	}

	switch {
	case deps.GamescopeInstalled():
		var outW, outH int
		// A sensible default for the game's internal rendering resolution
		var inW, inH = 1920, 1080

		m, err := monitors.GetMonitors()
		if err == nil && len(m) > 0 {
			outW = m[0].CurrentMode.Width
			outH = m[0].CurrentMode.Height
		} else {
			// Fallbacks if monitor detection fails
			outW = 1280
			outH = 720
			inW = 1280
			inH = 720
		}
		defaultOptions := []gamescope.Option{
			gamescope.WithWine(true),
			gamescope.WithDefaultWinePrefix(winePrefix),

			// 1. Separate render resolution from output resolution
			gamescope.WithResolution(inW, inH),
			gamescope.WithOutputResolution(outW, outH),

			// 2. Default to Fullscreen for better Wayland compatibility
			gamescope.WithFullscreen(true),

			// 3. Use AMD FSR for high-quality upscaling
			gamescope.WithScaler("fsr"),

			// 4. Expose Wayland if applicable
			gamescope.WithExposeWayland(monitors.IsWayland()),
		}
		if len(gamescopeOptions) > 0 {
			defaultOptions = append(defaultOptions, gamescopeOptions...)
		}

		return gamescope.New(
			defaultOptions...,
		), nil
	default:
		defaultOptions := []wine.Option{
			wine.WithDefaultPrefix(winePrefix),
		}
		if len(wineOpts) > 0 {
			defaultOptions = append(defaultOptions, wineOpts...)
		}
		return wine.New(defaultOptions...), nil
	}
}
