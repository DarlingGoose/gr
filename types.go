package gr

import (
	"context"
	"os/exec"
)

type Status int

const (
	StatusUnknown Status = iota
	StatusRunning
	StatusExited
	StatusStopped
)

func (s Status) String() string {
	switch s {
	case StatusRunning:
		return "running"
	case StatusExited:
		return "exited"
	case StatusStopped:
		return "stopped"
	default:
		return "unknown"
	}
}

type Runner interface {
	Run(ctx context.Context, cmd string, opts ...Option) (*Process, error)
	List(ctx context.Context, opts ...Option) ([]*Process, error)
	Find(ctx context.Context, opts ...Option) (*Process, error)
	GetOption(key string) (interface{}, error)
	GetOptionKeys() ([]string, error)
}

type Options struct {
	background bool
	workingDir string
	args       []string
	envs       []string

	// For Wine.
	systemArch string // "32", "64", "win32", "win64"
	wineprefix string

	// Dependencies to install before running, usually via winetricks.
	dependencies []string

	// Search/filter options.
	name      string
	pid       int
	session   string
	sessionID string
	logFile   string
}

type Process struct {
	ImageName string
	PID       int
	WinePID   int
	Session   string
	SessionID string
	MemUsage  string
	Status    Status

	Cmdline []string
	Environ []string

	Cmd *exec.Cmd
}
