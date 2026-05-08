package gr

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"unicode"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

var (
	versionRegex = regexp.MustCompile(`(([vV](er){0,1}(sion){0,1})[0-9\.\-]+)`)
	sepRegex     = regexp.MustCompile(`[_\-]`)
	titleCaser   = cases.Title(language.English)
)

func FirstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}

func DeriveGameName(inputPath, executablePath string, inputWasDir bool) string {
	var name string
	if inputWasDir {
		name = filepath.Base(inputPath)
	} else {
		name = strings.TrimSuffix(filepath.Base(executablePath), filepath.Ext(executablePath))
	}

	// Clean name
	name = versionRegex.ReplaceAllString(name, "")
	name = sepRegex.ReplaceAllString(name, " ")
	name = strings.TrimSpace(name)

	// Title-case only if it looks like English
	if isMostlyASCII(name) {
		name = titleCaser.String(strings.ToLower(name))
	}

	return name
}

func SanitizeName(name string) string {
	name = strings.TrimSpace(strings.ToLower(name))
	replacer := strings.NewReplacer(" ", "-", "/", "-", "\\", "-", "_", "-")
	name = replacer.Replace(name)
	var builder strings.Builder
	for _, r := range name {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' {
			builder.WriteRune(r)
		}
	}
	sanitized := strings.Trim(builder.String(), "-")
	if sanitized == "" {
		return "game"
	}
	return sanitized
}

func isMostlyASCII(s string) bool {
	if s == "" {
		return false
	}

	var asciiCount, total int
	for _, r := range s {
		if unicode.IsLetter(r) {
			total++
			if r <= unicode.MaxASCII {
				asciiCount++
			}
		}
	}

	// avoid division by zero
	if total == 0 {
		return false
	}

	// threshold: 80% ASCII letters
	return float64(asciiCount)/float64(total) > 0.8
}

func GetWinePrefix(programName, game string) (string, error) {
	if game == "" {
		return "", errors.New("missing game name")
	}
	game = SanitizeName(game)
	userDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	if programName == "" {
		return filepath.Join(userDir, ".local", game), nil
	}
	return filepath.Join(userDir, ".local", programName, game), nil

}

func GetOption(opt interface{}, key string) (interface{}, error) {
	d, err := json.Marshal(opt)
	if err != nil {
		return nil, err
	}
	data := map[string]interface{}{}
	err = json.Unmarshal(d, &data)
	if err != nil {
		return nil, err
	}
	v, ok := data[key]
	if !ok {
		return nil, fmt.Errorf("key does not exist: %s", key)
	}

	return v, nil
}

func GetOptionKeys(opt interface{}) ([]string, error) {
	d, err := json.Marshal(opt)
	if err != nil {
		return nil, err
	}
	data := map[string]interface{}{}
	err = json.Unmarshal(d, &data)
	if err != nil {
		return nil, err
	}
	var keys []string
	for k := range data {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys, nil
}

func FindExe(args ...string) string {
	for _, a := range args {
		if filepath.Ext(filepath.Base(a)) == ".exe" {
			return a
		}
	}
	return ""

}
