package gr

import (
	"encoding/json"
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
