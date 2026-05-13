package autorunner

import (
	"errors"

	"github.com/DarlingGoose/gr"
	"github.com/DarlingGoose/gr/gamescope"
	"github.com/DarlingGoose/gr/monitors"
	"github.com/DarlingGoose/gr/wine"
)

type RunnerConfig struct {
	UseGamescope bool               `json:"use_gamescope"`
	Wine         *wine.Options      `json:"wine,omitempty"`
	Gamescope    *gamescope.Options `json:"gamescope,omitempty"`
}

func NewRunner(winePrefix string) (gr.Runner, error) {
	deps := CheckDependencies(true)
	return newRunner(winePrefix, deps, nil, nil)
}

func NewRunnerWithOptions(winePrefix string, wineOpts []wine.Option, gamescopeOptions []gamescope.Option) (gr.Runner, error) {
	deps := CheckDependencies(true)
	return newRunner(winePrefix, deps, wineOpts, gamescopeOptions)
}

func DefaultRunnerConfig(winePrefix string) (RunnerConfig, error) {
	deps := CheckDependencies(true)
	return defaultRunnerConfig(winePrefix, deps, nil, nil)
}

func DefaultRunnerConfigWithOptions(winePrefix string, wineOpts []wine.Option, gamescopeOptions []gamescope.Option) (RunnerConfig, error) {
	deps := CheckDependencies(true)
	return defaultRunnerConfig(winePrefix, deps, wineOpts, gamescopeOptions)
}

func RunnerConfigFor(r gr.Runner) (RunnerConfig, bool) {
	switch r := r.(type) {
	case *gamescope.Runner:
		o := r.GetOptions()
		return RunnerConfig{
			UseGamescope: true,
			Gamescope:    &o,
		}, true
	case *wine.Runner:
		o := r.GetOptions()
		return RunnerConfig{
			Wine: &o,
		}, true
	default:
		return RunnerConfig{}, false
	}
}

func newRunner(winePrefix string, deps DependencyStatus, wineOpts []wine.Option, gamescopeOptions []gamescope.Option) (gr.Runner, error) {
	cfg, err := defaultRunnerConfig(winePrefix, deps, wineOpts, gamescopeOptions)
	if err != nil {
		return nil, err
	}

	if cfg.Gamescope != nil {
		return gamescope.NewFromOptions(*cfg.Gamescope), nil
	}

	return wine.NewFromOptions(*cfg.Wine), nil
}

func defaultRunnerConfig(winePrefix string, deps DependencyStatus, wineOpts []wine.Option, gamescopeOptions []gamescope.Option) (RunnerConfig, error) {
	if !deps.WineInstalled() {
		return RunnerConfig{}, errors.New("wine is not installed")
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
			gamescope.WithScaler("fit"),
			gamescope.WithFilter("linear"),
			// 4. Expose Wayland if applicable
			gamescope.WithExposeWayland(monitors.IsWayland()),
		}
		if len(gamescopeOptions) > 0 {
			defaultOptions = append(defaultOptions, gamescopeOptions...)
		}

		o := gamescope.ApplyOptions(defaultOptions...)
		return RunnerConfig{
			UseGamescope: true,
			Gamescope:    &o,
		}, nil
	default:
		defaultOptions := []wine.Option{
			wine.WithDefaultPrefix(winePrefix),
		}
		if len(wineOpts) > 0 {
			defaultOptions = append(defaultOptions, wineOpts...)
		}
		o := wine.ApplyOptions(defaultOptions...)
		return RunnerConfig{
			Wine: &o,
		}, nil
	}
}
