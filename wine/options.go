package wine

type Option func(*Options)

type Options struct {
	Name          string
	WineBin       string
	WineTricksBin string
	DefaultPrefix string
}

func ApplyOptions(opts ...Option) Options {
	o := Options{
		Name:          "wine",
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

func WithName(name string) Option {
	return func(r *Options) {
		if name != "" {
			r.Name = name
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
