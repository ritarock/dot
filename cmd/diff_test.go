package cmd

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_cmdDiff(t *testing.T) {
	t.Parallel()

	mkFiles := func(t *testing.T, dir string, files map[string]string) {
		t.Helper()
		for rel, content := range files {
			path := filepath.Join(dir, rel)
			require.NoError(t, os.MkdirAll(filepath.Dir(path), 0o755))
			require.NoError(t, os.WriteFile(path, []byte(content), 0o644))
		}
	}
	readFiles := func(t *testing.T, dir string) map[string]string {
		t.Helper()
		got := map[string]string{}
		err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
			if err != nil || d.IsDir() {
				return err
			}
			rel, err := filepath.Rel(dir, path)
			if err != nil {
				return err
			}
			data, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			got[rel] = string(data)
			return nil
		})
		if !errors.Is(err, fs.ErrNotExist) {
			require.NoError(t, err)
		}
		return got
	}

	tests := []struct {
		name   string
		setup  func(t *testing.T) (repo, home string)
		hasErr bool
	}{
		{
			name: "succeeds when nothing changed",
			setup: func(t *testing.T) (string, string) {
				repo, home := t.TempDir(), t.TempDir()
				mkFiles(t, repo, map[string]string{".zshrc": "same"})
				mkFiles(t, home, map[string]string{".zshrc": "same"})
				return repo, home
			},
		},
		{
			name: "succeeds when file differs",
			setup: func(t *testing.T) (string, string) {
				repo, home := t.TempDir(), t.TempDir()
				mkFiles(t, repo, map[string]string{".zshrc": "new"})
				mkFiles(t, home, map[string]string{".zshrc": "old"})
				return repo, home
			},
		},
		{
			name: "succeeds when file is missing in home",
			setup: func(t *testing.T) (string, string) {
				repo, home := t.TempDir(), t.TempDir()
				mkFiles(t, repo, map[string]string{".config/nvim/init.lua": "lua"})
				return repo, home
			},
		},
		{
			name: "ignores files only in home",
			setup: func(t *testing.T) (string, string) {
				repo, home := t.TempDir(), t.TempDir()
				mkFiles(t, repo, map[string]string{".zshrc": "same"})
				mkFiles(t, home, map[string]string{".zshrc": "same", ".bashrc": "bash"})
				return repo, home
			},
		},
		{
			name: "fails when repo does not exist",
			setup: func(t *testing.T) (string, string) {
				return filepath.Join(t.TempDir(), "missing"), t.TempDir()
			},
			hasErr: true,
		},
		{
			name: "fails when home has directory with same name",
			setup: func(t *testing.T) (string, string) {
				repo, home := t.TempDir(), t.TempDir()
				mkFiles(t, repo, map[string]string{".config": "x"})
				mkFiles(t, home, map[string]string{".config/nvim/init.lua": "lua"})
				return repo, home
			},
			hasErr: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			repo, home := test.setup(t)
			repoBefore, homeBefore := readFiles(t, repo), readFiles(t, home)

			err := cmdDiff(repo, home)

			if test.hasErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, repoBefore, readFiles(t, repo))
			assert.Equal(t, homeBefore, readFiles(t, home))
		})
	}
}
