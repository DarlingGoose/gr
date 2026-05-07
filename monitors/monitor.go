package monitors

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

type Mode struct {
	Width  int
	Height int
	Raw    string // original string, e.g. "2560x1440"
}

func (m Mode) String() string {
	if m.Width <= 0 || m.Height <= 0 {
		return m.Raw
	}
	return fmt.Sprintf("%dx%d", m.Width, m.Height)
}

type Monitor struct {
	Name        string
	CardPath    string
	Connected   bool
	Modes       []Mode
	CurrentMode Mode
}

func GetMonitors() ([]Monitor, error) {
	paths, err := filepath.Glob("/sys/class/drm/card*-*")
	if err != nil {
		return nil, err
	}

	var monitors []Monitor

	for _, p := range paths {
		name := strings.TrimPrefix(filepath.Base(p), "card")
		if i := strings.Index(name, "-"); i >= 0 {
			name = name[i+1:]
		}

		status, _ := readTrim(filepath.Join(p, "status"))
		modeLines, _ := readLines(filepath.Join(p, "modes"))

		modes := uniqueModes(modeLines)

		mon := Monitor{
			Name:      name,
			CardPath:  p,
			Connected: status == "connected",
			Modes:     modes,
		}

		if len(modes) > 0 {
			mon.CurrentMode = modes[0]
		}

		monitors = append(monitors, mon)
	}

	sort.Slice(monitors, func(i, j int) bool {
		return monitors[i].Name < monitors[j].Name
	})

	return monitors, nil
}

func ParseMode(s string) (Mode, error) {
	raw := strings.TrimSpace(s)
	parts := strings.Split(raw, "x")
	if len(parts) != 2 {
		return Mode{Raw: raw}, fmt.Errorf("invalid mode %q", raw)
	}

	width, err := strconv.Atoi(parts[0])
	if err != nil {
		return Mode{Raw: raw}, fmt.Errorf("invalid mode width %q: %w", raw, err)
	}

	height, err := strconv.Atoi(parts[1])
	if err != nil {
		return Mode{Raw: raw}, fmt.Errorf("invalid mode height %q: %w", raw, err)
	}

	return Mode{
		Width:  width,
		Height: height,
		Raw:    raw,
	}, nil
}

func readTrim(path string) (string, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(b)), nil
}

func readLines(path string) ([]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var out []string
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" {
			out = append(out, line)
		}
	}

	return out, scanner.Err()
}

func uniqueModes(lines []string) []Mode {
	seen := make(map[string]struct{}, len(lines))
	out := make([]Mode, 0, len(lines))

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		mode, err := ParseMode(line)
		if err != nil {
			continue
		}

		key := mode.String()
		if _, ok := seen[key]; ok {
			continue
		}

		seen[key] = struct{}{}
		out = append(out, mode)
	}

	return out
}
