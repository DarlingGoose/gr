package wine

type Option func(*Runner)

func WithWineBin(path string) Option {
	return func(r *Runner) {
		if path != "" {
			r.WineBin = path
		}
	}
}

func WithWineTricksBin(path string) Option {
	return func(r *Runner) {
		if path != "" {
			r.WineTricksBin = path
		}
	}
}

func WithDefaultPrefix(prefix string) Option {
	return func(r *Runner) {
		r.DefaultPrefix = prefix
	}
}
