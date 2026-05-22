package voice

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// InstallPack clones a remote repository containing a moody pack and installs it.
func InstallPack(repoURL string) error {
	fmt.Printf("📦 Installing community pack from %s\n", repoURL)

	tempDir, err := os.MkdirTemp("", "moody-install-*")
	if err != nil {
		return fmt.Errorf("failed to create temp directory: %w", err)
	}
	defer os.RemoveAll(tempDir) // clean up afterwards

	// Clone the repository
	fmt.Println("   Cloning repository...")
	cmd := exec.Command("git", "clone", "--depth", "1", repoURL, tempDir)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("git clone failed (ensure you have git installed and the URL is correct): %w", err)
	}

	// Look for manifest.json
	manifestPath := filepath.Join(tempDir, "manifest.json")
	if _, err := os.Stat(manifestPath); os.IsNotExist(err) {
		return fmt.Errorf("invalid pack: manifest.json not found in the repository root")
	}

	// Read manifest to get the pack's internal name
	data, err := os.ReadFile(manifestPath)
	if err != nil {
		return fmt.Errorf("failed to read manifest.json: %w", err)
	}

	var m Manifest
	if err := json.Unmarshal(data, &m); err != nil {
		return fmt.Errorf("manifest.json is invalid JSON: %w", err)
	}

	// Determine internal pack folder name. Usually language_personality (e.g. en_gordon)
	packName := strings.ToLower(m.Language + "_" + m.Personality)
	if packName == "_" {
		return fmt.Errorf("manifest.json must specify 'language' and 'personality'")
	}

	// Setup destination ~/.moody/packs/<packName>
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	destDir := filepath.Join(homeDir, ".moody", "packs", packName)

	fmt.Printf("   Updating local pack at: %s\n", destDir)

	// Remove old version if it exists
	if _, err := os.Stat(destDir); err == nil {
		os.RemoveAll(destDir)
	}

	// Make parent directories just in case
	if err := os.MkdirAll(filepath.Dir(destDir), 0755); err != nil {
		return err
	}

	// Move the temp folder to the official packs directory
	if err := os.Rename(tempDir, destDir); err != nil {
		// Sometimes Rename fails across mounted drives, try cp -r
		cpCmd := exec.Command("cp", "-r", tempDir, destDir)
		if err := cpCmd.Run(); err != nil {
			return fmt.Errorf("failed to copy pack to ~/.moody/packs: %w", err)
		}
	}

	fmt.Printf("✨ Successfully installed pack: %s (%s)\n", packName, m.Name)
	fmt.Printf("   You can now use it with: moody --pack %s\n", packName)
	return nil
}
