package gamescope

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/DarlingGoose/gr"
	"github.com/DarlingGoose/gr/wine"
)

type Runner struct {
	Options
}

func New(opts ...Option) *Runner {
	return &Runner{Options: ApplyOptions(opts...)}
}

func (r *Runner) GetOptions() Options {
	o := r.Options
	o.ExtraArgs = append([]string(nil), r.ExtraArgs...)
	return o
}

func (r *Runner) Save(path string) error {
	return r.GetOptions().Save(path)
}

func (r *Runner) GetOption(key string) (interface{}, error) {
	return gr.GetOption(r.Options, key)

}
func (r *Runner) GetOptionKeys() ([]string, error) {
	return gr.GetOptionKeys(r.Options)
}

func (r *Runner) Run(ctx context.Context, target string, opts ...gr.Option) (*gr.Process, error) {
	o := gr.ApplyOptions(opts...)

	if target == "" {
		return nil, errors.New("target command is required")
	}

	env := r.buildEnv(o)
	prefix := r.winePrefix(o)

	args := r.gamescopeArgs()

	args = append(args, "--")

	if r.UseWine {
		args = append(args, r.wineCommand(target, o.Args())...)
	} else {
		args = append(args, target)
		args = append(args, o.Args()...)
	}

	cmd := exec.CommandContext(ctx, r.GamescopeBin, args...)
	cmd.Env = env
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	if o.WorkingDir() != "" {
		cmd.Dir = o.WorkingDir()
	}
	prepareGamescopeCommand(cmd)

	if o.Background() {
		if err := cmd.Start(); err != nil {
			return nil, fmt.Errorf("start gamescope: %w", err)
		}
		proc := processFromCmd(cmd, r.GamescopeBin, args, env, gr.StatusRunning)
		if pr, _ := r.List(ctx, gr.WithName(filepath.Base(target))); len(pr) == 1 {
			proc.WinePID = pr[0].PID
		}
		stopGamescopeOnCancel(ctx, proc.PID, r.wineCleanup(prefix, env))
		return proc, nil
	}

	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("start gamescope: %w", err)
	}

	proc := processFromCmd(cmd, r.GamescopeBin, args, env, gr.StatusRunning)
	if pr, _ := r.List(ctx, gr.WithName(filepath.Base(target))); len(pr) == 1 {
		proc.WinePID = pr[0].PID
	}
	cleanup := r.wineCleanup(prefix, env)
	defer cleanup()
	defer cleanupProcessGroup(proc.PID)

	if err := cmd.Wait(); err != nil {
		proc.Status = gr.StatusExited
		return proc, fmt.Errorf("run gamescope: %w", err)
	}

	proc.Status = gr.StatusExited
	return proc, nil
}

func (r *Runner) wineCommand(target string, targetArgs []string) []string {
	args := []string{r.WineBin}
	if r.WineStartWait {
		args = append(args, "start", "/wait", "/unix")
	}
	args = append(args, target)
	args = append(args, targetArgs...)
	return args
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
				p.WinePID = p.PID
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

	if r.Scaler != "" {
		args = append(args, "-S", r.Scaler)
	}

	if r.Filter != "" {
		args = append(args, "-F", r.Filter)
	}

	args = append(args, r.ExtraArgs...)

	return args
}

func (r *Runner) buildEnv(o gr.Options) []string {
	env := os.Environ()

	prefix := r.winePrefix(o)

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

func (r *Runner) winePrefix(o gr.Options) string {
	prefix := o.WinePrefix()
	if prefix == "" {
		prefix = r.DefaultWinePrefix
	}

	return prefix
}

func processFromCmd(cmd *exec.Cmd, imageName string, args []string, env []string, status gr.Status) *gr.Process {
	p := &gr.Process{
		ImageName: imageName,
		Status:    status,
		Cmdline:   append([]string{imageName}, args...),
		Environ:   append([]string(nil), env...),
		Cmd:       cmd,
	}

	if cmd.Process != nil {
		p.PID = cmd.Process.Pid
	}

	return p
}

const gamescopeShutdownGrace = 5 * time.Second

func prepareGamescopeCommand(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setpgid:   true,
		Pdeathsig: syscall.SIGTERM,
	}
	cmd.Cancel = func() error {
		if cmd.Process == nil {
			return os.ErrProcessDone
		}

		return terminateProcessGroup(cmd.Process.Pid)
	}
	cmd.WaitDelay = gamescopeShutdownGrace
}

func cleanupProcessGroup(pid int) {
	if pid <= 0 {
		return
	}

	_ = terminateProcessGroup(pid)
}

func stopGamescopeOnCancel(ctx context.Context, pid int, cleanup func()) {
	if ctx == nil || ctx.Done() == nil || pid <= 0 {
		return
	}

	go func() {
		<-ctx.Done()
		_ = terminateProcessGroup(pid)
		cleanup()
	}()
}

func (r *Runner) wineCleanup(prefix string, env []string) func() {
	if !r.UseWine || !r.KillWineOnExit || prefix == "" {
		return func() {}
	}

	return func() {
		ctx, cancel := context.WithTimeout(context.Background(), gamescopeShutdownGrace)
		defer cancel()

		cmd := exec.CommandContext(ctx, r.WineServerBin, "-k")
		cmd.Env = env
		_ = cmd.Run()
	}
}

func terminateProcessGroup(pid int) error {
	err := signalProcessGroup(pid, syscall.SIGTERM)
	time.AfterFunc(gamescopeShutdownGrace, func() {
		_ = signalProcessGroup(pid, syscall.SIGKILL)
	})
	return err
}

func signalProcessGroup(pid int, sig syscall.Signal) error {
	if pid <= 0 {
		return os.ErrProcessDone
	}

	if err := syscall.Kill(-pid, sig); err != nil {
		if errors.Is(err, syscall.ESRCH) {
			return os.ErrProcessDone
		}
		return err
	}

	return nil
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
