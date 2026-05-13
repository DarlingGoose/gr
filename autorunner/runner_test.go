package autorunner

import (
	"testing"

	"github.com/DarlingGoose/gr/gamescope"
	"github.com/DarlingGoose/gr/wine"
)

//func TestNewRunnerRequiresMissingWine(t *testing.T) {
//	_, err := newRunner("/tmp/prefix", DependencyStatus{})
//	if err == nil {
//		t.Fatal("newRunner error = nil, want error")
//	}
//}

func TestDefaultRunnerConfigUsesWineDefaults(t *testing.T) {
	cfg, err := defaultRunnerConfig("/tmp/prefix", DependencyStatus{WinePath: "/usr/bin/wine"}, []wine.Option{
		wine.WithName("custom-wine"),
	}, nil)
	if err != nil {
		t.Fatal(err)
	}

	if cfg.UseGamescope {
		t.Fatal("UseGamescope = true, want false")
	}
	if cfg.Wine == nil {
		t.Fatal("Wine = nil")
	}
	if cfg.Gamescope != nil {
		t.Fatalf("Gamescope = %#v, want nil", cfg.Gamescope)
	}
	if got := cfg.Wine.Name; got != "custom-wine" {
		t.Fatalf("Wine.Name = %q, want %q", got, "custom-wine")
	}
	if got := cfg.Wine.DefaultPrefix; got != "/tmp/prefix" {
		t.Fatalf("Wine.DefaultPrefix = %q, want %q", got, "/tmp/prefix")
	}
}

func TestDefaultRunnerConfigUsesGamescopeDefaults(t *testing.T) {
	cfg, err := defaultRunnerConfig("/tmp/prefix", DependencyStatus{
		WinePath:      "/usr/bin/wine",
		GamescopePath: "/usr/bin/gamescope",
	}, nil, []gamescope.Option{
		gamescope.WithName("custom-gamescope"),
		gamescope.WithResolution(800, 600),
	})
	if err != nil {
		t.Fatal(err)
	}

	if !cfg.UseGamescope {
		t.Fatal("UseGamescope = false, want true")
	}
	if cfg.Wine != nil {
		t.Fatalf("Wine = %#v, want nil", cfg.Wine)
	}
	if cfg.Gamescope == nil {
		t.Fatal("Gamescope = nil")
	}
	if got := cfg.Gamescope.Name; got != "custom-gamescope" {
		t.Fatalf("Gamescope.Name = %q, want %q", got, "custom-gamescope")
	}
	if got := cfg.Gamescope.DefaultWinePrefix; got != "/tmp/prefix" {
		t.Fatalf("Gamescope.DefaultWinePrefix = %q, want %q", got, "/tmp/prefix")
	}
	if !cfg.Gamescope.UseWine {
		t.Fatal("Gamescope.UseWine = false, want true")
	}
	if cfg.Gamescope.Width != 800 || cfg.Gamescope.Height != 600 {
		t.Fatalf("Gamescope resolution = %dx%d, want 800x600", cfg.Gamescope.Width, cfg.Gamescope.Height)
	}
}

func TestRunnerConfigFor(t *testing.T) {
	r := wine.New(
		wine.WithName("custom-wine"),
		wine.WithDefaultPrefix("/tmp/prefix"),
	)

	cfg, ok := RunnerConfigFor(r)
	if !ok {
		t.Fatal("RunnerConfigFor ok = false, want true")
	}
	if cfg.Wine == nil {
		t.Fatal("Wine = nil")
	}
	if got := cfg.Wine.Name; got != "custom-wine" {
		t.Fatalf("Wine.Name = %q, want %q", got, "custom-wine")
	}
	if got := cfg.Wine.DefaultPrefix; got != "/tmp/prefix" {
		t.Fatalf("Wine.DefaultPrefix = %q, want %q", got, "/tmp/prefix")
	}
}

//
//func TestNewRunnerUsesWineWhenWineInstalled(t *testing.T) {
//	r, err := newRunner("/tmp/prefix", DependencyStatus{
//		WinePath: "/usr/bin/wine",
//	})
//	if err != nil {
//		t.Fatal(err)
//	}
//
//	wineRunner, ok := r.(*wine.Runner)
//	if !ok {
//		t.Fatalf("runner type = %T, want *wine.Runner", r)
//	}
//	if got := wineRunner.DefaultPrefix; got != "/tmp/prefix" {
//		t.Fatalf("DefaultPrefix = %q, want %q", got, "/tmp/prefix")
//	}
//}
