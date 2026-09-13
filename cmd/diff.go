package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var diffCmd = &cobra.Command{
	Use:   "diff",
	Short: "show what apply would change",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		repo, home, err := dirs()
		if err != nil {
			return err
		}
		return cmdDiff(repo, home)
	},
}

func init() {
	rootCmd.AddCommand(diffCmd)
}

func cmdDiff(repo, home string) error {
	changed, err := changedFiles(repo, home, apply)
	if err != nil {
		return err
	}
	if len(changed) == 0 {
		fmt.Println("no changes")
	}

	return nil
}
