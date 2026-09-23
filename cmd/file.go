package cmd

import (
	"bytes"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
)

const dotPrefix = "dot_"

func encodeRel(rel string) string {
	sep := string(filepath.Separator)
	elems := strings.Split(rel, sep)
	for i, e := range elems {
		if strings.HasPrefix(e, ".") && e != "." && e != ".." {
			elems[i] = dotPrefix + e[len("."):]
		}
	}
	return strings.Join(elems, sep)
}

func decodeRel(rel string) string {
	sep := string(filepath.Separator)
	elems := strings.Split(rel, sep)
	for i, e := range elems {
		if strings.HasPrefix(e, dotPrefix) {
			elems[i] = "." + e[len(dotPrefix):]
		}
	}
	return strings.Join(elems, sep)
}

func copyFile(src, dst string) error {
	info, err := os.Stat(src)
	if err != nil {
		return err
	}
	data, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return err
	}
	if err := os.WriteFile(dst, data, info.Mode().Perm()); err != nil {
		return err
	}

	return os.Chmod(dst, info.Mode().Perm())
}

func managedFiles(repo string) ([]string, error) {
	var files []string
	err := filepath.WalkDir(repo, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if path != repo && strings.HasPrefix(d.Name(), ".") {
			if d.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		if d.IsDir() {
			return nil
		}
		rel, err := filepath.Rel(repo, path)
		if err != nil {
			return err
		}
		if rel == runnerFile {
			return nil
		}
		files = append(files, rel)
		return nil
	})
	if errors.Is(err, fs.ErrNotExist) {
		return nil, fmt.Errorf("repo directory %s does not exist (add a file first)", repo)
	}
	if err != nil {
		return nil, err
	}
	slices.Sort(files)
	return files, nil
}

func collectFiles(root string) ([]string, error) {
	info, err := os.Stat(root)
	if err != nil {
		return nil, err
	}
	if !info.IsDir() {
		if !info.Mode().IsRegular() {
			return nil, fmt.Errorf("%s is not a regular file", root)
		}
		return []string{root}, nil
	}

	var files []string
	err = filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			if path != root && d.Name() == ".git" {
				return filepath.SkipDir
			}
			return nil
		}
		if !d.Type().IsRegular() {
			fmt.Println("skip (not a regular file):", path)
			return nil
		}
		files = append(files, path)
		return nil
	})
	if err != nil {
		return nil, err
	}
	slices.Sort(files)

	return files, nil
}

func filesDiffer(a, b string) (bool, error) {
	da, err := os.ReadFile(a)
	if err != nil {
		return false, err
	}
	db, err := os.ReadFile(b)
	if errors.Is(err, fs.ErrNotExist) {
		return true, nil
	}
	if err != nil {
		return false, err
	}

	return !bytes.Equal(da, db), nil
}

func showDiff(old, new string) error {
	if _, err := os.Stat(old); errors.Is(err, fs.ErrNotExist) {
		old = os.DevNull
	}

	cmd := exec.Command("git", "diff", "--no-index", "--color", "--", old, new)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) && exitErr.ExitCode() == 1 {
		return nil
	}
	return err
}
