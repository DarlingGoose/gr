package gr

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

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

type GameConfig struct {
	ExePath string `json:"exe_path,omitempty"`
	Config  Config `json:"config,omitempty"`
}

func NewConfig(opts ...Option) Config {
	return ApplyOptions(opts...).Config()
}

func NewGameConfig(exePath string, opts ...Option) GameConfig {
	return GameConfig{
		ExePath: exePath,
		Config:  NewConfig(opts...),
	}
}

func LoadConfig(path string) (Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Config{}, fmt.Errorf("read config: %w", err)
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return Config{}, fmt.Errorf("decode config: %w", err)
	}

	return cfg, nil
}

func LoadGameConfig(path string) (GameConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return GameConfig{}, fmt.Errorf("read game config: %w", err)
	}

	var cfg GameConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		return GameConfig{}, fmt.Errorf("decode game config: %w", err)
	}

	return cfg, nil
}

func DeleteConfig(path string) error {
	if err := os.Remove(path); err != nil && !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("delete config: %w", err)
	}

	return nil
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

func (c Config) Save(path string) error {
	return writeJSON(path, c, "config")
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

func (c GameConfig) Save(path string) error {
	return writeJSON(path, c, "game config")
}

func (c GameConfig) Options() []Option {
	return c.Config.Options()
}

func writeJSON(path string, v interface{}, name string) error {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return fmt.Errorf("encode %s: %w", name, err)
	}

	if dir := filepath.Dir(path); dir != "." && dir != "" {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return fmt.Errorf("create %s directory: %w", name, err)
		}
	}

	if err := os.WriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("write %s: %w", name, err)
	}

	return nil
}
