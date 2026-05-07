package autorunner

import (
	"errors"
	"testing"
)

func TestDependencyStatusMissingAndVerify(t *testing.T) {
	status := DependencyStatus{
		WineBin:       "wine",
		GamescopeBin:  "gamescope",
		WinetricksBin: "winetricks",
	}

	missing := status.Missing(true, true)
	want := []string{"wine", "gamescope", "winetricks"}
	if len(missing) != len(want) {
		t.Fatalf("len(missing) = %d, want %d: %#v", len(missing), len(want), missing)
	}

	for i := range want {
		if missing[i] != want[i] {
			t.Fatalf("missing[%d] = %q, want %q", i, missing[i], want[i])
		}
	}

	if err := status.Verify(true, true); !errors.Is(err, ErrMissingDependencies) {
		t.Fatalf("Verify() error = %v, want ErrMissingDependencies", err)
	}
}

func TestDependencyStatusInstalled(t *testing.T) {
	status := DependencyStatus{
		WinePath:       "/usr/bin/wine",
		GamescopePath:  "/usr/bin/gamescope",
		WinetricksPath: "/usr/bin/winetricks",
	}

	if err := status.Verify(true, true); err != nil {
		t.Fatalf("Verify() error = %v", err)
	}
}
