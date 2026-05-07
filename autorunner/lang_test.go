package autorunner

import (
	"os"
	"path/filepath"
	"testing"
)

func TestWineLangForWindowsLangID(t *testing.T) {
	tests := map[uint16]string{
		0x0411: "ja_JP.UTF-8",
		0x0412: "ko_KR.UTF-8",
		0x0804: "zh_CN.UTF-8",
		0x0404: "zh_TW.UTF-8",
		0x0409: "",
	}

	for langID, want := range tests {
		if got := wineLangForWindowsLangID(langID); got != want {
			t.Fatalf("wineLangForWindowsLangID(0x%x) = %q, want %q", langID, got, want)
		}
	}
}

func TestDetectWineLangFromMarkers(t *testing.T) {
	path := filepath.Join(t.TempDir(), "game.exe")
	if err := os.WriteFile(path, []byte("Japanese release"), 0o644); err != nil {
		t.Fatal(err)
	}

	lang, err := detectWineLangFromMarkers(path)
	if err != nil {
		t.Fatal(err)
	}

	if lang != "ja_JP.UTF-8" {
		t.Fatalf("detectWineLangFromMarkers() = %q", lang)
	}
}
