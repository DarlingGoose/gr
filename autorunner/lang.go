package autorunner

import (
	"bytes"
	"debug/pe"
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

func DetectWineLang(path string) (string, error) {
	langs, err := peResourceLanguageIDs(path)
	if err != nil {
		return "", err
	}

	for _, langID := range langs {
		if lang := wineLangForWindowsLangID(langID); lang != "" {
			return lang, nil
		}
	}

	return detectWineLangFromMarkers(path)
}

func peResourceLanguageIDs(path string) ([]uint16, error) {
	f, err := pe.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open PE file: %w", err)
	}
	defer f.Close()

	var rsrc *pe.Section
	for _, section := range f.Sections {
		if section.Name == ".rsrc" {
			rsrc = section
			break
		}
	}

	if rsrc == nil {
		return nil, nil
	}

	data, err := rsrc.Data()
	if err != nil {
		return nil, fmt.Errorf("read PE resources: %w", err)
	}

	var versionLangs []uint16
	var allLangs []uint16
	walkResourceDirectory(data, 0, 0, 0, &versionLangs, &allLangs)

	if len(versionLangs) > 0 {
		return uniqueLangIDs(versionLangs), nil
	}

	return uniqueLangIDs(allLangs), nil
}

func walkResourceDirectory(data []byte, offset, level int, resourceType uint32, versionLangs, allLangs *[]uint16) {
	if level > 3 || offset < 0 || offset+16 > len(data) {
		return
	}

	named := int(binary.LittleEndian.Uint16(data[offset+12:]))
	ids := int(binary.LittleEndian.Uint16(data[offset+14:]))
	entryCount := named + ids
	entriesOffset := offset + 16

	if entryCount < 0 || entriesOffset+(entryCount*8) > len(data) {
		return
	}

	for i := 0; i < entryCount; i++ {
		entryOffset := entriesOffset + i*8
		nameRaw := binary.LittleEndian.Uint32(data[entryOffset:])
		dataRaw := binary.LittleEndian.Uint32(data[entryOffset+4:])
		isNameString := nameRaw&0x80000000 != 0
		isDirectory := dataRaw&0x80000000 != 0
		id := nameRaw & 0x7fffffff

		nextType := resourceType
		if level == 0 && !isNameString {
			nextType = id
		}

		if isDirectory {
			nextOffset := int(dataRaw & 0x7fffffff)
			walkResourceDirectory(data, nextOffset, level+1, nextType, versionLangs, allLangs)
			continue
		}

		if level >= 2 && !isNameString {
			langID := uint16(id)
			*allLangs = append(*allLangs, langID)
			if resourceType == peResourceTypeVersion {
				*versionLangs = append(*versionLangs, langID)
			}
		}
	}
}

func uniqueLangIDs(ids []uint16) []uint16 {
	seen := make(map[uint16]bool, len(ids))
	out := make([]uint16, 0, len(ids))
	for _, id := range ids {
		if seen[id] {
			continue
		}
		seen[id] = true
		out = append(out, id)
	}
	return out
}

func wineLangForWindowsLangID(langID uint16) string {
	switch langID & 0x03ff {
	case 0x04:
		switch langID {
		case 0x0404:
			return "zh_TW.UTF-8"
		default:
			return "zh_CN.UTF-8"
		}
	case 0x11:
		return "ja_JP.UTF-8"
	case 0x12:
		return "ko_KR.UTF-8"
	case 0x19:
		return "ru_RU.UTF-8"
	case 0x07:
		return "de_DE.UTF-8"
	case 0x0c:
		return "fr_FR.UTF-8"
	case 0x10:
		return "it_IT.UTF-8"
	case 0x0a:
		switch langID {
		case 0x080a, 0x0c0a, 0x100a, 0x140a, 0x180a, 0x1c0a, 0x200a, 0x240a, 0x280a, 0x2c0a, 0x300a, 0x340a, 0x380a, 0x3c0a, 0x400a, 0x440a, 0x480a, 0x4c0a, 0x500a:
			return "es_ES.UTF-8"
		default:
			return "es_MX.UTF-8"
		}
	default:
		return ""
	}
}

func detectWineLangFromMarkers(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", fmt.Errorf("open file for locale scan: %w", err)
	}
	defer f.Close()

	data, err := io.ReadAll(io.LimitReader(f, maxLangScanBytes))
	if err != nil {
		return "", fmt.Errorf("scan locale markers: %w", err)
	}

	lowerName := strings.ToLower(filepath.Base(path))
	lowerData := bytes.ToLower(data)

	for _, marker := range wineLangMarkers {
		if strings.Contains(lowerName, marker.marker) || bytes.Contains(lowerData, []byte(marker.marker)) {
			return marker.lang, nil
		}
	}

	return "", nil
}

const (
	peResourceTypeVersion = 16
	maxLangScanBytes      = 16 << 20
)

var wineLangMarkers = []struct {
	marker string
	lang   string
}{
	{marker: "ja_jp", lang: "ja_JP.UTF-8"},
	{marker: "japanese", lang: "ja_JP.UTF-8"},
	{marker: "nihongo", lang: "ja_JP.UTF-8"},
	{marker: "ko_kr", lang: "ko_KR.UTF-8"},
	{marker: "korean", lang: "ko_KR.UTF-8"},
	{marker: "zh_cn", lang: "zh_CN.UTF-8"},
	{marker: "simplified chinese", lang: "zh_CN.UTF-8"},
	{marker: "zh_tw", lang: "zh_TW.UTF-8"},
	{marker: "traditional chinese", lang: "zh_TW.UTF-8"},
}
