package wine

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"syscall"
	"time"

	"github.com/DarlingGoose/gr"
)

func List(ctx context.Context, wineBin string, env []string, opts ...gr.Option) ([]*gr.Process, error) {
	o := gr.ApplyOptions(opts...)

	prefix := o.WinePrefix()

	if prefix == "" {
		return nil, errors.New("wine prefix is required for wine process listing")
	}

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, wineBin, "tasklist")
	cmd.Env = env

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	// Important: make Win###e + anything it spawns killable as a group.
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setpgid: true,
	}

	cmd.Cancel = func() error {
		if cmd.Process == nil {
			return os.ErrProcessDone
		}

		// Kill the whole process group, not just the wine wrapper.
		if err := syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL); err != nil {
			if errors.Is(err, syscall.ESRCH) {
				return os.ErrProcessDone
			}
			return err
		}

		return nil
	}

	// Prevent Run/Wait from hanging forever if descendants keep pipes open.
	cmd.WaitDelay = 10 * time.Second

	err := cmd.Run()
	if err != nil {
		if errors.Is(ctx.Err(), context.DeadlineExceeded) {
			return nil, fmt.Errorf("wine tasklist timed out for prefix %q", prefix)
		}

		return nil, fmt.Errorf(
			"wine tasklist failed: %w: %s",
			err,
			strings.TrimSpace(stderr.String()),
		)
	}

	processes := ParseTasklist(stdout.String())

	filtered := make([]*gr.Process, 0, len(processes))
	for _, p := range processes {
		if o.MatchProcess(p) {
			p.WinePID = p.PID
			filtered = append(filtered, p)
		}
	}

	return filtered, nil
}
