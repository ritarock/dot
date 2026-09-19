package cmd

import (
	"fmt"
	"os"
	"path/filepath"

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
		if err != nil || !filepath.IsLocal(rel) {
			return fmt.Errorf("%s is not under home directory %s", abs, home)
		}
		repoRel := encodeRel(rel)
		if decodeRel(repoRel) != rel {
			return fmt.Errorf("%s cannot be added: path elements starting with %q are reserved", abs, dotPrefix)
		}
		info, err := os.Stat(abs)
		if err != nil {
			return err
		}
		if !info.Mode().IsRegular() {
			return fmt.Errorf("%s is not a regular file", abs)
		}
		if err := copyFile(abs, filepath.Join(repo, repoRel)); err != nil {
			return err
		}
		fmt.Println("added", rel)
	}
	return nil
}
