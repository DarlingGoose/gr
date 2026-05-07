package autorunner

import (
	"errors"
	"fmt"
	"os/exec"
)

type DependencyStatus struct {
	WineBin       string
	GamescopeBin  string
	WinetricksBin string

	WinePath       string
	GamescopePath  string
	WinetricksPath string
}

func CheckDependencies(useGamescope bool) DependencyStatus {
	return CheckDependencyBins("wine", "gamescope", "winetricks", useGamescope)
}

func CheckDependencyBins(wineBin, gamescopeBin, winetricksBin string, useGamescope bool) DependencyStatus {
	status := DependencyStatus{
		WineBin:       fallbackBin(wineBin, "wine"),
		GamescopeBin:  fallbackBin(gamescopeBin, "gamescope"),
		WinetricksBin: fallbackBin(winetricksBin, "winetricks"),
	}

	status.WinePath, _ = exec.LookPath(status.WineBin)
	status.WinetricksPath, _ = exec.LookPath(status.WinetricksBin)

	if useGamescope {
		status.GamescopePath, _ = exec.LookPath(status.GamescopeBin)
	}

	return status
}

func (s DependencyStatus) WineInstalled() bool {
	return s.WinePath != ""
}

func (s DependencyStatus) GamescopeInstalled() bool {
	return s.GamescopePath != ""
}

func (s DependencyStatus) WinetricksInstalled() bool {
	return s.WinetricksPath != ""
}

func (s DependencyStatus) Missing(useGamescope, requireWinetricks bool) []string {
	var missing []string

	if !s.WineInstalled() {
		missing = append(missing, s.WineBin)
	}

	if useGamescope && !s.GamescopeInstalled() {
		missing = append(missing, s.GamescopeBin)
	}

	if requireWinetricks && !s.WinetricksInstalled() {
		missing = append(missing, s.WinetricksBin)
	}

	return missing
}

func (s DependencyStatus) Verify(useGamescope, requireWinetricks bool) error {
	missing := s.Missing(useGamescope, requireWinetricks)
	if len(missing) == 0 {
		return nil
	}

	return fmt.Errorf("%w: %v", ErrMissingDependencies, missing)
}

var ErrMissingDependencies = errors.New("missing runtime dependencies")

func fallbackBin(bin, fallback string) string {
	if bin == "" {
		return fallback
	}
	return bin
}
