package installer

import (
	"archive/zip"
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"regexp"
	"strings"
)

type ArchiveKind string

const (
	ArchiveUnknown          ArchiveKind = ""
	ArchiveRARSFX           ArchiveKind = "rar-sfx"
	ArchiveInstallShieldSFX ArchiveKind = "installshield-sfx"
	ArchiveZip              ArchiveKind = "zip"
)

const (
	DefaultUnrarBin    = "unrar"
	DefaultSevenZipBin = "7z"
)

var ErrNotArchive = errors.New("path does not look like a supported archive")

type ArchiveDetection struct {
	Path        string
	Kind        ArchiveKind
	FirstVolume bool
	Reasons     []string
}

type ArchiveExtraction struct {
	Detection ArchiveDetection
	DestDir   string
}

type ExtractConfig struct {
	DestDir     string
	UnrarBin    string
	SevenZipBin string
}

func DetectArchive(path string) (ArchiveDetection, error) {
	if path == "" {
		return ArchiveDetection{}, fmt.Errorf("archive path is required")
	}

	absPath, err := filepath.Abs(path)
	if err != nil {
		return ArchiveDetection{}, fmt.Errorf("resolve archive path: %w", err)
	}

	data, err := readArchiveScanBytes(absPath)
	if err != nil {
		return ArchiveDetection{}, err
	}

	lowerBase := strings.ToLower(filepath.Base(absPath))
	var reasons []string

	hasPEHeader := len(data) >= 2 && bytes.Equal(data[:2], []byte("MZ"))
	hasZIP := bytes.HasPrefix(data, []byte("PK\x03\x04")) ||
		bytes.HasPrefix(data, []byte("PK\x05\x06")) ||
		bytes.HasPrefix(data, []byte("PK\x07\x08")) ||
		strings.HasSuffix(lowerBase, ".zip")
	if hasZIP {
		reasons = append(reasons, "file looks like zip archive")
		return ArchiveDetection{
			Path:    absPath,
			Kind:    ArchiveZip,
			Reasons: reasons,
		}, nil
	}

	hasRAR := bytes.Contains(data, []byte("Rar!\x1a\x07\x00")) ||
		bytes.Contains(data, []byte("Rar!\x1a\x07\x01\x00"))
	if hasPEHeader && hasRAR {
		reasons = append(reasons, "pe executable contains rar signature")
	}

	firstVolume := isRARFirstVolumeName(lowerBase)
	if firstVolume {
		reasons = append(reasons, "filename looks like first rar volume")
	}

	if hasPEHeader && hasRAR {
		return ArchiveDetection{
			Path:        absPath,
			Kind:        ArchiveRARSFX,
			FirstVolume: firstVolume,
			Reasons:     reasons,
		}, nil
	}

	hasInstallShield := bytes.Contains(bytes.ToLower(data), []byte("installshield"))
	if hasPEHeader && hasInstallShield {
		reasons = append(reasons, "pe executable contains installshield marker")
		return ArchiveDetection{
			Path:    absPath,
			Kind:    ArchiveInstallShieldSFX,
			Reasons: reasons,
		}, nil
	}

	return ArchiveDetection{
		Path:        absPath,
		Kind:        ArchiveUnknown,
		FirstVolume: firstVolume,
		Reasons:     reasons,
	}, ErrNotArchive
}

func ExtractArchive(ctx context.Context, path string, cfg ExtractConfig) (ArchiveExtraction, error) {
	detection, err := DetectArchive(path)
	if err != nil {
		return ArchiveExtraction{}, err
	}

	destDir := cfg.DestDir
	if destDir == "" {
		destDir = defaultExtractDir(detection.Path)
	}
	absDest, err := filepath.Abs(destDir)
	if err != nil {
		return ArchiveExtraction{}, fmt.Errorf("resolve extract dir: %w", err)
	}
	if err := os.MkdirAll(absDest, 0o755); err != nil {
		return ArchiveExtraction{}, fmt.Errorf("create extract dir: %w", err)
	}

	switch detection.Kind {
	case ArchiveRARSFX:
		if err := extractRAR(ctx, detection.Path, absDest, cfg.UnrarBin); err != nil {
			return ArchiveExtraction{}, err
		}
	case ArchiveInstallShieldSFX:
		if err := extractWith7z(ctx, detection.Path, absDest, cfg.SevenZipBin); err != nil {
			return ArchiveExtraction{}, err
		}
	case ArchiveZip:
		if err := extractZip(detection.Path, absDest); err != nil {
			return ArchiveExtraction{}, err
		}
	default:
		return ArchiveExtraction{}, ErrNotArchive
	}

	return ArchiveExtraction{
		Detection: detection,
		DestDir:   absDest,
	}, nil
}

