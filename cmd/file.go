package cmd

import (
	"os"
	"path/filepath"
)

func dirs() (repo, home string, err error) {
	home, err = os.UserHomeDir()
	if err != nil {
		return "", "", err
	}

	repo = os.Getenv("DOT_DIR")
	if repo == "" {
		repo = filepath.Join(home, "dotfiles")
	}

	return repo, home, nil
}

func copyFile(src, dst string) error {
	info, err := os.Stat(src)
	if err != nil {
		return err
	}
	data, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return err
	}
	if err := os.WriteFile(dst, data, info.Mode().Perm()); err != nil {
		return err
	}

	return os.Chmod(dst, info.Mode().Perm())
}
