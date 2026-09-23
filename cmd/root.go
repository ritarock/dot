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
mirrors the structure of $HOME, with leading dots renamed to "dot_"
(e.g. ~/.config/nvim -> dot_config/nvim). Copying has no configuration;
the only special file is an optional runner.yaml at the repository root,
which defines the tasks that "dot run" executes.`,
	SilenceUsage:      true,
	CompletionOptions: cobra.CompletionOptions{DisableDefaultCmd: true},
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
