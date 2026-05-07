package installer

import (
	"os"
	"path/filepath"
	"testing"
)

func TestInstallerReasonsFromFilename(t *testing.T) {
	path := filepath.Join(t.TempDir(), "GameSetup.exe")
	if err := os.WriteFile(path, []byte("game"), 0o644); err != nil {
		t.Fatal(err)
	}

	reasons, err := installerReasons(path)
	if err != nil {
		t.Fatal(err)
	}

	if len(reasons) == 0 {
		t.Fatal("expected filename installer reason")
	}
}

func TestInstallerReasonsAvoidsUninstallFilename(t *testing.T) {
	path := filepath.Join(t.TempDir(), "uninstall.exe")
	if err := os.WriteFile(path, []byte("game"), 0o644); err != nil {
		t.Fatal(err)
	}

	reasons, err := installerReasons(path)
	if err != nil {
		t.Fatal(err)
	}

	if len(reasons) != 0 {
		t.Fatalf("expected no installer reasons, got %#v", reasons)
	}
}

func TestInstallerReasonsFromContent(t *testing.T) {
	path := filepath.Join(t.TempDir(), "game.exe")
	if err := os.WriteFile(path, []byte("created with Inno Setup"), 0o644); err != nil {
		t.Fatal(err)
	}

	reasons, err := installerReasons(path)
	if err != nil {
		t.Fatal(err)
	}

	if len(reasons) != 1 {
		t.Fatalf("expected one content reason, got %#v", reasons)
	}
}
