package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/DarlingGoose/gr"
	"github.com/DarlingGoose/gr/autorunner"
	"github.com/DarlingGoose/gr/gamescope"
	"github.com/DarlingGoose/gr/installer"
	"github.com/DarlingGoose/gr/wine"
)

func main() {
	ctx := context.Background()

	installerPath := flag.String("installer", "", "path to the setup exe")
	gamePath := flag.String("game", "", "optional installed game exe to run after setup exits")
	prefix := flag.String("prefix", defaultWinePrefix(), "wine prefix to use")
	useGamescope := flag.Bool("gamescope", false, "run through gamescope with wine")
	skipDeps := flag.Bool("skip-deps", false, "skip wine/gamescope/winetricks dependency checks")
	skipInstaller := flag.Bool("skip-installer", false, "skip setup and run the game directly")
	forceInstaller := flag.Bool("force-installer", false, "run setup even when the game exe already exists")
	resetWineServer := flag.Bool("reset-wineserver", false, "stop the wineserver for this prefix before running")
	flag.Parse()

	if *resetWineServer {
		if err := stopWineServer(ctx, *prefix); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	}

	r := newRunner(*useGamescope, *prefix)
	auto := autorunner.DefaultOptionsConfig{
		WinePrefix:          *prefix,
		UseGamescope:        *useGamescope,
		RequireWinetricks:   false,
		SkipDependencyCheck: *skipDeps,
	}

	if *gamePath == "" {
		if err := runInstallerOnly(ctx, r, *installerPath, auto); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		return
	}

	shouldSkipInstaller := *skipInstaller || (!*forceInstaller && fileExists(*gamePath))

	_, err := installer.InstallThenRun(ctx, r, installer.RunConfig{
		InstallerPath: *installerPath,
		GamePath:      *gamePath,
		Auto:          auto,
		SkipInstaller: shouldSkipInstaller,
	})
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func fileExists(path string) bool {
	if path == "" {
		return false
	}

	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}

func runInstallerOnly(ctx context.Context, r gr.Runner, setupPath string, auto autorunner.DefaultOptionsConfig) error {
	detection, err := installer.VerifyInstaller(setupPath)
	if err != nil {
		return err
	}

	opts, err := autorunner.AutoOptionsForExe(detection.Path, auto)
	if err != nil {
		return err
	}

	return r.Run(ctx, opts.ExePath, opts.Options...)
}

func newRunner(useGamescope bool, prefix string) gr.Runner {
	if useGamescope {
		return gamescope.New(
			gamescope.WithWine(true),
			gamescope.WithDefaultWinePrefix(prefix),
			gamescope.WithResolution(1280, 720),
			gamescope.WithFullscreen(false),
		)
	}

	return wine.New(wine.WithDefaultPrefix(prefix))
}

func defaultWinePrefix() string {
	home, err := os.UserHomeDir()
	if err != nil || home == "" {
		return filepath.Join(os.TempDir(), "gr-prefix")
	}

	return filepath.Join(home, ".local", "share", "gr", "prefixes", "")
}

func stopWineServer(ctx context.Context, prefix string) error {
	cmd := exec.CommandContext(ctx, "wineserver", "-k")
	cmd.Env = upsertEnv(os.Environ(), "WINEPREFIX", prefix)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("stop wineserver: %w", err)
	}

	return nil
}

func upsertEnv(env []string, key, value string) []string {
	prefix := key + "="
	for i, existing := range env {
		if strings.HasPrefix(existing, prefix) {
			env[i] = prefix + value
			return env
		}
	}

	return append(env, prefix+value)
}
