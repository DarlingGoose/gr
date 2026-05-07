package autorunner

import "testing"

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
