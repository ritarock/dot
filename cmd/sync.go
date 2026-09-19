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

func (act action) paths(repo, home, repoRel string) (src, dst string) {
	repoPath, homePath := filepath.Join(repo, repoRel), filepath.Join(home, decodeRel(repoRel))
	if act == update {
		return homePath, repoPath
	}
	return repoPath, homePath
}

func changedFiles(repo, home string, act action) ([]string, error) {
	files, err := managedFiles(repo)
	if err != nil {
		return nil, err
	}

	var changed []string
	for _, rel := range files {
		srcPath, dstPath := act.paths(repo, home, rel)
		if act == update {
			if _, err := os.Stat(srcPath); errors.Is(err, fs.ErrNotExist) {
				fmt.Println("skip (missing in home):", decodeRel(rel))
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

	for _, rel := range changed {
		if err := copyFile(act.paths(repo, home, rel)); err != nil {
			return err
		}
		fmt.Println("copied", decodeRel(rel))
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
