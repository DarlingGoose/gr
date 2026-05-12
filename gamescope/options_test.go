package gamescope

import (
	"context"
	"path/filepath"
	"reflect"
	"testing"
	"time"

	"github.com/DarlingGoose/gr"
)

func TestApplyOptionsAndGetOptions(t *testing.T) {
	r := New(
		WithName("custom-gamescope"),
		WithGamescopeBin("gamescope-custom"),
		WithWineBin("wine-custom"),
		WithWine(true),
		WithDefaultWinePrefix("/tmp/prefix"),
		WithResolution(1280, 720),
		WithOutputResolution(1920, 1080),
		WithRefreshRate(60),
		WithFullscreen(true),
		WithBorderless(true),
		WithForceGrab(true),
		WithSteamDeckMode(true),
		WithExposeWayland(true),
		WithScaler("stretch"),
		WithFilter("nearest"),
		WithExtraArgs("--one", "--two"),
	)

	o := r.GetOptions()

	if got := o.Name; got != "custom-gamescope" {
		t.Fatalf("Name = %q, want %q", got, "custom-gamescope")
	}
	if got := o.GamescopeBin; got != "gamescope-custom" {
		t.Fatalf("GamescopeBin = %q, want %q", got, "gamescope-custom")
	}
	if got := o.WineBin; got != "wine-custom" {
		t.Fatalf("WineBin = %q, want %q", got, "wine-custom")
	}
	if !o.UseWine {
		t.Fatal("UseWine = false, want true")
	}
	if got := o.DefaultWinePrefix; got != "/tmp/prefix" {
		t.Fatalf("DefaultWinePrefix = %q, want %q", got, "/tmp/prefix")
	}
	if o.Width != 1280 || o.Height != 720 {
		t.Fatalf("resolution = %dx%d, want 1280x720", o.Width, o.Height)
	}
	if o.OutputWidth != 1920 || o.OutputHeight != 1080 {
		t.Fatalf("output resolution = %dx%d, want 1920x1080", o.OutputWidth, o.OutputHeight)
	}
	if got := o.RefreshRate; got != 60 {
		t.Fatalf("RefreshRate = %d, want 60", got)
	}
	if !o.Fullscreen || !o.Borderless || !o.ForceGrab || !o.SteamDeckMode || !o.ExposeWayland {
		t.Fatal("boolean options were not all enabled")
	}
	if got := o.Scaler; got != "stretch" {
		t.Fatalf("Scaler = %q, want %q", got, "stretch")
	}
	if got := o.Filter; got != "nearest" {
		t.Fatalf("Filter = %q, want %q", got, "nearest")
	}

	o.ExtraArgs[0] = "changed"
	if got := r.GetOptions().ExtraArgs[0]; got != "--one" {
		t.Fatalf("GetOptions returned mutable ExtraArgs, first arg = %q", got)
	}
}

func TestApplyOptionsDefaults(t *testing.T) {
	o := ApplyOptions()

	if got := o.Name; got != "gamescope" {
		t.Fatalf("Name = %q, want %q", got, "gamescope")
	}
	if got := o.GamescopeBin; got != "gamescope" {
		t.Fatalf("GamescopeBin = %q, want %q", got, "gamescope")
	}
	if got := o.WineBin; got != "wine" {
		t.Fatalf("WineBin = %q, want %q", got, "wine")
	}
	if got := o.WineServerBin; got != "wineserver" {
		t.Fatalf("WineServerBin = %q, want %q", got, "wineserver")
	}
	if !o.WineStartWait {
		t.Fatal("WineStartWait = false, want true")
	}
	if !o.KillWineOnExit {
		t.Fatal("KillWineOnExit = false, want true")
	}
}

func TestGamescopeArgsIncludesScalerAndFilter(t *testing.T) {
	r := New(
		WithResolution(800, 600),
		WithFullscreen(true),
		WithScaler("stretch"),
		WithFilter("nearest"),
	)

	want := []string{"-w", "800", "-h", "600", "-f", "-S", "stretch", "-F", "nearest"}
	if got := r.gamescopeArgs(); !reflect.DeepEqual(got, want) {
		t.Fatalf("gamescopeArgs() = %#v, want %#v", got, want)
	}
}

func TestWineCommandUsesStartWaitByDefault(t *testing.T) {
	r := New(WithWine(true))

	want := []string{"wine", "start", "/wait", "/unix", "game.exe", "-arg"}
	if got := r.wineCommand("game.exe", []string{"-arg"}); !reflect.DeepEqual(got, want) {
		t.Fatalf("wineCommand() = %#v, want %#v", got, want)
	}
}

