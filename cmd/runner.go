package cmd

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"go.yaml.in/yaml/v3"
)

const runnerFile = "runner.yaml"

type task struct {
	Name string `yaml:"name"`
	Run  string `yaml:"run"`
}

type manifest struct {
	Tasks []task `yaml:"tasks"`
}

type runFunc func(t task, dir string) error

func loadTasks(repo string) ([]task, error) {
	file := filepath.Join(repo, runnerFile)
	data, err := os.ReadFile(file)
	if errors.Is(err, fs.ErrNotExist) {
		return nil, fmt.Errorf("%s does not exist (create it to define tasks)", file)
	}
	if err != nil {
		return nil, err
	}

	return parseTasks(data)
}

func parseTasks(data []byte) ([]task, error) {
	var m manifest
	dec := yaml.NewDecoder(bytes.NewReader(data))
	dec.KnownFields(true)
	if err := dec.Decode(&m); err != nil && !errors.Is(err, io.EOF) {
		return nil, fmt.Errorf("%s is invalid: %v", runnerFile, err)
	}

	if len(m.Tasks) == 0 {
		return nil, fmt.Errorf("%s has no tasks", runnerFile)
	}

	seen := map[string]bool{}
	for i, t := range m.Tasks {
		if strings.TrimSpace(t.Name) == "" {
			return nil, fmt.Errorf("%s task %d has no name", runnerFile, i+1)
		}
		if strings.TrimSpace(t.Run) == "" {
			return nil, fmt.Errorf("%s task %q has no run command", runnerFile, t.Name)
		}
		if seen[t.Name] {
			return nil, fmt.Errorf("%s has duplicate task %q", runnerFile, t.Name)
		}
		seen[t.Name] = true
	}

	return m.Tasks, nil
}

func selectTasks(tasks []task, names []string) ([]task, error) {
	if len(names) == 0 {
		return tasks, nil
	}

	byName := map[string]task{}
	for _, t := range tasks {
		byName[t.Name] = t
	}

	var selected []task
	for _, name := range names {
		t, ok := byName[name]
		if !ok {
			return nil, fmt.Errorf("task %q is not defined in %s", name, runnerFile)
		}
		selected = append(selected, t)
	}

	return selected, nil
}

func execTask(t task, dir string) error {
	cmd := exec.Command("sh", "-c", t.Run)
	cmd.Dir = dir
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}
