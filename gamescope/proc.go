//go:build linux

package gamescope

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/DarlingGoose/gr"
)

func listNativeProcesses(ctx context.Context, o gr.Options) ([]*gr.Process, error) {
	all, err := listProc(ctx)
	if err != nil {
		return nil, err
	}

	out := make([]*gr.Process, 0, len(all))

	for _, p := range all {
		if !o.MatchProcess(p) {
			continue
		}

		out = append(out, p)
	}

	return out, nil
}

func listProc(ctx context.Context) ([]*gr.Process, error) {
	entries, err := os.ReadDir("/proc")
	if err != nil {
		return nil, fmt.Errorf("read /proc: %w", err)
	}

	out := make([]*gr.Process, 0)

	for _, entry := range entries {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		if !entry.IsDir() {
			continue
		}

		pid, err := strconv.Atoi(entry.Name())
		if err != nil {
			continue
		}

		procDir := filepath.Join("/proc", entry.Name())

		environ, err := readProcNulFile(filepath.Join(procDir, "environ"))
		if err != nil {
			continue
		}

		cmdline, _ := readProcNulFile(filepath.Join(procDir, "cmdline"))
		status := readProcStatus(filepath.Join(procDir, "status"))

		image := imageName(cmdline)
		if image == "" {
			image = status.name
		}

		out = append(out, &gr.Process{
			ImageName: image,
			PID:       pid,
			Session:   status.name,
			SessionID: status.sessionID,
			MemUsage:  status.memUsage,
			Status:    status.status,
			Cmdline:   cmdline,
			Environ:   environ,
		})
	}

	return out, nil
}

func readProcNulFile(path string) ([]string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	data = bytes.TrimRight(data, "\x00")
	if len(data) == 0 {
		return nil, nil
	}

	parts := bytes.Split(data, []byte{0})
	out := make([]string, 0, len(parts))

	for _, part := range parts {
		if len(part) == 0 {
			continue
		}

		out = append(out, string(part))
	}

	return out, nil
}

func imageName(cmdline []string) string {
	for _, arg := range cmdline {
		lower := strings.ToLower(arg)
		if strings.HasSuffix(lower, ".exe") {
			return filepath.Base(arg)
		}
	}

	for _, arg := range cmdline {
		base := filepath.Base(arg)
		if base != "" {
			return base
		}
	}

	return ""
}

type procStatus struct {
	name      string
	sessionID string
	memUsage  string
	status    gr.Status
}

func readProcStatus(path string) procStatus {
	data, err := os.ReadFile(path)
	if err != nil {
		return procStatus{status: gr.StatusUnknown}
	}

	var s procStatus
	s.status = gr.StatusUnknown

	for _, line := range strings.Split(string(data), "\n") {
		key, value, ok := strings.Cut(line, ":")
		if !ok {
			continue
		}

		value = strings.TrimSpace(value)

		switch key {
		case "Name":
			s.name = value
		case "State":
			s.status = parseLinuxProcState(value)
		case "VmRSS":
			s.memUsage = value
		case "NSpid":
			fields := strings.Fields(value)
			if len(fields) > 0 {
				s.sessionID = fields[len(fields)-1]
			}
		}
	}

	return s
}

func parseLinuxProcState(v string) gr.Status {
	switch {
	case strings.HasPrefix(v, "R"),
		strings.HasPrefix(v, "S"),
		strings.HasPrefix(v, "D"),
		strings.HasPrefix(v, "I"):
		return gr.StatusRunning

	case strings.HasPrefix(v, "T"),
		strings.HasPrefix(v, "t"):
		return gr.StatusStopped

	case strings.HasPrefix(v, "Z"),
		strings.HasPrefix(v, "X"):
		return gr.StatusExited

	default:
		return gr.StatusUnknown
	}
}
