package wine

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/DarlingGoose/gr"
)

func TestApplyOptionsAndGetOptions(t *testing.T) {
	r := New(
		WithName("custom-wine"),
		WithWineBin("wine-custom"),
		WithWineTricksBin("winetricks-custom"),
		WithDefaultPrefix("/tmp/prefix"),
	)

	o := r.GetOptions()

	if got := o.Name; got != "custom-wine" {
		t.Fatalf("Name = %q, want %q", got, "custom-wine")
	}
	if got := o.WineBin; got != "wine-custom" {
		t.Fatalf("WineBin = %q, want %q", got, "wine-custom")
	}
	if got := o.WineTricksBin; got != "winetricks-custom" {
		t.Fatalf("WineTricksBin = %q, want %q", got, "winetricks-custom")
	}
	if got := o.DefaultPrefix; got != "/tmp/prefix" {
		t.Fatalf("DefaultPrefix = %q, want %q", got, "/tmp/prefix")
	}
}

func TestApplyOptionsDefaults(t *testing.T) {
	o := ApplyOptions()

	if got := o.Name; got != "wine" {
		t.Fatalf("Name = %q, want %q", got, "wine")
	}
	if got := o.WineBin; got != "wine" {
		t.Fatalf("WineBin = %q, want %q", got, "wine")
	}
	if got := o.WineTricksBin; got != "winetricks" {
		t.Fatalf("WineTricksBin = %q, want %q", got, "winetricks")
	}
}

func TestRunReturnsProcess(t *testing.T) {
	r := New(
		WithWineBin("/bin/true"),
		WithDefaultPrefix(t.TempDir()),
	)

	proc, err := r.Run(context.Background(), "ignored")
	if err != nil {
		t.Fatal(err)
	}

	if proc == nil {
		t.Fatal("Run returned nil process")
	}
	if got := proc.ImageName; got != "/bin/true" {
		t.Fatalf("ImageName = %q, want %q", got, "/bin/true")
	}
	if got := proc.Status; got != gr.StatusExited {
		t.Fatalf("Status = %s, want %s", got, gr.StatusExited)
	}
	if proc.PID <= 0 {
		t.Fatalf("PID = %d, want positive PID", proc.PID)
	}
	if proc.Cmd == nil {
		t.Fatal("Cmd = nil")
	}
}

func TestRunSetsWorkingDir(t *testing.T) {
	workingDir := t.TempDir()
	r := New(
		WithWineBin("/bin/true"),
		WithDefaultPrefix(t.TempDir()),
	)

	proc, err := r.Run(context.Background(), "ignored", gr.WithWorkingDir(workingDir))
	if err != nil {
		t.Fatal(err)
	}

	if got := proc.Cmd.Dir; got != workingDir {
		t.Fatalf("Cmd.Dir = %q, want %q", got, workingDir)
	}
}

func TestSaveAndLoadRunner(t *testing.T) {
	path := filepath.Join(t.TempDir(), "wine.json")
	r := New(
		WithName("custom-wine"),
		WithWineBin("wine-custom"),
		WithWineTricksBin("winetricks-custom"),
		WithDefaultPrefix("/tmp/prefix"),
	)

	if err := r.Save(path); err != nil {
		t.Fatal(err)
	}

	loaded, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}

	o := loaded.GetOptions()
	if got := o.Name; got != "custom-wine" {
		t.Fatalf("Name = %q, want %q", got, "custom-wine")
	}
	if got := o.WineBin; got != "wine-custom" {
		t.Fatalf("WineBin = %q, want %q", got, "wine-custom")
	}
	if got := o.WineTricksBin; got != "winetricks-custom" {
		t.Fatalf("WineTricksBin = %q, want %q", got, "winetricks-custom")
	}
	if got := o.DefaultPrefix; got != "/tmp/prefix" {
		t.Fatalf("DefaultPrefix = %q, want %q", got, "/tmp/prefix")
	}
}
