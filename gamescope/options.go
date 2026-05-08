package gamescope

type Option func(*Options)

type Options struct {
	Name         string
	GamescopeBin string
	WineBin      string

	DefaultWinePrefix string

	UseWine bool

	Width        int
	Height       int
	RefreshRate  int
	OutputWidth  int
	OutputHeight int

	Fullscreen bool
	Borderless bool
	ForceGrab  bool

	SteamDeckMode bool
	ExposeWayland bool

	Scaler string
	Filter string

	ExtraArgs []string
}

func ApplyOptions(opts ...Option) Options {
	o := Options{
		Name:         "gamescope",
		GamescopeBin: "gamescope",
		WineBin:      "wine",
	}

	for _, opt := range opts {
		if opt != nil {
			opt(&o)
		}
	}

	o.ExtraArgs = append([]string(nil), o.ExtraArgs...)

	return o
}

func WithName(name string) Option {
	return func(r *Options) {
		if name != "" {
			r.Name = name
		}
	}
}

func WithGamescopeBin(path string) Option {
	return func(r *Options) {
		if path != "" {
			r.GamescopeBin = path
		}
	}
}

func WithWineBin(path string) Option {
	return func(r *Options) {
		if path != "" {
			r.WineBin = path
		}
	}
}

func WithWine(v bool) Option {
	return func(r *Options) {
		r.UseWine = v
	}
}

func WithDefaultWinePrefix(prefix string) Option {
	return func(r *Options) {
		r.DefaultWinePrefix = prefix
	}
}

func WithResolution(width, height int) Option {
	return func(r *Options) {
		r.Width = width
		r.Height = height
	}
}

func WithOutputResolution(width, height int) Option {
	return func(r *Options) {
		r.OutputWidth = width
		r.OutputHeight = height
	}
}

func WithRefreshRate(hz int) Option {
	return func(r *Options) {
		r.RefreshRate = hz
	}
}

func WithFullscreen(v bool) Option {
	return func(r *Options) {
		r.Fullscreen = v
	}
}

func WithBorderless(v bool) Option {
	return func(r *Options) {
		r.Borderless = v
	}
}

func WithForceGrab(v bool) Option {
	return func(r *Options) {
		r.ForceGrab = v
	}
}

func WithSteamDeckMode(v bool) Option {
	return func(r *Options) {
		r.SteamDeckMode = v
	}
}

func WithExposeWayland(v bool) Option {
	return func(r *Options) {
		r.ExposeWayland = v
	}
}

func WithScaler(scaler string) Option {
	return func(r *Options) {
		r.Scaler = scaler
	}
}

func WithFilter(filter string) Option {
	return func(r *Options) {
		r.Filter = filter
	}
}

func WithExtraArgs(args ...string) Option {
	return func(r *Options) {
		r.ExtraArgs = append(r.ExtraArgs, args...)
	}
}
