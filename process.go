package gr

import (
	"path/filepath"
	"strconv"
	"strings"
)

func (o Options) Background() bool {
	return o.background
}

func (o Options) WorkingDir() string {
	return o.workingDir
}

func (o Options) Args() []string {
	return append([]string(nil), o.args...)
}

func (o Options) Envs() []string {
	return append([]string(nil), o.envs...)
}

func (o Options) WinePrefix() string {
	return o.wineprefix
}

func (o Options) SystemArch() string {
	return o.systemArch
}

func (o Options) Dependencies() []string {
	return append([]string(nil), o.dependencies...)
}

func (o Options) Name() string {
	return o.name
}

func (o Options) PID() int {
	return o.pid
}

func (o Options) Session() string {
	return o.session
}

func (o Options) SessionID() string {
	return o.sessionID
}

func (o Options) MatchProcess(p *Process) bool {
	if p == nil {
		return false
	}

	//if o.pid > 0 && p.PID != o.pid {
	//	return false
	//}

	if o.name != "" {
		want := strings.ToLower(o.name)
		image := strings.ToLower(p.ImageName)
		base := strings.ToLower(filepath.Base(p.ImageName))

		matched := image == want ||
			base == want ||
			strings.Contains(image, want) ||
			strings.Contains(base, want)

		if !matched {
			for _, arg := range p.Cmdline {
				if strings.Contains(strings.ToLower(arg), want) {
					matched = true
					break
				}
			}
		}

		if !matched {
			return false
		}
	}

	if o.session != "" && p.Session != o.session {
		return false
	}

	if o.sessionID != "" && p.SessionID != o.sessionID {
		return false
	}

	return true
}

func ParsePID(s string) int {
	pid, _ := strconv.Atoi(strings.TrimSpace(s))
	return pid
}
