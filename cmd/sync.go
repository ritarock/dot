package cmd

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

type action int

const (
	apply action = iota + 1
	update
)

// roots returns the copy source and destination roots for act.
func (act action) roots(repo, home string) (src, dst string) {
	if act == update {
		return home, repo
	}
	return repo, home
}

// changedFiles shows the diff of each managed file that act would change
// and returns their relative paths.
func changedFiles(repo, home string, act action) ([]string, error) {
	files, err := managedFiles(repo)
	if err != nil {
		return nil, err
	}

	src, dst := act.roots(repo, home)
	var changed []string
	for _, rel := range files {
		srcPath, dstPath := filepath.Join(src, rel), filepath.Join(dst, rel)
		if act == update {
			if _, err := os.Stat(srcPath); errors.Is(err, fs.ErrNotExist) {
				fmt.Println("skip (missing in home):", rel)
				continue
			}
		}
		differ, err := filesDiffer(srcPath, dstPath)
		if err != nil {
			return nil, err
		}
		if !differ {
			continue
		}
		if err := showDiff(dstPath, srcPath); err != nil {
			return nil, err
		}
		changed = append(changed, rel)
	}

	return changed, nil
}

func sync(repo, home string, yes bool, stdin io.Reader, act action) error {
	changed, err := changedFiles(repo, home, act)
	if err != nil {
		return err
	}

	if len(changed) == 0 {
		fmt.Println("already up to date")
		return nil
	}

	if act == apply && !yes && !confirm(fmt.Sprintf("Apply %d file(s)? [y/N]: ", len(changed)), stdin) {
		fmt.Println("aborted")
		return nil
	}

	src, dst := act.roots(repo, home)
	for _, rel := range changed {
		if err := copyFile(filepath.Join(src, rel), filepath.Join(dst, rel)); err != nil {
			return err
		}
		fmt.Println("copied", rel)
	}

	return nil
}

func confirm(prompt string, stdin io.Reader) bool {
	fmt.Print(prompt)
	line, err := bufio.NewReader(stdin).ReadString('\n')
	if err != nil && line == "" {
		return false
	}
	switch strings.ToLower(strings.TrimSpace(line)) {
	case "y", "yes":
		return true
	}
	return false
}
