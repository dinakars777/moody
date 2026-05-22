package voice

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/dinakars777/moody/mood"
)

var packSlugPattern = regexp.MustCompile(`^[a-z0-9][a-z0-9_-]{0,31}$`)

var validMoodLabels = map[mood.MoodLabel]bool{
	mood.MoodHappy:      true,
	mood.MoodGrumpy:     true,
	mood.MoodAnxious:    true,
	mood.MoodDramatic:   true,
	mood.MoodDeadInside: true,
}

var validEventNames = map[string]bool{
	"ai_done":        true,
	"battery_crit":   true,
	"battery_low":    true,
	"charger_in":     true,
	"charger_out":    true,
	"display_in":     true,
	"display_out":    true,
	"headphones_in":  true,
	"headphones_out": true,
	"lid_close":      true,
	"lid_open":       true,
	"slap":           true,
	"usb_in":         true,
	"usb_out":        true,
	"wifi_back":      true,
	"wifi_lost":      true,
}

var validAudioExtensions = map[string]bool{
	".aiff": true,
	".m4a":  true,
	".mp3":  true,
	".mp4":  true,
	".wav":  true,
}

// ValidatePack checks that a directory contains a usable Moody voice pack.
func ValidatePack(packDir string) (*Manifest, string, error) {
	manifest, err := ReadManifest(packDir)
	if err != nil {
		return nil, "", err
	}

	var issues []string
	requireString := func(field, value string) {
		if strings.TrimSpace(value) == "" {
			issues = append(issues, fmt.Sprintf("manifest.%s is required", field))
		}
	}

	requireString("name", manifest.Name)
	requireString("language", manifest.Language)
	requireString("personality", manifest.Personality)
	requireString("version", manifest.Version)
	requireString("author", manifest.Author)
	requireString("description", manifest.Description)

	if manifest.Language != "" && !packSlugPattern.MatchString(manifest.Language) {
		issues = append(issues, "manifest.language must be a lowercase slug")
	}
	if manifest.Personality != "" && !packSlugPattern.MatchString(manifest.Personality) {
		issues = append(issues, "manifest.personality must be a lowercase slug")
	}

	packName := PackName(manifest)
	if packName == "_" {
		issues = append(issues, "manifest.language and manifest.personality must produce a pack name")
	}
	issues = append(issues, validateMetadata(manifest)...)

	lineCount, lineIssues := validateLines(manifest.Lines)
	issues = append(issues, lineIssues...)

	audioCount, audioIssues := validateAudio(packDir)
	issues = append(issues, audioIssues...)

	if lineCount == 0 && audioCount == 0 {
		issues = append(issues, "pack must include at least one text line or supported audio file")
	}

	if len(issues) > 0 {
		sort.Strings(issues)
		return nil, "", errors.New(strings.Join(issues, "; "))
	}
	return manifest, packName, nil
}

func validateMetadata(manifest *Manifest) []string {
	var issues []string
	for _, eventName := range manifest.SupportedEvents {
		if !validEventNames[eventName] {
			issues = append(issues, fmt.Sprintf("supportedEvents.%s uses an unknown event", eventName))
		}
	}
	for _, format := range manifest.AudioFormats {
		if !strings.HasPrefix(format, ".") {
			format = "." + format
		}
		if !validAudioExtensions[strings.ToLower(format)] {
			issues = append(issues, fmt.Sprintf("audioFormats.%s is not supported", format))
		}
	}
	for path, checksum := range manifest.Checksums {
		if filepath.IsAbs(path) || strings.Contains(path, "..") {
			issues = append(issues, fmt.Sprintf("checksums.%s must be a relative pack path", path))
		}
		if strings.TrimSpace(checksum) == "" {
			issues = append(issues, fmt.Sprintf("checksums.%s is empty", path))
		}
	}
	return issues
}

// ReadManifest reads manifest.json from a pack directory.
func ReadManifest(packDir string) (*Manifest, error) {
	manifestPath := filepath.Join(packDir, "manifest.json")
	data, err := os.ReadFile(manifestPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, fmt.Errorf("invalid pack: manifest.json not found in the repository root")
		}
		return nil, fmt.Errorf("failed to read manifest.json: %w", err)
	}

	var manifest Manifest
	if err := json.Unmarshal(data, &manifest); err != nil {
		return nil, fmt.Errorf("manifest.json is invalid JSON: %w", err)
	}
	return &manifest, nil
}

// PackName returns the local pack directory name for a manifest.
func PackName(manifest *Manifest) string {
	return strings.ToLower(strings.TrimSpace(manifest.Language) + "_" + strings.TrimSpace(manifest.Personality))
}

func validateLines(lines map[string]map[mood.MoodLabel][]string) (int, []string) {
	var issues []string
	lineCount := 0
	for eventName, moodLines := range lines {
		if !validEventNames[eventName] {
			issues = append(issues, fmt.Sprintf("lines.%s uses an unknown event", eventName))
			continue
		}
		for label, values := range moodLines {
			if !validMoodLabels[label] {
				issues = append(issues, fmt.Sprintf("lines.%s.%s uses an unknown mood", eventName, label))
				continue
			}
			for _, value := range values {
				if strings.TrimSpace(value) != "" {
					lineCount++
				}
			}
		}
	}
	return lineCount, issues
}

func validateAudio(packDir string) (int, []string) {
	audioDir := filepath.Join(packDir, "audio")
	info, err := os.Stat(audioDir)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return 0, nil
		}
		return 0, []string{fmt.Sprintf("audio directory cannot be read: %v", err)}
	}
	if !info.IsDir() {
		return 0, []string{"audio must be a directory"}
	}

	audioCount := 0
	var issues []string
	err = filepath.WalkDir(audioDir, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			issues = append(issues, fmt.Sprintf("cannot read %s: %v", path, err))
			return nil
		}
		if path == audioDir {
			return nil
		}
		if strings.HasPrefix(entry.Name(), ".") {
			if entry.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		rel, err := filepath.Rel(audioDir, path)
		if err != nil {
			issues = append(issues, fmt.Sprintf("cannot inspect %s: %v", path, err))
			return nil
		}
		parts := strings.Split(rel, string(os.PathSeparator))
		if len(parts) == 0 || !validEventNames[parts[0]] {
			issues = append(issues, fmt.Sprintf("audio/%s uses an unknown event", parts[0]))
			if entry.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		if entry.IsDir() {
			return nil
		}
		if entry.Type()&os.ModeSymlink != 0 {
			issues = append(issues, fmt.Sprintf("audio/%s is a symlink", rel))
			return nil
		}
		if validAudioExtensions[strings.ToLower(filepath.Ext(entry.Name()))] {
			audioCount++
		}
		return nil
	})
	if err != nil {
		issues = append(issues, fmt.Sprintf("audio directory cannot be walked: %v", err))
	}
	return audioCount, issues
}
