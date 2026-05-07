package gamescope

type Option func(*Runner)

func WithGamescopeBin(path string) Option {
	return func(r *Runner) {
		if path != "" {
			r.GamescopeBin = path
		}
	}
}

func WithWineBin(path string) Option {
	return func(r *Runner) {
		if path != "" {
			r.WineBin = path
		}
	}
}

func WithWine(v bool) Option {
	return func(r *Runner) {
		r.UseWine = v
	}
}

func WithDefaultWinePrefix(prefix string) Option {
	return func(r *Runner) {
		r.DefaultWinePrefix = prefix
	}
}

func WithResolution(width, height int) Option {
	return func(r *Runner) {
		r.Width = width
		r.Height = height
	}
}

func WithOutputResolution(width, height int) Option {
	return func(r *Runner) {
		r.OutputWidth = width
		r.OutputHeight = height
	}
}

func WithRefreshRate(hz int) Option {
	return func(r *Runner) {
		r.RefreshRate = hz
	}
}

func WithFullscreen(v bool) Option {
	return func(r *Runner) {
		r.Fullscreen = v
	}
}

func WithBorderless(v bool) Option {
	return func(r *Runner) {
		r.Borderless = v
	}
}

func WithForceGrab(v bool) Option {
	return func(r *Runner) {
		r.ForceGrab = v
	}
}

func WithSteamDeckMode(v bool) Option {
	return func(r *Runner) {
		r.SteamDeckMode = v
	}
}

func WithExposeWayland(v bool) Option {
	return func(r *Runner) {
		r.ExposeWayland = v
	}
}

func WithExtraArgs(args ...string) Option {
	return func(r *Runner) {
		r.ExtraArgs = append(r.ExtraArgs, args...)
	}
}
