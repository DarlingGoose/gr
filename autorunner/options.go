package autorunner

import (
	"fmt"
	"path/filepath"

	"github.com/DarlingGoose/gr"
)

type DefaultOptionsConfig struct {
	WinePrefix          string
	WorkingDir          string
	UseGamescope        bool
	RequireWinetricks   bool
	ForceWineArch       bool
	WineBin             string
	GamescopeBin        string
	WinetricksBin       string
	Dependencies        []string
	Args                []string
	Env                 WineEnvConfig
	SkipDependencyCheck bool
}

type DefaultOptions struct {
	ExePath      string
	Arch         FileArch
	Options      []gr.Option
	Dependencies DependencyStatus
}

func AutoOptionsForExe(exePath string, cfg DefaultOptionsConfig) (DefaultOptions, error) {
	if exePath == "" {
		return DefaultOptions{}, fmt.Errorf("exe path is required")
	}

	absExe, err := filepath.Abs(exePath)
	if err != nil {
		return DefaultOptions{}, fmt.Errorf("resolve exe path: %w", err)
	}

	arch, err := DetectFileArch(absExe)
	if err != nil {
		return DefaultOptions{}, err
	}

	deps := CheckDependencyBins(cfg.WineBin, cfg.GamescopeBin, cfg.WinetricksBin, cfg.UseGamescope)
	if !cfg.SkipDependencyCheck {
		requireWinetricks := cfg.RequireWinetricks || len(cfg.Dependencies) > 0
		if err := deps.Verify(cfg.UseGamescope, requireWinetricks); err != nil {
			return DefaultOptions{}, err
		}
	}

	envCfg := cfg.Env
	usedDefaultEnv := false
	if isZeroWineEnvConfig(envCfg) {
		envCfg = DefaultWineEnvConfig()
		usedDefaultEnv = true
	}
	if usedDefaultEnv || envCfg.Lang == "" {
		if lang, err := DetectWineLang(absExe); err == nil && lang != "" {
			envCfg.Lang = lang
		}
	}

	opts := make([]gr.Option, 0, 5)
	if cfg.WinePrefix != "" {
		opts = append(opts, gr.WithWinePrefix(cfg.WinePrefix))
	}
	workingDir := cfg.WorkingDir
	if workingDir == "" {
		workingDir = filepath.Dir(absExe)
	}
	if workingDir != "" {
		opts = append(opts, gr.WithWorkingDir(workingDir))
	}
	if wineArch := defaultWineArch(arch, cfg); wineArch != "" {
		opts = append(opts, gr.WithSystemArch(wineArch))
	}
	if env := RecommendedWineEnv(envCfg); len(env) > 0 {
		opts = append(opts, gr.WithEnv(env...))
	}
	if len(cfg.Dependencies) > 0 {
		opts = append(opts, gr.WithDependencies(cfg.Dependencies...))
	}
	if len(cfg.Args) > 0 {
		opts = append(opts, gr.WithArgs(cfg.Args...))
	}

	return DefaultOptions{
		ExePath:      absExe,
		Arch:         arch,
		Options:      opts,
		Dependencies: deps,
	}, nil
}

func isZeroWineEnvConfig(cfg WineEnvConfig) bool {
	return cfg.Lang == "" &&
		cfg.WineDebug == "" &&
		!cfg.DisableWineMenuBuild &&
		!cfg.QuietDXVKLogs &&
		!cfg.UnattendedWinetricks &&
		len(cfg.Extra) == 0
}

func defaultWineArch(arch FileArch, cfg DefaultOptionsConfig) string {
	if cfg.ForceWineArch {
		return arch.WineArch()
	}

	if arch == ArchWin64 {
		return arch.WineArch()
	}

	return ""
}
