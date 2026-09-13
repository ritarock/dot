package cmd

import (
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "dot",
	Short: "A minimal copy-based dotfiles manager",
	Long: `A minimal copy-based dotfiles manager

The repository directory (default ~/dotfiles, override with $DOT_DIR)
mirrors the structure of $HOME. There is no config file.`,
	SilenceUsage: true,
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

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
