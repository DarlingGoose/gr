package wine

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/DarlingGoose/gr"
)

type Runner struct {
	Options
}

func (r *Runner) GetOptionKeys() ([]string, error) {
	return gr.GetOptionKeys(r.Options)
}

func (r *Runner) GetOption(key string) (interface{}, error) {
	return gr.GetOption(r.Options, key)
}

func New(opts ...Option) *Runner {
	return &Runner{Options: ApplyOptions(opts...)}
}

func (r *Runner) GetOptions() Options {
	return r.Options
}

func (r *Runner) Save(path string) error {
	return r.GetOptions().Save(path)
}

func (r *Runner) Run(ctx context.Context, command string, opts ...gr.Option) (*gr.Process, error) {
	o := gr.ApplyOptions(opts...)

	prefix := o.WinePrefix()
	if prefix == "" {
		prefix = r.DefaultPrefix
	}

	if prefix == "" {
		return nil, errors.New("wine prefix is required")
	}

	if err := os.MkdirAll(prefix, 0o755); err != nil {
		return nil, fmt.Errorf("create wine prefix: %w", err)
	}

	env := r.buildEnv(prefix, o)

	if deps := o.Dependencies(); len(deps) > 0 {
		if err := r.installDeps(ctx, env, deps); err != nil {
			return nil, err
		}
	}

	args := append([]string{command}, o.Args()...)

	cmd := exec.CommandContext(ctx, r.WineBin, args...)
	cmd.Env = env
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	if o.WorkingDir() != "" {
		cmd.Dir = o.WorkingDir()
	}

	if o.Background() {
		if err := cmd.Start(); err != nil {
			return nil, fmt.Errorf("start wine command: %w", err)
		}
		proc := processFromCmd(cmd, r.WineBin, args, env, gr.StatusRunning)
		c := 10
		for {
			if pr, _ := r.List(ctx, gr.WithName(filepath.Base(gr.FindExe(args...))), gr.WithWinePrefix(o.WinePrefix())); len(pr) == 1 {
				proc.WinePID = pr[0].PID
				break
			}
			if c <= 0 {
				break
			}
			time.Sleep(2 * time.Second)
			c--
		}
		return proc, nil
	}

	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("start wine command: %w", err)
	}

	proc := processFromCmd(cmd, r.WineBin, args, env, gr.StatusRunning)
	c := 10
	for {
		if pr, _ := r.List(ctx, gr.WithName(filepath.Base(gr.FindExe(args...))), gr.WithWinePrefix(o.WinePrefix())); len(pr) == 1 {
			proc.WinePID = pr[0].PID
			break
		}
		if c <= 0 {
			break
		}
		time.Sleep(2 * time.Second)
		c--
	}
	if err := cmd.Wait(); err != nil {
		proc.Status = gr.StatusExited
		return proc, fmt.Errorf("run wine command: %w", err)
	}

	proc.Status = gr.StatusExited
	return proc, nil
}

func (r *Runner) List(ctx context.Context, opts ...gr.Option) ([]*gr.Process, error) {
	o := gr.ApplyOptions(opts...)

	prefix := o.WinePrefix()
	return List(ctx, o.WinePrefix(), r.buildEnv(prefix, o), opts...)

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

func (r *Runner) installDeps(ctx context.Context, env []string, deps []string) error {
	args := append([]string{"-q"}, deps...)

	cmd := exec.CommandContext(ctx, r.WineTricksBin, args...)
	cmd.Env = env
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("install winetricks deps %v: %w", deps, err)
	}

	return nil
}

func (r *Runner) buildEnv(prefix string, o gr.Options) []string {
	env := os.Environ()

	env = upsertEnv(env, "WINEPREFIX", prefix)

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

		k, _, ok := strings.Cut(e, "=")
		if !ok {
			env = append(env, e)
			continue
		}

		env = upsertEnv(env, k, e[len(k)+1:])
	}

	return env
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

var tasklistLineRE = regexp.MustCompile(`^(.+?)\s+([0-9]+)\s+(.+?)\s+([0-9]+)\s+(.+)$`)

func ParseTasklist(s string) []*gr.Process {
	sc := bufio.NewScanner(strings.NewReader(s))

	var out []*gr.Process
	for sc.Scan() {
		line := strings.TrimRight(sc.Text(), "\r\n")
		if line == "" ||
			strings.HasPrefix(line, "Image Name") ||
			strings.HasPrefix(line, "====") {
			continue
		}

		m := tasklistLineRE.FindStringSubmatch(line)
		if len(m) != 6 {
			continue
		}

		pid, err := strconv.Atoi(strings.TrimSpace(m[2]))
		if err != nil {
			continue
		}

		out = append(out, &gr.Process{
			ImageName: strings.TrimSpace(m[1]),
			PID:       pid,
			Session:   strings.TrimSpace(m[3]),
			SessionID: strings.TrimSpace(m[4]),
			MemUsage:  strings.TrimSpace(m[5]),
		})
	}

	return out
}
