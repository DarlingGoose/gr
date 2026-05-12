package gamescope

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/DarlingGoose/gr"
)

func TestRunForegroundDoesNotProbeWineTasklist(t *testing.T) {
	dir := t.TempDir()
	logPath := filepath.Join(dir, "wine.log")
	wineBin := filepath.Join(dir, "wine")

	script := `#!/bin/sh
printf '%s\n' "$*" > "$WINE_LOG"
cat <<'EOF'
Image Name                     PID Session Name        Session#    Mem Usage
========================= ======== ================ =========== ============
Game.exe                       123 Console                    1      10 K
EOF
`
	if err := os.WriteFile(wineBin, []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}

	r := New(
		WithGamescopeBin("/bin/true"),
		WithWineBin(wineBin),
		WithWine(true),
		WithDefaultWinePrefix(filepath.Join(dir, "prefix")),
	)

	proc, err := r.Run(context.Background(), "Game.exe", gr.WithEnv("WINE_LOG="+logPath))
	if err != nil {
		t.Fatal(err)
	}
	if proc == nil {
		t.Fatal("Run returned nil process")
	}
	if proc.WinePID != 0 {
		t.Fatalf("WinePID = %d, want 0 for foreground run", proc.WinePID)
	}
	if _, err := os.Stat(logPath); !os.IsNotExist(err) {
		t.Fatalf("wine tasklist was probed during foreground run, stat err = %v", err)
	}
}
