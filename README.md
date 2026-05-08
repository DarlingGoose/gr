# GameRunner (GR)

```go
ctx := context.Background()

defaults, err := autorunner.AutoOptionsForExe("game.exe", autorunner.DefaultOptionsConfig{
	WinePrefix:   "/home/me/.local/share/gr/prefixes/game",
	UseGamescope: true,
	Dependencies: []string{"vcrun2022", "dxvk"},
})
if err != nil {
	return err
}

r := gamescope.New(
	gamescope.WithWine(true),
	gamescope.WithDefaultWinePrefix("/home/me/.local/share/gr/prefixes/game"),
	gamescope.WithResolution(1280, 720),
	gamescope.WithFullscreen(true),
)

proc, err := r.Run(ctx, defaults.ExePath, defaults.Options...)
if err != nil {
	return err
}
_ = proc
return nil
```

`autorunner` can detect PE executable architecture, build recommended Wine env
vars, check for `wine`, `gamescope`, and `winetricks`, and generate `gr.Option`
defaults for a Windows executable.

When `LANG` is not explicitly configured, `autorunner` tries to infer it from PE
resource language IDs and known locale markers. For example, Japanese metadata
will select `LANG=ja_JP.UTF-8`; otherwise it falls back to `C.UTF-8`.

For 32-bit executables, `autorunner` detects the architecture but does not force
`WINEARCH=win32` by default. Modern WoW64 Wine builds can run 32-bit executables
inside a 64-bit prefix and may reject pure win32 prefixes. Set
`ForceWineArch: true` only when you specifically need the old win32-prefix
behavior.

```go
r := gamescope.New(
	gamescope.WithWine(true),
	gamescope.WithDefaultWinePrefix("/home/me/.local/share/gr/prefixes/game"),
	gamescope.WithFullscreen(true),
)

_, err := installer.InstallThenRun(ctx, r, installer.RunConfig{
	InstallerPath: "setup.exe",
	GamePath:      "Game.exe",
	Auto: autorunner.DefaultOptionsConfig{
		WinePrefix:   "/home/me/.local/share/gr/prefixes/game",
		UseGamescope: true,
		Dependencies: []string{"vcrun2022", "dxvk"},
	},
})
```

`installer` validates that the setup executable looks like an installer before it
runs it. After the installer exits, it runs the game with generated Wine options
for the game executable.
