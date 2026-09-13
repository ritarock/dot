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

func sync(repo, home string, yes bool, stdin io.Reader, act action) error {
	files, err := managedFiles(repo)
	if err != nil {
		return err
	}

	var changed []string
	for _, rel := range files {
		repoPath := filepath.Join(repo, rel)
		homePath := filepath.Join(home, rel)
		if act == update {
			if _, err := os.Stat(homePath); errors.Is(err, fs.ErrNotExist) {
				fmt.Println("skip (missing in home): ", rel)
				continue
			}
		}
		differ, err := filesDiffer(repoPath, homePath)
		if err != nil {
			return err
		}
		if !differ {
			continue
		}
		changed = append(changed, rel)
		if act == apply {
			err = showDiff(homePath, repoPath)
		} else {
			err = showDiff(repoPath, homePath)
		}
		if err != nil {
			return err
		}
	}

	if len(changed) == 0 {
		fmt.Println("already updated")
		return nil
	}

	if act == apply && !yes && !confirm(fmt.Sprintf("Apply %d file(s)? [y/N]: ", len(changed)), stdin) {
		fmt.Println("aborted")
		return nil
	}

	for _, rel := range changed {
		src, dst := filepath.Join(repo, rel), filepath.Join(home, rel)
		if act == update {
			src, dst = dst, src
		}
		if err := copyFile(src, dst); err != nil {
			return err
		}
		fmt.Println("copied", rel)
	}

	return nil
}

func confirm(prompt string, stdin io.Reader) bool {
	fmt.Println(prompt)
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
