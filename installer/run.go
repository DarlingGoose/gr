package installer

import (
	"context"
	"fmt"

	"github.com/DarlingGoose/gr"
	"github.com/DarlingGoose/gr/autorunner"
)

type RunConfig struct {
	InstallerPath string
	GamePath      string

	Auto autorunner.DefaultOptionsConfig

	InstallerArgs []string
	GameArgs      []string

	ForceInstaller bool
	SkipInstaller  bool
}

type RunPlan struct {
	InstallerDetection Detection
	InstallerOptions   autorunner.DefaultOptions
	GameOptions        autorunner.DefaultOptions
}

func Plan(cfg RunConfig) (RunPlan, error) {
	if cfg.GamePath == "" {
		return RunPlan{}, fmt.Errorf("game path is required")
	}

	var plan RunPlan

	if !cfg.SkipInstaller && cfg.InstallerPath != "" {
		detection, err := Detect(cfg.InstallerPath)
		if err != nil {
			return RunPlan{}, err
		}

		if detection.Kind != KindInstaller && !cfg.ForceInstaller {
			return RunPlan{}, ErrNotInstaller
		}

		installerCfg := cfg.Auto
		installerCfg.Args = append([]string(nil), cfg.InstallerArgs...)

		opts, err := autorunner.AutoOptionsForExe(detection.Path, installerCfg)
		if err != nil {
			return RunPlan{}, err
		}

		plan.InstallerDetection = detection
		plan.InstallerOptions = opts
	}

	gameCfg := cfg.Auto
	gameCfg.Args = append([]string(nil), cfg.GameArgs...)

	gameOpts, err := autorunner.AutoOptionsForExe(cfg.GamePath, gameCfg)
	if err != nil {
		return RunPlan{}, err
	}
	plan.GameOptions = gameOpts

	return plan, nil
}

func InstallThenRun(ctx context.Context, runner gr.Runner, cfg RunConfig) (RunPlan, error) {
	if runner == nil {
		return RunPlan{}, fmt.Errorf("runner is required")
	}

	plan, err := Plan(cfg)
	if err != nil {
		return RunPlan{}, err
	}

	if plan.InstallerOptions.ExePath != "" {
		if err := runner.Run(ctx, plan.InstallerOptions.ExePath, plan.InstallerOptions.Options...); err != nil {
			return plan, fmt.Errorf("run installer: %w", err)
		}
	}

	if err := runner.Run(ctx, plan.GameOptions.ExePath, plan.GameOptions.Options...); err != nil {
		return plan, fmt.Errorf("run game: %w", err)
	}

	return plan, nil
}
