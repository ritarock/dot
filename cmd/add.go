package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

var addCmd = &cobra.Command{
	Use:   "add <file>...",
	Short: "copy files from $HOME into the repo",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		repo, home, err := dirs()
		if err != nil {
			return err
		}
		return cmdAdd(repo, home, args)
	},
}

func init() {
	rootCmd.AddCommand(addCmd)
}

func cmdAdd(repo, home string, paths []string) error {
	for _, p := range paths {
		abs, err := filepath.Abs(p)
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(home, abs)
		if err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
			return fmt.Errorf("%s is not under home directory %s", abs, home)
		}
		info, err := os.Stat(abs)
		if err != nil {
			return err
		}
		if !info.Mode().IsRegular() {
			return fmt.Errorf("%s is not a regular file", abs)
		}
		if err := copyFile(abs, filepath.Join(repo, rel)); err != nil {
			return err
		}
		fmt.Println("added", rel)
	}
	return nil
}
