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

	score, err := scoreExe(absPath)
	if err != nil {
		return Detection{}, err
	}

	kind := KindGame
	reasons := score.gameReasons

	// Default to game unless installer evidence is strong enough.
	//
	// This avoids false positives from VN/game EXEs that contain generic
	// "install", "setup", "InstallShield", or "NSIS" strings in embedded
	// resources/runtime code.
	if score.installerScore >= minInstallerScore &&
		score.installerScore > score.gameScore+installerMargin {
		kind = KindInstaller
		reasons = score.installerReasons
	}

	if len(reasons) == 0 {
		reasons = append(reasons, fmt.Sprintf(
			"installer score=%d game score=%d; defaulting to game",
			score.installerScore,
			score.gameScore,
		))
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
	base := strings.ToLower(filepath.Base(path))

	var reasons []string

	// Preserve original behavior:
	// do not classify uninstall.exe as an installer based on filename.
	if !strings.Contains(base, "uninstall") && !strings.Contains(base, "unins") {
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
	data, err := readScanBytes(path)
	if err != nil {
		return nil, err
	}

	data = bytes.ToLower(data)

	var reasons []string

	// Preserve original behavior:
	// this function reports content marker hits, it does not decide final kind.
	for _, marker := range contentInstallerMarkers {
		if bytes.Contains(data, []byte(marker)) {
			reasons = append(reasons, "content contains "+marker)
		}
	}

	return reasons, nil
}

const maxInstallerScanBytes = 16 << 20

const (
	minInstallerScore = 4
	installerMargin   = 1
)

type exeScore struct {
	installerScore   int
	installerReasons []string

	gameScore   int
	gameReasons []string
}

func scoreExe(path string) (exeScore, error) {
	var score exeScore

	base := strings.ToLower(filepath.Base(path))
	scoreFilename(base, &score)

	data, err := readScanBytes(path)
	if err != nil {
		return exeScore{}, err
	}

	lower := bytes.ToLower(data)

	scoreInstallerContent(lower, &score)
	scoreGameContent(lower, &score)

	return score, nil
}

func readScanBytes(path string) ([]byte, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open exe for scan: %w", err)
	}
	defer f.Close()

	data, err := io.ReadAll(io.LimitReader(f, maxInstallerScanBytes))
	if err != nil {
		return nil, fmt.Errorf("scan exe markers: %w", err)
	}

	return data, nil
}

func scoreFilename(base string, score *exeScore) {
	if strings.Contains(base, "uninstall") || strings.Contains(base, "unins") {
		score.installerScore += 5
		score.installerReasons = append(score.installerReasons, "filename looks like uninstaller")
		return
	}

	for _, marker := range filenameInstallerMarkers {
		if strings.Contains(base, marker) {
			score.installerScore += 4
			score.installerReasons = append(score.installerReasons, "filename contains "+marker)
			break
		}
	}

	for _, marker := range filenameGameMarkers {
		if strings.Contains(base, marker) {
			score.gameScore += 1
			score.gameReasons = append(score.gameReasons, "filename contains game marker "+marker)
			break
		}
	}
}

func scoreInstallerContent(data []byte, score *exeScore) {
	for _, marker := range strongInstallerContentMarkers {
		if bytes.Contains(data, []byte(marker)) {
			score.installerScore += 4
			score.installerReasons = append(score.installerReasons, "content strongly contains "+marker)
		}
	}

	for _, marker := range mediumInstallerContentMarkers {
		if bytes.Contains(data, []byte(marker)) {
			score.installerScore += 2
			score.installerReasons = append(score.installerReasons, "content contains "+marker)
		}
	}

	for _, marker := range weakInstallerContentMarkers {
		if bytes.Contains(data, []byte(marker)) {
			score.installerScore += 1
			score.installerReasons = append(score.installerReasons, "content weakly contains "+marker)
		}
	}
}

func scoreGameContent(data []byte, score *exeScore) {
	for _, marker := range strongGameContentMarkers {
		if bytes.Contains(data, []byte(marker)) {
			score.gameScore += 4
			score.gameReasons = append(score.gameReasons, "content strongly contains game marker "+marker)
		}
	}

	for _, marker := range mediumGameContentMarkers {
		if bytes.Contains(data, []byte(marker)) {
			score.gameScore += 2
			score.gameReasons = append(score.gameReasons, "content contains game marker "+marker)
		}
	}

	for _, marker := range weakGameContentMarkers {
		if bytes.Contains(data, []byte(marker)) {
			score.gameScore += 1
			score.gameReasons = append(score.gameReasons, "content weakly contains game marker "+marker)
		}
	}
}

var filenameInstallerMarkers = []string{
	"setup",
	"installer",
	"install",
	"bootstrapper",
	"redist",
	"redistributable",
}

var filenameGameMarkers = []string{
	"game",
	"player",
	"launcher",
	"start",
	"play",
}

// Kept for compatibility with your original file.
// The new detector does not treat all of these as equal.
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

var strongInstallerContentMarkers = []string{
	// Inno Setup
	"inno setup setup data",
	"jrsoftware\\inno setup",
	"setup.e32",
	"setup-0.bin",

	// NSIS
	"nullsoft install system",
	"nsis error",
	"nsisdl.dll",
	"$pluginsdir",

	// WiX Burn
	"wix burn",
	"burn bootstrapper",
	"burnengine",
	"bundle.wixburn",

	// InstallShield / Wise
	"installshield wizard",
	"installshield setup launcher",
	"wise installation wizard",
}

var mediumInstallerContentMarkers = []string{
	"inno setup",
	"installshield",
	"setup wizard",
	"wise installation",
	"installation wizard",
	"uninstall wizard",
	"extracting files",
	"select destination location",
	"choose install location",
	"license agreement",
	"i accept the agreement",
}

var weakInstallerContentMarkers = []string{
	// These commonly appear in real games too, so they are intentionally weak.
	"nsis",
	"install",
	"installer",
	"uninstall",
	"uninstaller",
	"setup",
	"redist",
	"redistributable",
}

var strongGameContentMarkers = []string{
	// Kirikiri / KAG / VN engines
	"kirikiri",
	"kirikiri2",
	"kirikiri z",
	"tvp(kirikiri)",
	"startup.tjs",
	"system.tjs",
	"config.tjs",
	"data.xp3",
	".xp3",
	"kag",

	// RPG Maker
	"rpg maker",
	"rgss",
	"rgss102e.dll",
	"rgss202e.dll",
	"rgss301.dll",
	"game.ini",
	"data\\actors.rvdata",
	"data\\actors.rvdata2",
	"data\\system.rvdata",
	"data\\system.rvdata2",
}

var mediumGameContentMarkers = []string{
	// Common game/runtime markers
	"direct3d",
	"directdraw",
	"directsound",
	"xinput",
	"dinput",
	"dsound.dll",
	"d3d9.dll",
	"d3dx9_",
	"steam_api.dll",
	"steam_api64.dll",

	// VN-ish / script-ish
	".ks",
	".tjs",
	"scenario",
	"savedata",
	"save data",
	"bgm",
	"voice",
	"se",
	"cg",
}

var weakGameContentMarkers = []string{
	"fullscreen",
	"windowed",
	"resolution",
	"joystick",
	"gamepad",
	"new game",
	"continue",
	"load game",
	"config",
	"options",
	"volume",
	"font",
}
