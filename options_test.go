package gr

import "testing"

func TestOptionsGetters(t *testing.T) {
	o := ApplyOptions(
		WithBackground(true),
		WithWorkingDir("/tmp/game"),
		WithArgs("one", "two"),
		WithEnv("A=B", "C=D"),
		WithWinePrefix("/tmp/prefix"),
		WithSystemArch("win64"),
		WithDependencies("vcrun2022", "dxvk"),
		WithName("game.exe"),
		WithPID(1234),
		WithSession("console"),
		WithSessionID("1"),
	)

	if !o.Background() {
		t.Fatal("Background() = false, want true")
	}
	if got := o.WorkingDir(); got != "/tmp/game" {
		t.Fatalf("WorkingDir() = %q, want %q", got, "/tmp/game")
	}
	if got := o.WinePrefix(); got != "/tmp/prefix" {
		t.Fatalf("WinePrefix() = %q, want %q", got, "/tmp/prefix")
	}
	if got := o.SystemArch(); got != "win64" {
		t.Fatalf("SystemArch() = %q, want %q", got, "win64")
	}
	if got := o.Name(); got != "game.exe" {
		t.Fatalf("Name() = %q, want %q", got, "game.exe")
	}
	if got := o.PID(); got != 1234 {
		t.Fatalf("PID() = %d, want %d", got, 1234)
	}
	if got := o.Session(); got != "console" {
		t.Fatalf("Session() = %q, want %q", got, "console")
	}
	if got := o.SessionID(); got != "1" {
		t.Fatalf("SessionID() = %q, want %q", got, "1")
	}

	args := o.Args()
	args[0] = "changed"
	if got := o.Args()[0]; got != "one" {
		t.Fatalf("Args() returned mutable state, first arg = %q", got)
	}

	envs := o.Envs()
	envs[0] = "changed"
	if got := o.Envs()[0]; got != "A=B" {
		t.Fatalf("Envs() returned mutable state, first env = %q", got)
	}

	deps := o.Dependencies()
	deps[0] = "changed"
	if got := o.Dependencies()[0]; got != "vcrun2022" {
		t.Fatalf("Dependencies() returned mutable state, first dep = %q", got)
	}
}
