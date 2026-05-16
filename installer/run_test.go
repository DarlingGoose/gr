package installer

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/DarlingGoose/gr"
	"github.com/DarlingGoose/gr/autorunner"
)

func TestNewRunConfigDefaults(t *testing.T) {
	cfg := NewRunConfig("Game.exe")

	if cfg.GamePath != "Game.exe" {
		t.Fatalf("GamePath = %q, want %q", cfg.GamePath, "Game.exe")
	}
	if cfg.UnrarBin != DefaultUnrarBin {
		t.Fatalf("UnrarBin = %q, want %q", cfg.UnrarBin, DefaultUnrarBin)
	}
	if cfg.SevenZipBin != DefaultSevenZipBin {
		t.Fatalf("SevenZipBin = %q, want %q", cfg.SevenZipBin, DefaultSevenZipBin)
	}
	if cfg.SkipArchiveExtraction {
		t.Fatal("SkipArchiveExtraction = true, want false")
	}
	if cfg.SkipInstaller {
		t.Fatal("SkipInstaller = true, want false")
	}
	if cfg.Auto.Env.Lang != "" {
		t.Fatalf("Auto.Env.Lang = %q, want empty so locale detection can run", cfg.Auto.Env.Lang)
	}
}

func TestInstallThenRunExtractsSelfExtractingArchiveAndResolvesRelativeGamePath(t *testing.T) {
	dir := t.TempDir()
	archivePath := filepath.Join(dir, "RJ204938.part1.exe")
	writeMinimalPERARSFX(t, archivePath)

	extractDir := filepath.Join(dir, "extract")
	if err := os.MkdirAll(extractDir, 0o755); err != nil {
		t.Fatal(err)
	}
	gamePath := filepath.Join(extractDir, "Game.exe")
	if err := os.WriteFile(gamePath, minimalPE(0x14c), 0o644); err != nil {
		t.Fatal(err)
	}

	unrarPath := filepath.Join(dir, "unrar")
	if err := os.WriteFile(unrarPath, []byte("#!/bin/sh\nexit 0\n"), 0o755); err != nil {
		t.Fatal(err)
	}

	runner := &recordingRunner{}
	plan, err := InstallThenRun(context.Background(), runner, RunConfig{
		InstallerPath: archivePath,
		GamePath:      "Game.exe",
		ExtractDir:    extractDir,
		UnrarBin:      unrarPath,
		Auto: autorunner.DefaultOptionsConfig{
			SkipDependencyCheck: true,
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	if plan.ArchiveExtraction == nil {
		t.Fatal("ArchiveExtraction = nil, want extraction")
	}
	if plan.GameOptions.ExePath != gamePath {
		t.Fatalf("GameOptions.ExePath = %q, want %q", plan.GameOptions.ExePath, gamePath)
	}
	if len(runner.commands) != 1 || runner.commands[0] != gamePath {
		t.Fatalf("runner commands = %#v, want only %q", runner.commands, gamePath)
	}
}

type recordingRunner struct {
	commands []string
}

func (r *recordingRunner) Run(_ context.Context, cmd string, _ ...gr.Option) (*gr.Process, error) {
	r.commands = append(r.commands, cmd)
	return &gr.Process{ImageName: filepath.Base(cmd), Status: gr.StatusExited}, nil
}

func (r *recordingRunner) List(context.Context, ...gr.Option) ([]*gr.Process, error) {
	return nil, nil
}

func (r *recordingRunner) Find(context.Context, ...gr.Option) (*gr.Process, error) {
	return nil, nil
}

func (r *recordingRunner) GetOption(string) (interface{}, error) {
	return nil, nil
}

func (r *recordingRunner) GetOptionKeys() ([]string, error) {
	return nil, nil
}
