package wine

type Option func(*Options)

type Options struct {
	WineBin       string
	WineTricksBin string
	DefaultPrefix string
}

func ApplyOptions(opts ...Option) Options {
	o := Options{
		WineBin:       "wine",
		WineTricksBin: "winetricks",
	}

	for _, opt := range opts {
		if opt != nil {
			opt(&o)
		}
	}

	return o
}

func WithWineBin(path string) Option {
	return func(r *Options) {
		if path != "" {
			r.WineBin = path
		}
	}
}

func WithWineTricksBin(path string) Option {
	return func(r *Options) {
		if path != "" {
			r.WineTricksBin = path
		}
	}
}

func WithDefaultPrefix(prefix string) Option {
	return func(r *Options) {
		r.DefaultPrefix = prefix
	}
}