func extractRAR(ctx context.Context, path, destDir, unrarBin string) error {
	if unrarBin == "" {
		unrarBin = DefaultUnrarBin
	}

	cmd := exec.CommandContext(ctx, unrarBin, "x", "-o+", path, destDir)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("extract archive with %s: %w: %s", unrarBin, err, strings.TrimSpace(string(output)))
	}

	return nil
}

func extractWith7z(ctx context.Context, path, destDir, sevenZipBin string) error {
	if sevenZipBin == "" {
		sevenZipBin = DefaultSevenZipBin
	}

	cmd := exec.CommandContext(ctx, sevenZipBin, "x", "-y", "-o"+destDir, path)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("extract archive with %s: %w: %s", sevenZipBin, err, strings.TrimSpace(string(output)))
	}

	return nil
}

func extractZip(path, destDir string) error {
	reader, err := zip.OpenReader(path)
	if err != nil {
		return fmt.Errorf("open zip archive: %w", err)
	}
	defer reader.Close()

	for _, file := range reader.File {
		if err := extractZipFile(file, destDir); err != nil {
			return err
		}
	}

	return nil
}

func extractZipFile(file *zip.File, destDir string) error {
	cleanName := path.Clean(strings.ReplaceAll(file.Name, "\\", "/"))
	if cleanName == "." || path.IsAbs(cleanName) || strings.HasPrefix(cleanName, "../") || cleanName == ".." {
		return fmt.Errorf("zip entry escapes destination: %s", file.Name)
	}

	destPath := filepath.Join(append([]string{destDir}, strings.Split(cleanName, "/")...)...)
	if file.FileInfo().IsDir() {
		if err := os.MkdirAll(destPath, file.Mode()); err != nil {
			return fmt.Errorf("create zip directory: %w", err)
		}
		return nil
	}

	if err := os.MkdirAll(filepath.Dir(destPath), 0o755); err != nil {
		return fmt.Errorf("create zip parent directory: %w", err)
	}

	src, err := file.Open()
	if err != nil {
		return fmt.Errorf("open zip entry: %w", err)
	}
	defer src.Close()

	dst, err := os.OpenFile(destPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, file.Mode())
	if err != nil {
		return fmt.Errorf("create zip entry: %w", err)
	}
	defer dst.Close()

	if _, err := io.Copy(dst, src); err != nil {
		return fmt.Errorf("extract zip entry: %w", err)
	}

	return nil
}

func looksLikeSupportedArchive(path string) bool {
	if path == "" {
		return false
	}
	detection, err := DetectArchive(path)
	return err == nil && detection.Kind != ArchiveUnknown
}

const maxArchiveScanBytes = 16 << 20

var rarPartRE = regexp.MustCompile(`(?i)\.part0*1\.exe$`)

func readArchiveScanBytes(path string) ([]byte, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open archive for scan: %w", err)
	}
	defer f.Close()

	data, err := io.ReadAll(io.LimitReader(f, maxArchiveScanBytes))
	if err != nil {
		return nil, fmt.Errorf("scan archive markers: %w", err)
	}

	return data, nil
}

func isRARFirstVolumeName(base string) bool {
	return rarPartRE.MatchString(base)
}

func defaultExtractDir(path string) string {
	dir := filepath.Dir(path)
	base := filepath.Base(path)
	base = rarPartRE.ReplaceAllString(base, "")
	if base == filepath.Base(path) {
		base = strings.TrimSuffix(base, filepath.Ext(base))
	}
	if base == "" || base == "." {
		base = "extracted"
	}
	return filepath.Join(dir, base)
}
