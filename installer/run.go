package installer

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/DarlingGoose/gr"
	"github.com/DarlingGoose/gr/autorunner"
)

type RunConfig struct {
	ArchivePath   string
	InstallerPath string
	GamePath      string
	ExtractDir    string

	Auto autorunner.DefaultOptionsConfig

	InstallerArgs []string
	GameArgs      []string

	UnrarBin              string
	SevenZipBin           string
	ForceInstaller        bool
	SkipInstaller         bool
	SkipArchiveExtraction bool
}

type RunPlan struct {
	ArchiveExtraction  *ArchiveExtraction
	InstallerDetection Detection
	InstallerOptions   autorunner.DefaultOptions
	GameOptions        autorunner.DefaultOptions
}

func NewRunConfig(gamePath string) RunConfig {
	return RunConfig{
		GamePath:    gamePath,
		UnrarBin:    DefaultUnrarBin,
		SevenZipBin: DefaultSevenZipBin,
	}
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

	var archiveExtraction *ArchiveExtraction
	if !cfg.SkipArchiveExtraction {
		preparedCfg, extraction, err := prepareArchive(ctx, cfg)
		if err != nil {
			return RunPlan{}, err
		}
		cfg = preparedCfg
		archiveExtraction = extraction
	}

	plan, err := Plan(cfg)
	if err != nil {
		return RunPlan{}, err
	}
	plan.ArchiveExtraction = archiveExtraction

	if plan.InstallerOptions.ExePath != "" {
		if _, err := runner.Run(ctx, plan.InstallerOptions.ExePath, plan.InstallerOptions.Options...); err != nil {
			return plan, fmt.Errorf("run installer: %w", err)
		}
	}

	if _, err := runner.Run(ctx, plan.GameOptions.ExePath, plan.GameOptions.Options...); err != nil {
		return plan, fmt.Errorf("run game: %w", err)
	}

	return plan, nil
}

func prepareArchive(ctx context.Context, cfg RunConfig) (RunConfig, *ArchiveExtraction, error) {
	archivePath := cfg.ArchivePath
	installerIsArchive := false
	if archivePath == "" && cfg.InstallerPath != "" && looksLikeSupportedArchive(cfg.InstallerPath) {
		archivePath = cfg.InstallerPath
		installerIsArchive = true
	}
	if archivePath == "" {
		return cfg, nil, nil
	}

	extraction, err := ExtractArchive(ctx, archivePath, ExtractConfig{
		DestDir:     cfg.ExtractDir,
		UnrarBin:    cfg.UnrarBin,
		SevenZipBin: cfg.SevenZipBin,
	})
	if err != nil {
		return RunConfig{}, nil, err
	}

	if installerIsArchive {
		cfg.InstallerPath = ""
	}
	cfg.GamePath = resolveExtractedPath(extraction.DestDir, cfg.GamePath)
	cfg.InstallerPath = resolveExtractedPath(extraction.DestDir, cfg.InstallerPath)

	return cfg, &extraction, nil
}

func resolveExtractedPath(destDir, path string) string {
	if path == "" || filepath.IsAbs(path) {
		return path
	}
	if _, err := os.Stat(path); err == nil {
		return path
	}
	return filepath.Join(destDir, path)
}
