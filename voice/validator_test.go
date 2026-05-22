package voice

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestValidatePackAcceptsTextLines(t *testing.T) {
	packDir := writePack(t, `{
		"name": "English Test",
		"language": "en",
		"personality": "test",
		"version": "1.0.0",
		"author": "tester",
		"nsfw": false,
		"description": "Test pack",
		"lines": {
			"slap": {
				"happy": ["Careful."]
			}
		}
	}`)

	manifest, packName, err := ValidatePack(packDir)
	if err != nil {
		t.Fatalf("ValidatePack() error = %v", err)
	}
	if packName != "en_test" {
		t.Fatalf("packName = %q, want %q", packName, "en_test")
	}
	if manifest.Name != "English Test" {
		t.Fatalf("manifest name = %q", manifest.Name)
	}
}

func TestValidatePackRejectsMissingContent(t *testing.T) {
	packDir := writePack(t, `{
		"name": "English Empty",
		"language": "en",
		"personality": "empty",
		"version": "1.0.0",
		"author": "tester",
		"nsfw": false,
		"description": "Empty pack"
	}`)

	_, _, err := ValidatePack(packDir)
	if err == nil {
		t.Fatal("ValidatePack() error = nil, want content error")
	}
	if !strings.Contains(err.Error(), "at least one text line or supported audio file") {
		t.Fatalf("error = %q", err)
	}
}

func TestValidatePackRejectsUnsafeSlug(t *testing.T) {
	packDir := writePack(t, `{
		"name": "English Bad",
		"language": "../en",
		"personality": "bad",
		"version": "1.0.0",
		"author": "tester",
		"nsfw": false,
		"description": "Bad pack",
		"lines": {
			"slap": {
				"happy": ["Careful."]
			}
		}
	}`)

	_, _, err := ValidatePack(packDir)
	if err == nil {
		t.Fatal("ValidatePack() error = nil, want slug error")
	}
	if !strings.Contains(err.Error(), "language must be a lowercase slug") {
		t.Fatalf("error = %q", err)
	}
}

func TestValidatePackRejectsUnknownMetadataEvents(t *testing.T) {
	packDir := writePack(t, `{
		"name": "English Bad Metadata",
		"language": "en",
		"personality": "badmeta",
		"version": "1.0.0",
		"author": "tester",
		"nsfw": false,
		"description": "Bad metadata pack",
		"supportedEvents": ["made_up"],
		"lines": {
			"slap": {
				"happy": ["Careful."]
			}
		}
	}`)

	_, _, err := ValidatePack(packDir)
	if err == nil {
		t.Fatal("ValidatePack() error = nil, want metadata error")
	}
	if !strings.Contains(err.Error(), "supportedEvents.made_up uses an unknown event") {
		t.Fatalf("error = %q", err)
	}
}

func TestInstallPackFromDirKeepsExistingPackWhenValidationFails(t *testing.T) {
	packsDir := t.TempDir()
	existingDir := filepath.Join(packsDir, "en_empty")
	if err := os.MkdirAll(existingDir, 0755); err != nil {
		t.Fatal(err)
	}
	marker := filepath.Join(existingDir, "marker.txt")
	if err := os.WriteFile(marker, []byte("keep me"), 0644); err != nil {
		t.Fatal(err)
	}

	invalidPack := writePack(t, `{
		"name": "English Empty",
		"language": "en",
		"personality": "empty",
		"version": "1.0.0",
		"author": "tester",
		"nsfw": false,
		"description": "Empty pack"
	}`)

	if _, _, err := installPackFromDir(invalidPack, packsDir); err == nil {
		t.Fatal("installPackFromDir() error = nil, want validation error")
	}
	data, err := os.ReadFile(marker)
	if err != nil {
		t.Fatalf("existing marker was removed: %v", err)
	}
	if string(data) != "keep me" {
		t.Fatalf("marker = %q", string(data))
	}
}

func TestInstallPackFromDirStagesThenInstallsValidPack(t *testing.T) {
	packsDir := t.TempDir()
	packDir := writePack(t, `{
		"name": "English Test",
		"language": "en",
		"personality": "test",
		"version": "1.0.0",
		"author": "tester",
		"nsfw": false,
		"description": "Test pack",
		"lines": {
			"slap": {
				"happy": ["Careful."]
			}
		}
	}`)
	if err := os.MkdirAll(filepath.Join(packDir, ".git"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(packDir, ".git", "config"), []byte("ignored"), 0644); err != nil {
		t.Fatal(err)
	}

	packName, manifest, err := installPackFromDir(packDir, packsDir)
	if err != nil {
		t.Fatalf("installPackFromDir() error = %v", err)
	}
	if packName != "en_test" || manifest.Name != "English Test" {
		t.Fatalf("installed %q (%s)", packName, manifest.Name)
	}
	if _, err := os.Stat(filepath.Join(packsDir, "en_test", "manifest.json")); err != nil {
		t.Fatalf("installed manifest missing: %v", err)
	}
	if _, err := os.Stat(filepath.Join(packsDir, "en_test", ".git")); !os.IsNotExist(err) {
		t.Fatalf(".git directory should not be installed, stat err = %v", err)
	}
}

func writePack(t *testing.T, manifest string) string {
	t.Helper()
	packDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(packDir, "manifest.json"), []byte(manifest), 0644); err != nil {
		t.Fatal(err)
	}
	return packDir
}
