package installer

import (
	"archive/zip"
	"encoding/binary"
	"os"
	"path/filepath"
	"testing"
)

func TestDetectArchiveRARSelfExtractingFirstVolume(t *testing.T) {
	path := filepath.Join(t.TempDir(), "RJ204938.part1.exe")
	writeMinimalPERARSFX(t, path)

	detection, err := DetectArchive(path)
	if err != nil {
		t.Fatal(err)
	}

	if detection.Kind != ArchiveRARSFX {
		t.Fatalf("Kind = %q, want %q", detection.Kind, ArchiveRARSFX)
	}
	if !detection.FirstVolume {
		t.Fatal("FirstVolume = false, want true")
	}
	if len(detection.Reasons) == 0 {
		t.Fatal("expected detection reasons")
	}
}

func TestDetectArchiveInstallShieldSelfExtracting(t *testing.T) {
	path := filepath.Join(t.TempDir(), "Goodbye Tired Stars 1.05.exe")
	writeMinimalPEInstallShieldSFX(t, path)

	detection, err := DetectArchive(path)
	if err != nil {
		t.Fatal(err)
	}

	if detection.Kind != ArchiveInstallShieldSFX {
		t.Fatalf("Kind = %q, want %q", detection.Kind, ArchiveInstallShieldSFX)
	}
	if len(detection.Reasons) == 0 {
		t.Fatal("expected detection reasons")
	}
}

func TestDetectArchiveZip(t *testing.T) {
	path := filepath.Join(t.TempDir(), "sonataria_ver.1.05.zip")
	writeZip(t, path, map[string]string{"Game.exe": "game"})

	detection, err := DetectArchive(path)
	if err != nil {
		t.Fatal(err)
	}

	if detection.Kind != ArchiveZip {
		t.Fatalf("Kind = %q, want %q", detection.Kind, ArchiveZip)
	}
}

func TestExtractArchiveZip(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "sonataria_ver.1.05.zip")
	writeZip(t, path, map[string]string{
		"Game.exe":        "game",
		"data/config.ini": "config",
	})

	destDir := filepath.Join(dir, "extract")
	extraction, err := ExtractArchive(t.Context(), path, ExtractConfig{
		DestDir: destDir,
	})
	if err != nil {
		t.Fatal(err)
	}

	if extraction.Detection.Kind != ArchiveZip {
		t.Fatalf("Kind = %q, want %q", extraction.Detection.Kind, ArchiveZip)
	}
	assertFileContent(t, filepath.Join(destDir, "Game.exe"), "game")
	assertFileContent(t, filepath.Join(destDir, "data", "config.ini"), "config")
}

func TestExtractArchiveZipRejectsTraversal(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "bad.zip")
	writeZip(t, path, map[string]string{
		"../escape.txt": "bad",
	})

	err := extractZip(path, filepath.Join(dir, "extract"))
	if err == nil {
		t.Fatal("expected traversal error")
	}
	if _, statErr := os.Stat(filepath.Join(dir, "escape.txt")); !os.IsNotExist(statErr) {
		t.Fatalf("escape file exists or stat failed unexpectedly: %v", statErr)
	}
}

func TestDefaultExtractDirStripsPartOneExe(t *testing.T) {
	got := defaultExtractDir(filepath.Join("/tmp", "RJ204938.part1.exe"))
	want := filepath.Join("/tmp", "RJ204938")
	if got != want {
		t.Fatalf("defaultExtractDir() = %q, want %q", got, want)
	}
}

func writeMinimalPERARSFX(t *testing.T, path string) {
	t.Helper()

	data := minimalPE(0x14c)
	data = append(data, []byte("Rar!\x1a\x07\x00")...)
	if err := os.WriteFile(path, data, 0o644); err != nil {
		t.Fatal(err)
	}
}

func writeMinimalPEInstallShieldSFX(t *testing.T, path string) {
	t.Helper()

	data := minimalPE(0x14c)
	data = append(data, []byte("InstallShield Self-Extracting Archive")...)
	if err := os.WriteFile(path, data, 0o644); err != nil {
		t.Fatal(err)
	}
}

func writeZip(t *testing.T, path string, files map[string]string) {
	t.Helper()

	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	w := zip.NewWriter(f)
	defer w.Close()

	for name, content := range files {
		entry, err := w.Create(name)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := entry.Write([]byte(content)); err != nil {
			t.Fatal(err)
		}
	}
}

func assertFileContent(t *testing.T, path, want string) {
	t.Helper()

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != want {
		t.Fatalf("%s = %q, want %q", path, string(data), want)
	}
}

func minimalPE(machine uint16) []byte {
	data := make([]byte, 0x100)
	copy(data[:2], "MZ")
	binary.LittleEndian.PutUint32(data[0x3c:], 0x80)
	copy(data[0x80:0x84], "PE\x00\x00")
	binary.LittleEndian.PutUint16(data[0x84:], machine)
	return data
}
