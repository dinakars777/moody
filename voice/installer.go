package voice

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"time"
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

	packsDir, err := defaultPacksDir()
	if err != nil {
		return err
	}

	packName, manifest, err := installPackFromDir(tempDir, packsDir)
	if err != nil {
		return err
	}

	fmt.Printf("✨ Successfully installed pack: %s (%s)\n", packName, manifest.Name)
	fmt.Printf("   You can now use it with: moody --pack %s\n", packName)
	return nil
}

func defaultPacksDir() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(homeDir, ".moody", "packs"), nil
}

func installPackFromDir(srcDir, packsDir string) (string, *Manifest, error) {
	manifest, packName, err := ValidatePack(srcDir)
	if err != nil {
		return "", nil, err
	}

	destDir := filepath.Join(packsDir, packName)
	fmt.Printf("   Updating local pack at: %s\n", destDir)

	if err := os.MkdirAll(packsDir, 0755); err != nil {
		return "", nil, err
	}

	stagingDir, err := os.MkdirTemp(packsDir, "."+packName+"-*")
	if err != nil {
		return "", nil, fmt.Errorf("failed to create staging directory: %w", err)
	}
	defer os.RemoveAll(stagingDir)

	if err := copyDir(srcDir, stagingDir); err != nil {
		return "", nil, fmt.Errorf("failed to stage pack: %w", err)
	}
	if _, _, err := ValidatePack(stagingDir); err != nil {
		return "", nil, fmt.Errorf("staged pack failed validation: %w", err)
	}

	var backupDir string
	if _, err := os.Stat(destDir); err == nil {
		backupDir = filepath.Join(packsDir, fmt.Sprintf(".%s-backup-%d", packName, time.Now().UnixNano()))
		if err := os.Rename(destDir, backupDir); err != nil {
			return "", nil, fmt.Errorf("failed to move existing pack aside: %w", err)
		}
		defer os.RemoveAll(backupDir)
	} else if !errors.Is(err, os.ErrNotExist) {
		return "", nil, fmt.Errorf("failed to inspect existing pack: %w", err)
	}

	if err := os.Rename(stagingDir, destDir); err != nil {
		if backupDir != "" {
			_ = os.Rename(backupDir, destDir)
		}
		return "", nil, fmt.Errorf("failed to install staged pack: %w", err)
	}

	return packName, manifest, nil
}

func copyDir(srcDir, destDir string) error {
	return filepath.WalkDir(srcDir, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(srcDir, path)
		if err != nil {
			return err
		}
		if rel == "." {
			return nil
		}
		if entry.Name() == ".git" && entry.IsDir() {
			return filepath.SkipDir
		}
		if entry.Type()&os.ModeSymlink != 0 {
			return fmt.Errorf("refusing to install symlink: %s", rel)
		}

		target := filepath.Join(destDir, rel)
		info, err := entry.Info()
		if err != nil {
			return err
		}
		if entry.IsDir() {
			return os.MkdirAll(target, info.Mode())
		}
		return copyFile(path, target, info.Mode())
	})
}

func copyFile(srcPath, destPath string, mode fs.FileMode) error {
	if err := os.MkdirAll(filepath.Dir(destPath), 0755); err != nil {
		return err
	}
	src, err := os.Open(srcPath)
	if err != nil {
		return err
	}
	defer src.Close()

	dest, err := os.OpenFile(destPath, os.O_CREATE|os.O_EXCL|os.O_WRONLY, mode)
	if err != nil {
		return err
	}
	defer dest.Close()

	_, err = io.Copy(dest, src)
	return err
}
