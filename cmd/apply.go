package cmd

import (
	"github.com/spf13/cobra"
)

var applyCmd = &cobra.Command{
	Use:   "apply",
	Short: "copy repo -> $HOME (confirm unless -y)",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		repo, home, err := dirs()
		if err != nil {
			return err
		}
		yes, _ := cmd.Flags().GetBool("yes")
		return sync(repo, home, yes, cmd.InOrStdin(), apply)
	},
}

func init() {
	rootCmd.AddCommand(applyCmd)
	applyCmd.Flags().BoolP("yes", "y", false, "skip confirmation")
}
