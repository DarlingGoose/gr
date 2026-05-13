package gr

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestConfigRoundTrip(t *testing.T) {
	cfg := NewConfig(
		WithBackground(true),
		WithWorkingDir("/tmp/game"),
		WithArgs("one", "two"),
		WithEnv("A=B"),
		WithWinePrefix("/tmp/prefix"),
		WithSystemArch("win64"),
		WithDependencies("dxvk"),
		WithName("game.exe"),
		WithPID(123),
		WithSession("console"),
		WithSessionID("1"),
	)

	data, err := json.Marshal(cfg)
	if err != nil {
		t.Fatal(err)
	}

	var decoded Config
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatal(err)
	}

	o := ApplyOptions(decoded.Options()...)
	if !o.Background() {
		t.Fatal("Background() = false, want true")
	}
	if got := o.WorkingDir(); got != "/tmp/game" {
		t.Fatalf("WorkingDir() = %q, want %q", got, "/tmp/game")
	}
	if got := o.Args(); len(got) != 2 || got[0] != "one" || got[1] != "two" {
		t.Fatalf("Args() = %#v, want [one two]", got)
	}
	if got := o.Envs(); len(got) != 1 || got[0] != "A=B" {
		t.Fatalf("Envs() = %#v, want [A=B]", got)
	}
	if got := o.WinePrefix(); got != "/tmp/prefix" {
		t.Fatalf("WinePrefix() = %q, want %q", got, "/tmp/prefix")
	}
	if got := o.SystemArch(); got != "win64" {
		t.Fatalf("SystemArch() = %q, want %q", got, "win64")
	}
	if got := o.Dependencies(); len(got) != 1 || got[0] != "dxvk" {
		t.Fatalf("Dependencies() = %#v, want [dxvk]", got)
	}
	if got := o.Name(); got != "game.exe" {
		t.Fatalf("Name() = %q, want %q", got, "game.exe")
	}
	if got := o.PID(); got != 123 {
		t.Fatalf("PID() = %d, want 123", got)
	}
	if got := o.Session(); got != "console" {
		t.Fatalf("Session() = %q, want %q", got, "console")
	}
	if got := o.SessionID(); got != "1" {
		t.Fatalf("SessionID() = %q, want %q", got, "1")
	}
}

func TestConfigSaveLoadAndDelete(t *testing.T) {
	path := filepath.Join(t.TempDir(), "configs", "game.json")
	cfg := NewConfig(
		WithWorkingDir("/tmp/game"),
		WithArgs("one", "two"),
		WithWinePrefix("/tmp/prefix"),
	)

	if err := cfg.Save(path); err != nil {
		t.Fatal(err)
	}

	loaded, err := LoadConfig(path)
	if err != nil {
		t.Fatal(err)
	}

	o := ApplyOptions(loaded.Options()...)
	if got := o.WorkingDir(); got != "/tmp/game" {
		t.Fatalf("WorkingDir() = %q, want %q", got, "/tmp/game")
	}
	if got := o.Args(); len(got) != 2 || got[0] != "one" || got[1] != "two" {
		t.Fatalf("Args() = %#v, want [one two]", got)
	}
	if got := o.WinePrefix(); got != "/tmp/prefix" {
		t.Fatalf("WinePrefix() = %q, want %q", got, "/tmp/prefix")
	}

	if err := DeleteConfig(path); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(path); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("deleted config still exists, stat err = %v", err)
	}
	if err := DeleteConfig(path); err != nil {
		t.Fatalf("deleting missing config returned error: %v", err)
	}
}

func TestGameConfigSaveLoad(t *testing.T) {
	path := filepath.Join(t.TempDir(), "games", "game.json")
	cfg := NewGameConfig(
		"/games/game.exe",
		WithBackground(true),
		WithEnv("LANG=ja_JP.UTF-8"),
	)

	if err := cfg.Save(path); err != nil {
		t.Fatal(err)
	}

	loaded, err := LoadGameConfig(path)
	if err != nil {
		t.Fatal(err)
	}
	if got := loaded.ExePath; got != "/games/game.exe" {
		t.Fatalf("ExePath = %q, want %q", got, "/games/game.exe")
	}

	o := ApplyOptions(loaded.Options()...)
	if !o.Background() {
		t.Fatal("Background() = false, want true")
	}
	if got := o.Envs(); len(got) != 1 || got[0] != "LANG=ja_JP.UTF-8" {
		t.Fatalf("Envs() = %#v, want [LANG=ja_JP.UTF-8]", got)
	}
}
