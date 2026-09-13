package cmd

import (
	"fmt"
	"path/filepath"

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
	files, err := managedFiles(repo)
	if err != nil {
		return err
	}

	changed := false
	for _, rel := range files {
		differ, err := filesDiffer(filepath.Join(repo, rel), filepath.Join(home, rel))
		if err != nil {
			return err
		}
		if !differ {
			continue
		}
		changed = true
		if err := showDiff(filepath.Join(home, rel), filepath.Join(repo, rel)); err != nil {
			return err
		}
	}
	if !changed {
		fmt.Println("no changed")
	}

	return nil
}
