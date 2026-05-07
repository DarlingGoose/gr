package monitors

import (
	"os"
	"strings"
)

type SessionType string

const (
	SessionWayland SessionType = "wayland"
	SessionX11     SessionType = "x11"
	SessionUnknown SessionType = "unknown"
)

func CurrentSessionType() SessionType {
	// Most reliable when set by systemd/logind/display managers.
	if v := strings.ToLower(strings.TrimSpace(os.Getenv("XDG_SESSION_TYPE"))); v != "" {
		switch v {
		case "wayland":
			return SessionWayland
		case "x11":
			return SessionX11
		}
	}

	// Fallbacks for apps launched inside the session.
	if os.Getenv("WAYLAND_DISPLAY") != "" {
		return SessionWayland
	}

	if os.Getenv("DISPLAY") != "" {
		return SessionX11
	}

	return SessionUnknown
}

func IsWayland() bool {
	return CurrentSessionType() == SessionWayland
}

func IsX11() bool {
	return CurrentSessionType() == SessionX11
}
