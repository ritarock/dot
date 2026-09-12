package cmd

import (
	"os"

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
