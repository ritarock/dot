package cmd

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_parseTasks(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		data   string
		want   []task
		hasErr bool
	}{
		{
			name: "parses tasks in order",
			data: "tasks:\n  - name: brew\n    run: brew install jq\n  - name: mise\n    run: mise install\n",
			want: []task{
				{Name: "brew", Run: "brew install jq"},
				{Name: "mise", Run: "mise install"},
			},
		},
		{
			name: "parses a single task",
			data: "tasks:\n  - name: brew\n    run: brew install jq\n",
			want: []task{{Name: "brew", Run: "brew install jq"}},
		},
		{
			name: "parses a multiline run command",
			data: "tasks:\n  - name: brew\n    run: |\n      brew update\n      brew install jq\n",
			want: []task{{Name: "brew", Run: "brew update\nbrew install jq\n"}},
		},
		{
			name:   "fails when data is empty",
			data:   "",
			hasErr: true,
		},
		{
			name:   "fails when tasks key is missing",
			data:   "other: 1\n",
			hasErr: true,
		},
		{
			name:   "fails when tasks is an empty list",
			data:   "tasks: []\n",
			hasErr: true,
		},
		{
			name:   "fails when tasks is null",
			data:   "tasks:\n",
			hasErr: true,
		},
		{
			name:   "fails when name is missing",
			data:   "tasks:\n  - run: echo hi\n",
			hasErr: true,
		},
		{
			name:   "fails when name is empty",
			data:   "tasks:\n  - name: \"\"\n    run: echo hi\n",
			hasErr: true,
		},
		{
			name:   "fails when run is missing",
			data:   "tasks:\n  - name: brew\n",
			hasErr: true,
		},
		{
			name:   "fails when run is empty",
			data:   "tasks:\n  - name: brew\n    run: \"\"\n",
			hasErr: true,
		},
		{
			name:   "fails when task names are duplicated",
			data:   "tasks:\n  - name: brew\n    run: echo a\n  - name: brew\n    run: echo b\n",
			hasErr: true,
		},
		{
			name:   "fails when yaml is malformed",
			data:   "tasks: [\n",
			hasErr: true,
		},
		{
			name:   "fails when an unknown key is present",
			data:   "tasks:\n  - name: brew\n    cmd: echo hi\n",
			hasErr: true,
		},
		{
			name:   "fails when tasks is not a list",
			data:   "tasks: brew\n",
			hasErr: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			got, err := parseTasks([]byte(test.data))

			if test.hasErr {
				assert.Error(t, err)
				assert.Nil(t, got)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, test.want, got)
		})
	}
}

func Test_loadTasks(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		setup  func(t *testing.T, dir string) string
		want   []task
		hasErr bool
	}{
		{
			name: "loads tasks from runner.yaml",
			setup: func(t *testing.T, dir string) string {
				writeFiles(t, dir, map[string]string{runnerFile: "tasks:\n  - name: brew\n    run: brew install jq\n"})
				return dir
			},
			want: []task{{Name: "brew", Run: "brew install jq"}},
		},
		{
			name: "fails when runner.yaml does not exist",
			setup: func(t *testing.T, dir string) string {
				writeFiles(t, dir, map[string]string{"dot_zshrc": "x"})
				return dir
			},
			hasErr: true,
		},
		{
			name: "fails when repo does not exist",
			setup: func(t *testing.T, dir string) string {
				return filepath.Join(dir, "missing")
			},
			hasErr: true,
		},
		{
			name: "fails when runner.yaml is a directory",
			setup: func(t *testing.T, dir string) string {
				require.NoError(t, os.MkdirAll(filepath.Join(dir, runnerFile), 0o755))
				return dir
			},
			hasErr: true,
		},
		{
			name: "fails when runner.yaml is invalid",
			setup: func(t *testing.T, dir string) string {
				writeFiles(t, dir, map[string]string{runnerFile: "tasks: [\n"})
				return dir
			},
			hasErr: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			repo := test.setup(t, t.TempDir())

			got, err := loadTasks(repo)

			if test.hasErr {
				assert.Error(t, err)
				assert.Nil(t, got)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, test.want, got)
		})
	}
}

func Test_selectTasks(t *testing.T) {
	t.Parallel()

	all := []task{
		{Name: "brew", Run: "brew install jq"},
		{Name: "mise", Run: "mise install"},
	}

	tests := []struct {
		name   string
		names  []string
		want   []task
		hasErr bool
	}{
		{
			name:  "returns all tasks when no names given",
			names: nil,
			want:  all,
		},
		{
			name:  "returns the named task",
			names: []string{"mise"},
			want:  []task{{Name: "mise", Run: "mise install"}},
		},
		{
			name:  "returns tasks in argument order",
			names: []string{"mise", "brew"},
			want: []task{
				{Name: "mise", Run: "mise install"},
				{Name: "brew", Run: "brew install jq"},
			},
		},
		{
			name:  "returns the task twice when named twice",
			names: []string{"brew", "brew"},
			want: []task{
				{Name: "brew", Run: "brew install jq"},
				{Name: "brew", Run: "brew install jq"},
			},
		},
		{
			name:   "fails when the name is not defined",
			names:  []string{"nope"},
			hasErr: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			got, err := selectTasks(all, test.names)

			if test.hasErr {
				assert.Error(t, err)
				assert.Nil(t, got)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, test.want, got)
		})
	}
}

func Test_execTask(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		run      string
		wantFile string
		hasErr   bool
	}{
		{
			name: "succeeds when command exits zero",
			run:  "true",
		},
		{
			name:   "fails when command exits non-zero",
			run:    "exit 3",
			hasErr: true,
		},
		{
			name:     "runs the command in the given directory",
			run:      "touch marker",
			wantFile: "marker",
		},
		{
			name:     "expands shell syntax",
			run:      `touch "$(printf out)"`,
			wantFile: "out",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			dir := t.TempDir()

			err := execTask(task{Name: test.name, Run: test.run}, dir)

			if test.hasErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			if test.wantFile != "" {
				assert.FileExists(t, filepath.Join(dir, test.wantFile))
			}
		})
	}
}
