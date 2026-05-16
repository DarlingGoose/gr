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
resource language IDs and known locale markers, including Japanese Shift-JIS
markers such as `SHIFTJIS_CHARSET`. For example, Japanese metadata will select
`LANG=ja_JP.UTF-8` and `LC_ALL=ja_JP.UTF-8`; otherwise it falls back to
`C.UTF-8`.

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

cfg := installer.NewRunConfig("Game.exe")
cfg.InstallerPath = "setup.exe"
cfg.Auto = autorunner.DefaultOptionsConfig{
	WinePrefix:   "/home/me/.local/share/gr/prefixes/game",
	UseGamescope: true,
	Dependencies: []string{"vcrun2022", "dxvk"},
}

_, err := installer.InstallThenRun(ctx, r, cfg)
```

`installer` validates that the setup executable looks like an installer before it
runs it. After the installer exits, it runs the game with generated Wine options
for the game executable.

Archives can be extracted before running the game. For a multipart RAR
self-extracting set such as `RJ204938.part1.exe`, `RJ204938.part2.rar`, and
later parts in the same directory, pass the first volume. `unrar x` will resolve
the remaining parts from that directory.

```go
cfg := installer.NewRunConfig("Game.exe")
cfg.ArchivePath = "RJ204938.part1.exe"
cfg.ExtractDir = "/home/me/.local/share/gr/games/RJ204938"
cfg.Auto = autorunner.DefaultOptionsConfig{
	WinePrefix: "/home/me/.local/share/gr/prefixes/rj204938",
}

_, err := installer.InstallThenRun(ctx, r, cfg)
```

Plain ZIP archives are extracted directly:

```go
cfg := installer.NewRunConfig("Game.exe")
cfg.ArchivePath = "sonataria_ver.1.05.zip"
cfg.ExtractDir = "/home/me/.local/share/gr/games/sonataria"

_, err := installer.InstallThenRun(ctx, r, cfg)
```

InstallShield self-extracting archives use `7z x`:

```go
cfg := installer.NewRunConfig("Game.exe")
cfg.ArchivePath = "Goodbye Tired Stars 1.05.exe"
cfg.ExtractDir = "/home/me/.local/share/gr/games/goodbye-tired-stars"

_, err := installer.InstallThenRun(ctx, r, cfg)
```
