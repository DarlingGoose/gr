package wine

import "testing"

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
