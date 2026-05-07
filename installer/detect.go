package installer

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/DarlingGoose/gr/autorunner"
)

type ExeKind string

const (
	KindUnknown   ExeKind = ""
	KindGame      ExeKind = "game"
	KindInstaller ExeKind = "installer"
)

var ErrNotInstaller = errors.New("exe does not look like an installer")

type Detection struct {
	Path    string
	Arch    autorunner.FileArch
	Kind    ExeKind
	Reasons []string
}

func Detect(path string) (Detection, error) {
	if path == "" {
		return Detection{}, fmt.Errorf("exe path is required")
	}

	absPath, err := filepath.Abs(path)
	if err != nil {
		return Detection{}, fmt.Errorf("resolve exe path: %w", err)
	}

	arch, err := autorunner.DetectFileArch(absPath)
	if err != nil {
		return Detection{}, err
	}

	reasons, err := installerReasons(absPath)
	if err != nil {
		return Detection{}, err
	}

	kind := KindGame
	if len(reasons) > 0 {
		kind = KindInstaller
	}

	return Detection{
		Path:    absPath,
		Arch:    arch,
		Kind:    kind,
		Reasons: reasons,
	}, nil
}

func IsInstaller(path string) (bool, error) {
	detection, err := Detect(path)
	if err != nil {
		return false, err
	}

	return detection.Kind == KindInstaller, nil
}

func VerifyInstaller(path string) (Detection, error) {
	detection, err := Detect(path)
	if err != nil {
		return Detection{}, err
	}

	if detection.Kind != KindInstaller {
		return detection, ErrNotInstaller
	}

	return detection, nil
}

func installerReasons(path string) ([]string, error) {
	var reasons []string

	base := strings.ToLower(filepath.Base(path))
	if !strings.Contains(base, "uninstall") {
		for _, marker := range filenameInstallerMarkers {
			if strings.Contains(base, marker) {
				reasons = append(reasons, "filename contains "+marker)
				break
			}
		}
	}

	contentReasons, err := installerContentReasons(path)
	if err != nil {
		return nil, err
	}

	reasons = append(reasons, contentReasons...)
	return reasons, nil
}

func installerContentReasons(path string) ([]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open exe for installer scan: %w", err)
	}
	defer f.Close()

	data, err := io.ReadAll(io.LimitReader(f, maxInstallerScanBytes))
	if err != nil {
		return nil, fmt.Errorf("scan exe installer markers: %w", err)
	}

	data = bytes.ToLower(data)
	var reasons []string
	for _, marker := range contentInstallerMarkers {
		if bytes.Contains(data, []byte(marker)) {
			reasons = append(reasons, "content contains "+marker)
		}
	}

	return reasons, nil
}

const maxInstallerScanBytes = 16 << 20

var filenameInstallerMarkers = []string{
	"setup",
	"installer",
	"install",
	"bootstrapper",
	"redist",
	"redistributable",
}

var contentInstallerMarkers = []string{
	"inno setup",
	"installshield",
	"nullsoft install system",
	"nsis",
	"wix burn",
	"burn bootstrapper",
	"setup wizard",
	"wise installation",
}
