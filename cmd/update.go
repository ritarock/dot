package cmd

import (
	"github.com/spf13/cobra"
)

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "copy $HOME -> repo",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		repo, home, err := dirs()
		if err != nil {
			return err
		}

		return sync(repo, home, false, cmd.InOrStdin(), update)
	},
}

func init() {
	rootCmd.AddCommand(updateCmd)
}
