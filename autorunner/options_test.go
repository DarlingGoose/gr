package autorunner

import (
	"encoding/binary"
	"os"
	"path/filepath"
	"testing"

	"github.com/DarlingGoose/gr"
)

func TestDefaultWineArchSkipsWin32ForWOW64(t *testing.T) {
	if got := defaultWineArch(ArchWin32, DefaultOptionsConfig{}); got != "" {
		t.Fatalf("defaultWineArch(ArchWin32) = %q, want empty", got)
	}
}

func TestDefaultWineArchUsesWin64(t *testing.T) {
	if got := defaultWineArch(ArchWin64, DefaultOptionsConfig{}); got != "win64" {
		t.Fatalf("defaultWineArch(ArchWin64) = %q, want win64", got)
	}
}

func TestDefaultWineArchCanForceWin32(t *testing.T) {
	cfg := DefaultOptionsConfig{ForceWineArch: true}
	if got := defaultWineArch(ArchWin32, cfg); got != "win32" {
		t.Fatalf("defaultWineArch(ArchWin32, force) = %q, want win32", got)
	}
}

func TestAutoOptionsForExeDefaultsWorkingDirToExeDir(t *testing.T) {
	exePath := filepath.Join(t.TempDir(), "Game.exe")
	writeMinimalPE(t, exePath, 0x14c)

	defaults, err := AutoOptionsForExe(exePath, DefaultOptionsConfig{
		SkipDependencyCheck: true,
	})
	if err != nil {
		t.Fatal(err)
	}

	o := gr.ApplyOptions(defaults.Options...)
	if got, want := o.WorkingDir(), filepath.Dir(defaults.ExePath); got != want {
		t.Fatalf("WorkingDir() = %q, want %q", got, want)
	}
}

func TestAutoOptionsForExeUsesConfiguredWorkingDir(t *testing.T) {
	exePath := filepath.Join(t.TempDir(), "Game.exe")
	writeMinimalPE(t, exePath, 0x14c)

	workingDir := t.TempDir()
	defaults, err := AutoOptionsForExe(exePath, DefaultOptionsConfig{
		WorkingDir:          workingDir,
		SkipDependencyCheck: true,
	})
	if err != nil {
		t.Fatal(err)
	}

	o := gr.ApplyOptions(defaults.Options...)
	if got := o.WorkingDir(); got != workingDir {
		t.Fatalf("WorkingDir() = %q, want %q", got, workingDir)
	}
}

func TestAutoOptionsForExeDetectsJapaneseLocaleMarkers(t *testing.T) {
	exePath := filepath.Join(t.TempDir(), "Game.exe")
	writeMinimalPE(t, exePath, 0x14c)
	f, err := os.OpenFile(exePath, os.O_APPEND|os.O_WRONLY, 0)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := f.WriteString("Font.CharsetSHIFTJIS_CHARSET"); err != nil {
		_ = f.Close()
		t.Fatal(err)
	}
	if err := f.Close(); err != nil {
		t.Fatal(err)
	}

	defaults, err := AutoOptionsForExe(exePath, DefaultOptionsConfig{
		SkipDependencyCheck: true,
	})
	if err != nil {
		t.Fatal(err)
	}

	o := gr.ApplyOptions(defaults.Options...)
	wantEnv := map[string]bool{
		"LANG=ja_JP.UTF-8":   false,
		"LC_ALL=ja_JP.UTF-8": false,
	}
	for _, env := range o.Envs() {
		if _, ok := wantEnv[env]; ok {
			wantEnv[env] = true
		}
	}
	for env, found := range wantEnv {
		if !found {
			t.Fatalf("generated env missing %q: %#v", env, o.Envs())
		}
	}
}

func writeMinimalPE(t *testing.T, path string, machine uint16) {
	t.Helper()

	data := make([]byte, 0x100)
	copy(data[:2], "MZ")
	binary.LittleEndian.PutUint32(data[0x3c:], 0x80)
	copy(data[0x80:0x84], "PE\x00\x00")
	binary.LittleEndian.PutUint16(data[0x84:], machine)

	if err := os.WriteFile(path, data, 0o644); err != nil {
		t.Fatal(err)
	}
}
