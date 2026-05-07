package wine

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/DarlingGoose/gr"
)

type Runner struct {
	WineBin       string
	WineTricksBin string
	DefaultPrefix string
}

func New(opts ...Option) *Runner {
	r := &Runner{
		WineBin:       "wine",
		WineTricksBin: "winetricks",
	}

	for _, opt := range opts {
		if opt != nil {
			opt(r)
		}
	}

	return r
}

func (r *Runner) Run(ctx context.Context, command string, opts ...gr.Option) error {
	o := gr.ApplyOptions(opts...)

	prefix := o.WinePrefix()
	if prefix == "" {
		prefix = r.DefaultPrefix
	}

	if prefix == "" {
		return errors.New("wine prefix is required")
	}

	if err := os.MkdirAll(prefix, 0o755); err != nil {
		return fmt.Errorf("create wine prefix: %w", err)
	}

	env := r.buildEnv(prefix, o)

	if deps := o.Dependencies(); len(deps) > 0 {
		if err := r.installDeps(ctx, env, deps); err != nil {
			return err
		}
	}

	args := append([]string{command}, o.Args()...)

	cmd := exec.CommandContext(ctx, r.WineBin, args...)
	cmd.Env = env
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	if o.Background() {
		if err := cmd.Start(); err != nil {
			return fmt.Errorf("start wine command: %w", err)
		}
		return nil
	}

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("run wine command: %w", err)
	}

	return nil
}

func (r *Runner) List(ctx context.Context, opts ...gr.Option) ([]*gr.Process, error) {
	o := gr.ApplyOptions(opts...)

	prefix := o.WinePrefix()
	if prefix == "" {
		prefix = r.DefaultPrefix
	}

	if prefix == "" {
		return nil, errors.New("wine prefix is required")
	}
	env := r.buildEnv(prefix, o)
	cmd := exec.CommandContext(ctx, r.WineBin, "tasklist")
	cmd.Env = env
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("wine tasklist failed: %w: %s", err, strings.TrimSpace(stderr.String()))
	}
	process := ParseTasklist(stdout.String())
	filtered := make([]*gr.Process, 0, len(process))
	for _, p := range process {
		if o.MatchProcess(p) {
			filtered = append(filtered, p)
		}
	}

	return filtered, nil
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

func wineImageName(cmdline []string) string {
	for _, arg := range cmdline {
		if strings.HasSuffix(strings.ToLower(arg), ".exe") {
			return filepath.Base(arg)
		}
	}

	if len(cmdline) > 0 {
		return filepath.Base(cmdline[0])
	}

	return ""
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
