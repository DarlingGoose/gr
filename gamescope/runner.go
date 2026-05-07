package gamescope

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/DarlingGoose/gr"
	"github.com/DarlingGoose/gr/wine"
)

type Runner struct {
	GamescopeBin string
	WineBin      string

	DefaultWinePrefix string

	UseWine bool

	Width        int
	Height       int
	RefreshRate  int
	OutputWidth  int
	OutputHeight int

	Fullscreen bool
	Borderless bool
	ForceGrab  bool

	SteamDeckMode bool
	ExposeWayland bool

	ExtraArgs []string
}

func New(opts ...Option) *Runner {
	r := &Runner{
		GamescopeBin: "gamescope",
		WineBin:      "wine",
	}

	for _, opt := range opts {
		if opt != nil {
			opt(r)
		}
	}

	return r
}

func (r *Runner) Run(ctx context.Context, target string, opts ...gr.Option) error {
	o := gr.ApplyOptions(opts...)

	if target == "" {
		return errors.New("target command is required")
	}

	env := r.buildEnv(o)

	args := r.gamescopeArgs()

	args = append(args, "--")

	if r.UseWine {
		args = append(args, r.WineBin, target)
	} else {
		args = append(args, target)
	}

	args = append(args, o.Args()...)

	cmd := exec.CommandContext(ctx, r.GamescopeBin, args...)
	cmd.Env = env
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	if o.Background() {
		if err := cmd.Start(); err != nil {
			return fmt.Errorf("start gamescope: %w", err)
		}
		return nil
	}

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("run gamescope: %w", err)
	}

	return nil
}

func (r *Runner) List(ctx context.Context, opts ...gr.Option) ([]*gr.Process, error) {
	o := gr.ApplyOptions(opts...)

	if r.UseWine || o.WinePrefix() != "" || r.DefaultWinePrefix != "" {
		prefix := o.WinePrefix()
		if prefix == "" {
			prefix = r.DefaultWinePrefix
		}

		if prefix == "" {
			return nil, errors.New("wine prefix is required for wine process listing")
		}
		env := r.buildEnv(o)
		cmd := exec.CommandContext(ctx, r.WineBin, "tasklist")
		cmd.Env = env
		var stdout bytes.Buffer
		var stderr bytes.Buffer
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr
		if err := cmd.Run(); err != nil {
			return nil, fmt.Errorf("wine tasklist failed: %w: %s", err, strings.TrimSpace(stderr.String()))
		}
		process := wine.ParseTasklist(stdout.String())
		filtered := make([]*gr.Process, 0, len(process))
		for _, p := range process {
			if o.MatchProcess(p) {
				filtered = append(filtered, p)
			}
		}

		return filtered, nil
	}

	return listNativeProcesses(ctx, o)
}

func (r *Runner) Find(ctx context.Context, opts ...gr.Option) (*gr.Process, error) {
	procs, err := r.List(ctx, opts...)
	if err != nil {
		return nil, err
	}

	if len(procs) == 0 {
		return nil, nil
	}

	return procs[0], nil
}

func (r *Runner) gamescopeArgs() []string {
	args := make([]string, 0, 16)

	if r.Width > 0 {
		args = append(args, "-w", strconv.Itoa(r.Width))
	}

	if r.Height > 0 {
		args = append(args, "-h", strconv.Itoa(r.Height))
	}

	if r.OutputWidth > 0 {
		args = append(args, "-W", strconv.Itoa(r.OutputWidth))
	}

	if r.OutputHeight > 0 {
		args = append(args, "-H", strconv.Itoa(r.OutputHeight))
	}

	if r.RefreshRate > 0 {
		args = append(args, "-r", strconv.Itoa(r.RefreshRate))
	}

	if r.Fullscreen {
		args = append(args, "-f")
	}

	if r.Borderless {
		args = append(args, "-b")
	}

	if r.ForceGrab {
		args = append(args, "--force-grab-cursor")
	}

	if r.SteamDeckMode {
		args = append(args, "-e")
	}

	if r.ExposeWayland {
		args = append(args, "--expose-wayland")
	}

	args = append(args, r.ExtraArgs...)

	return args
}

func (r *Runner) buildEnv(o gr.Options) []string {
	env := os.Environ()

	prefix := o.WinePrefix()
	if prefix == "" {
		prefix = r.DefaultWinePrefix
	}

	if prefix != "" {
		env = upsertEnv(env, "WINEPREFIX", prefix)
	}

	switch normalizeArch(o.SystemArch()) {
	case "win32":
		env = upsertEnv(env, "WINEARCH", "win32")
	case "win64":
		env = upsertEnv(env, "WINEARCH", "win64")
	}

	for _, e := range o.Envs() {
		if strings.TrimSpace(e) == "" {
			continue
		}

		k, v, ok := strings.Cut(e, "=")
		if !ok {
			env = append(env, e)
			continue
		}

		env = upsertEnv(env, k, v)
	}

	return env
}

func normalizeArch(arch string) string {
	switch strings.ToLower(strings.TrimSpace(arch)) {
	case "32", "x86", "i386", "win32":
		return "win32"
	case "64", "x64", "amd64", "x86_64", "win64":
		return "win64"
	default:
		return ""
	}
}

func upsertEnv(env []string, key, value string) []string {
	prefix := key + "="

	for i, e := range env {
		if strings.HasPrefix(e, prefix) {
			env[i] = prefix + value
			return env
		}
	}

	return append(env, prefix+value)
}
