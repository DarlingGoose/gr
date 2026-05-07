package autorunner

import (
	"debug/pe"
	"errors"
	"fmt"
	"strings"
)

type FileArch string

const (
	ArchUnknown FileArch = ""
	ArchWin32   FileArch = "win32"
	ArchWin64   FileArch = "win64"
)

var ErrUnknownArch = errors.New("unknown executable architecture")

func DetectFileArch(path string) (FileArch, error) {
	f, err := pe.Open(path)
	if err != nil {
		return ArchUnknown, fmt.Errorf("open PE file: %w", err)
	}
	defer f.Close()

	switch f.FileHeader.Machine {
	case pe.IMAGE_FILE_MACHINE_I386:
		return ArchWin32, nil
	case pe.IMAGE_FILE_MACHINE_AMD64, pe.IMAGE_FILE_MACHINE_ARM64:
		return ArchWin64, nil
	default:
		return ArchUnknown, fmt.Errorf("%w: machine 0x%x", ErrUnknownArch, f.FileHeader.Machine)
	}
}

func NormalizeArch(arch string) FileArch {
	switch strings.ToLower(strings.TrimSpace(arch)) {
	case "32", "x86", "i386", "386", "win32":
		return ArchWin32
	case "64", "x64", "amd64", "x86_64", "arm64", "aarch64", "win64":
		return ArchWin64
	default:
		return ArchUnknown
	}
}

func (a FileArch) WineArch() string {
	switch a {
	case ArchWin32:
		return "win32"
	case ArchWin64:
		return "win64"
	default:
		return ""
	}
}
