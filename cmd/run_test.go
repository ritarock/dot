package cmd

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_runTasks(t *testing.T) {
	t.Parallel()

	two := []task{
		{Name: "a", Run: "echo a"},
		{Name: "b", Run: "echo b"},
	}

	tests := []struct {
		name    string
		tasks   []task
		yes     bool
		stdin   string
		failOn  string
		wantRan []string
		hasErr  bool
	}{
		{
			name:    "runs all tasks when yes",
			tasks:   two,
			yes:     true,
			wantRan: []string{"a", "b"},
		},
		{
			name:    "runs tasks when confirmed with y",
			tasks:   two,
			stdin:   "y\n",
			wantRan: []string{"a", "b"},
		},
		{
			name:    "runs nothing when confirmation is declined",
			tasks:   two,
			stdin:   "n\n",
			wantRan: nil,
		},
		{
			name:    "runs nothing when stdin is empty",
			tasks:   two,
			stdin:   "",
			wantRan: nil,
		},
		{
			name:    "stops at the first failing task",
			tasks:   two,
			yes:     true,
			failOn:  "a",
			wantRan: []string{"a"},
			hasErr:  true,
		},
		{
			name:    "runs a single task",
			tasks:   []task{{Name: "a", Run: "echo a"}},
			yes:     true,
			wantRan: []string{"a"},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			dir := t.TempDir()
			var ran []string
			var gotDir string
			run := func(tk task, d string) error {
				ran = append(ran, tk.Name)
				gotDir = d
				if tk.Name == test.failOn {
					return errors.New("boom")
				}
				return nil
			}

			err := runTasks(test.tasks, dir, test.yes, strings.NewReader(test.stdin), run)

			if test.hasErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
			}
			assert.Equal(t, test.wantRan, ran)
			if len(ran) > 0 {
				assert.Equal(t, dir, gotDir)
			}
		})
	}
}

func Test_cmdRun(t *testing.T) {
	t.Parallel()

	const bothTasks = "tasks:\n  - name: a\n    run: touch a\n  - name: b\n    run: touch b\n"

	tests := []struct {
		name       string
		repoFiles  map[string]string
		names      []string
		yes        bool
		stdin      string
		wantFiles  map[string]string
		wantAbsent []string
		hasErr     bool
	}{
		{
			name:      "runs all tasks from runner.yaml when yes",
			repoFiles: map[string]string{runnerFile: bothTasks},
			yes:       true,
			wantFiles: map[string]string{"a": "", "b": ""},
		},
		{
			name:       "runs only the named task",
			repoFiles:  map[string]string{runnerFile: bothTasks},
			names:      []string{"b"},
			yes:        true,
			wantFiles:  map[string]string{"b": ""},
			wantAbsent: []string{"a"},
		},
		{
			name:      "runs tasks in argument order",
			repoFiles: map[string]string{runnerFile: "tasks:\n  - name: a\n    run: printf a >> log\n  - name: b\n    run: printf b >> log\n"},
			names:     []string{"b", "a"},
			yes:       true,
			wantFiles: map[string]string{"log": "ba"},
		},
		{
			name:       "runs nothing when confirmation is declined",
			repoFiles:  map[string]string{runnerFile: bothTasks},
			stdin:      "n\n",
			wantAbsent: []string{"a", "b"},
		},
		{
			name:      "fails when runner.yaml does not exist",
			repoFiles: map[string]string{"dot_zshrc": "x"},
			yes:       true,
			hasErr:    true,
		},
		{
			name:       "fails when the task name is not defined",
			repoFiles:  map[string]string{runnerFile: bothTasks},
			names:      []string{"nope"},
			yes:        true,
			wantAbsent: []string{"a", "b"},
			hasErr:     true,
		},
		{
			name:      "fails when a task exits non-zero",
			repoFiles: map[string]string{runnerFile: "tasks:\n  - name: a\n    run: exit 1\n"},
			yes:       true,
			hasErr:    true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			repo := filepath.Join(t.TempDir(), "repo")
			writeFiles(t, repo, test.repoFiles)

			err := cmdRun(repo, test.names, test.yes, strings.NewReader(test.stdin))

			if test.hasErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
			}

			for rel, content := range test.wantFiles {
				path := filepath.Join(repo, rel)
				require.FileExists(t, path)
				data, err := os.ReadFile(path)
				require.NoError(t, err)
				assert.Equal(t, content, string(data))
			}
			for _, rel := range test.wantAbsent {
				assert.NoFileExists(t, filepath.Join(repo, rel))
			}
		})
	}
}