func TestWineCommandCanDisableStartWait(t *testing.T) {
	r := New(
		WithWine(true),
		WithWineStartWait(false),
	)

	want := []string{"wine", "game.exe", "-arg"}
	if got := r.wineCommand("game.exe", []string{"-arg"}); !reflect.DeepEqual(got, want) {
		t.Fatalf("wineCommand() = %#v, want %#v", got, want)
	}
}

func TestRunReturnsProcess(t *testing.T) {
	r := New(WithGamescopeBin("/bin/true"))

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
	r := New(WithGamescopeBin("/bin/true"))

	proc, err := r.Run(context.Background(), "ignored", gr.WithWorkingDir(workingDir))
	if err != nil {
		t.Fatal(err)
	}

	if got := proc.Cmd.Dir; got != workingDir {
		t.Fatalf("Cmd.Dir = %q, want %q", got, workingDir)
	}
}

func TestRunStopsOnContextCancel(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	r := New(
		WithGamescopeBin("/bin/sh"),
		WithExtraArgs("-c", "trap 'exit 0' TERM; sleep 30"),
	)

	start := time.Now()
	proc, err := r.Run(ctx, "ignored")
	if err == nil {
		t.Fatal("Run error = nil, want context cancellation error")
	}
	if proc == nil {
		t.Fatal("Run returned nil process")
	}
	if elapsed := time.Since(start); elapsed > 2*time.Second {
		t.Fatalf("Run took %s after context cancellation, want under 2s", elapsed)
	}
}

func TestSaveAndLoadRunner(t *testing.T) {
	path := filepath.Join(t.TempDir(), "gamescope.json")
	r := New(
		WithName("custom-gamescope"),
		WithGamescopeBin("gamescope-custom"),
		WithWineBin("wine-custom"),
		WithWineServerBin("wineserver-custom"),
		WithWine(true),
		WithWineStartWait(false),
		WithKillWineOnExit(false),
		WithDefaultWinePrefix("/tmp/prefix"),
		WithResolution(1280, 720),
		WithOutputResolution(1920, 1080),
		WithRefreshRate(60),
		WithFullscreen(true),
		WithScaler("stretch"),
		WithFilter("nearest"),
		WithExtraArgs("--one", "--two"),
	)

	if err := r.Save(path); err != nil {
		t.Fatal(err)
	}

	loaded, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}

	o := loaded.GetOptions()
	if got := o.Name; got != "custom-gamescope" {
		t.Fatalf("Name = %q, want %q", got, "custom-gamescope")
	}
	if got := o.GamescopeBin; got != "gamescope-custom" {
		t.Fatalf("GamescopeBin = %q, want %q", got, "gamescope-custom")
	}
	if got := o.WineBin; got != "wine-custom" {
		t.Fatalf("WineBin = %q, want %q", got, "wine-custom")
	}
	if got := o.WineServerBin; got != "wineserver-custom" {
		t.Fatalf("WineServerBin = %q, want %q", got, "wineserver-custom")
	}
	if !o.UseWine {
		t.Fatal("UseWine = false, want true")
	}
	if o.WineStartWait {
		t.Fatal("WineStartWait = true, want false")
	}
	if o.KillWineOnExit {
		t.Fatal("KillWineOnExit = true, want false")
	}
	if got := o.DefaultWinePrefix; got != "/tmp/prefix" {
		t.Fatalf("DefaultWinePrefix = %q, want %q", got, "/tmp/prefix")
	}
	if o.Width != 1280 || o.Height != 720 {
		t.Fatalf("resolution = %dx%d, want 1280x720", o.Width, o.Height)
	}
	if o.OutputWidth != 1920 || o.OutputHeight != 1080 {
		t.Fatalf("output resolution = %dx%d, want 1920x1080", o.OutputWidth, o.OutputHeight)
	}
	if got := o.RefreshRate; got != 60 {
		t.Fatalf("RefreshRate = %d, want 60", got)
	}
	if !o.Fullscreen {
		t.Fatal("Fullscreen = false, want true")
	}
	if got := o.Scaler; got != "stretch" {
		t.Fatalf("Scaler = %q, want %q", got, "stretch")
	}
	if got := o.Filter; got != "nearest" {
		t.Fatalf("Filter = %q, want %q", got, "nearest")
	}
	if got := o.ExtraArgs; !reflect.DeepEqual(got, []string{"--one", "--two"}) {
		t.Fatalf("ExtraArgs = %#v, want [--one --two]", got)
	}
}
