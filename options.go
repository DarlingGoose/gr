package gr

type Option func(*Options)

func ApplyOptions(opts ...Option) Options {
	var o Options
	for _, opt := range opts {
		if opt != nil {
			opt(&o)
		}
	}
	return o
}

func WithBackground(v bool) Option {
	return func(o *Options) {
		o.background = v
	}
}

func WithArgs(args ...string) Option {
	return func(o *Options) {
		o.args = append(o.args, args...)
	}
}

func WithEnv(envs ...string) Option {
	return func(o *Options) {
		o.envs = append(o.envs, envs...)
	}
}

func WithWinePrefix(prefix string) Option {
	return func(o *Options) {
		o.wineprefix = prefix
	}
}

func WithSystemArch(arch string) Option {
	return func(o *Options) {
		o.systemArch = arch
	}
}

func WithDependencies(deps ...string) Option {
	return func(o *Options) {
		o.dependencies = append(o.dependencies, deps...)
	}
}

func WithName(name string) Option {
	return func(o *Options) {
		o.name = name
	}
}

func WithPID(pid int) Option {
	return func(o *Options) {
		o.pid = pid
	}
}

func WithSession(session string) Option {
	return func(o *Options) {
		o.session = session
	}
}

func WithSessionID(sessionID string) Option {
	return func(o *Options) {
		o.sessionID = sessionID
	}
}
