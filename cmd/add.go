package cmd

import (
	"fmt"
	"path/filepath"

	"github.com/spf13/cobra"
)

var addCmd = &cobra.Command{
	Use:   "add <path>...",
	Short: "copy files or directories from $HOME into the repo",
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
		if rel == "." {
			return fmt.Errorf("%s is the home directory itself; add paths under it instead", abs)
		}
		targets, err := collectFiles(abs)
		if err != nil {
			return err
		}
		for _, target := range targets {
			rel, err := filepath.Rel(home, target)
			if err != nil {
				return err
			}
			repoRel := encodeRel(rel)
			if decodeRel(repoRel) != rel {
				return fmt.Errorf("%s cannot be added: path elements starting with %q are reserved", target, dotPrefix)
			}
			if err := copyFile(target, filepath.Join(repo, repoRel)); err != nil {
				return err
			}
			fmt.Println("added", rel)
		}
	}
	return nil
}
