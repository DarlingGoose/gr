package gr

type Config struct {
	Background   bool     `json:"background,omitempty"`
	WorkingDir   string   `json:"working_dir,omitempty"`
	Args         []string `json:"args,omitempty"`
	Envs         []string `json:"envs,omitempty"`
	SystemArch   string   `json:"system_arch,omitempty"`
	WinePrefix   string   `json:"wine_prefix,omitempty"`
	Dependencies []string `json:"dependencies,omitempty"`
	Name         string   `json:"name,omitempty"`
	PID          int      `json:"pid,omitempty"`
	Session      string   `json:"session,omitempty"`
	SessionID    string   `json:"session_id,omitempty"`
	LogFile      string   `json:"logFile,omitempty"`
}

func NewConfig(opts ...Option) Config {
	return ApplyOptions(opts...).Config()
}

func (o Options) Config() Config {
	return Config{
		Background:   o.background,
		WorkingDir:   o.workingDir,
		Args:         append([]string(nil), o.args...),
		Envs:         append([]string(nil), o.envs...),
		SystemArch:   o.systemArch,
		WinePrefix:   o.wineprefix,
		Dependencies: append([]string(nil), o.dependencies...),
		Name:         o.name,
		PID:          o.pid,
		Session:      o.session,
		SessionID:    o.sessionID,
		LogFile:      o.logFile,
	}
}

func (c Config) Options() []Option {
	opts := make([]Option, 0, 10)

	if c.Background {
		opts = append(opts, WithBackground(c.Background))
	}
	if c.WorkingDir != "" {
		opts = append(opts, WithWorkingDir(c.WorkingDir))
	}
	if len(c.Args) > 0 {
		opts = append(opts, WithArgs(c.Args...))
	}
	if len(c.Envs) > 0 {
		opts = append(opts, WithEnv(c.Envs...))
	}
	if c.WinePrefix != "" {
		opts = append(opts, WithWinePrefix(c.WinePrefix))
	}
	if c.SystemArch != "" {
		opts = append(opts, WithSystemArch(c.SystemArch))
	}
	if len(c.Dependencies) > 0 {
		opts = append(opts, WithDependencies(c.Dependencies...))
	}
	if c.Name != "" {
		opts = append(opts, WithName(c.Name))
	}
	if c.PID != 0 {
		opts = append(opts, WithPID(c.PID))
	}
	if c.Session != "" {
		opts = append(opts, WithSession(c.Session))
	}
	if c.SessionID != "" {
		opts = append(opts, WithSessionID(c.SessionID))
	}
	if c.LogFile != "" {
		opts = append(opts, WithLogFile(c.LogFile))
	}

	return opts
}
