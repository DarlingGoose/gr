package autorunner

import "testing"

func TestNormalizeArch(t *testing.T) {
	tests := map[string]FileArch{
		"32":     ArchWin32,
		"x86":    ArchWin32,
		"win32":  ArchWin32,
		"64":     ArchWin64,
		"amd64":  ArchWin64,
		"x86_64": ArchWin64,
		"arm64":  ArchWin64,
		"":       ArchUnknown,
		"mips":   ArchUnknown,
	}

	for input, want := range tests {
		if got := NormalizeArch(input); got != want {
			t.Fatalf("NormalizeArch(%q) = %q, want %q", input, got, want)
		}
	}
}

func TestFileArchWineArch(t *testing.T) {
	if got := ArchWin32.WineArch(); got != "win32" {
		t.Fatalf("ArchWin32.WineArch() = %q", got)
	}

	if got := ArchWin64.WineArch(); got != "win64" {
		t.Fatalf("ArchWin64.WineArch() = %q", got)
	}

	if got := ArchUnknown.WineArch(); got != "" {
		t.Fatalf("ArchUnknown.WineArch() = %q", got)
	}
}
