package cmd

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_cmdAdd(t *testing.T) {
	t.Parallel()
	const wantContent = "hello, world\n"
	const wantPerm = os.FileMode(0o640)

	tests := []struct {
		name     string
		setup    func(t *testing.T, home, repo string) []string
		wantRels []string
		hasErr   bool
	}{
		{
			name: "adds a file directly under home",
			setup: func(t *testing.T, home, repo string) []string {
				writeFile(t, filepath.Join(home, ".zshrc"), wantContent, wantPerm)
				return []string{filepath.Join(home, ".zshrc")}
			},
			wantRels: []string{"dot_zshrc"},
		},
		{
			name: "adds a nested file preserving relative path",
			setup: func(t *testing.T, home, repo string) []string {
				writeFile(t, filepath.Join(home, ".config", "nvim", "init.lua"), wantContent, wantPerm)
				return []string{filepath.Join(home, ".config", "nvim", "init.lua")}
			},
			wantRels: []string{filepath.Join("dot_config", "nvim", "init.lua")},
		},
		{
			name: "adds multiple files",
			setup: func(t *testing.T, home, repo string) []string {
				writeFile(t, filepath.Join(home, ".zshrc"), wantContent, wantPerm)
				writeFile(t, filepath.Join(home, ".gitconfig"), wantContent, wantPerm)
				return []string{
					filepath.Join(home, ".zshrc"),
					filepath.Join(home, ".gitconfig"),
				}
			},
			wantRels: []string{"dot_zshrc", "dot_gitconfig"},
		},
		{
			name: "overwrites existing file in repo",
			setup: func(t *testing.T, home, repo string) []string {
				writeFile(t, filepath.Join(home, ".zshrc"), wantContent, wantPerm)
				writeFile(t, filepath.Join(repo, "dot_zshrc"), "old", 0o600)
				return []string{filepath.Join(home, ".zshrc")}
			},
			wantRels: []string{"dot_zshrc"},
		},
		{
			name: "does nothing when paths is empty",
			setup: func(t *testing.T, home, repo string) []string {
				return nil
			},
			wantRels: nil,
		},
		{
			name: "fails when path is outside home",
			setup: func(t *testing.T, home, repo string) []string {
				outside := filepath.Join(t.TempDir(), "outside.txt")
				writeFile(t, outside, wantContent, wantPerm)
				return []string{outside}
			},
			hasErr: true,
		},
		{
			name: "fails when path is the parent of home",
			setup: func(t *testing.T, home, repo string) []string {
				return []string{filepath.Dir(home)}
			},
			hasErr: true,
		},
		{
			name: "fails when file does not exist",
			setup: func(t *testing.T, home, repo string) []string {
				return []string{filepath.Join(home, "missing.txt")}
			},
			hasErr: true,
		},
		{
			name: "fails when path has element starting with dot_",
			setup: func(t *testing.T, home, repo string) []string {
				writeFile(t, filepath.Join(home, "dot_foo"), wantContent, wantPerm)
				return []string{filepath.Join(home, "dot_foo")}
			},
			hasErr: true,
		},
		{
			name: "fails when path is a directory",
			setup: func(t *testing.T, home, repo string) []string {
				dir := filepath.Join(home, ".config")
				require.NoError(t, os.MkdirAll(dir, 0o755))
				return []string{dir}
			},
			hasErr: true,
		},
		{
			name: "leaves earlier files copied when a later path fails",
			setup: func(t *testing.T, home, repo string) []string {
				writeFile(t, filepath.Join(home, ".zshrc"), wantContent, wantPerm)
				return []string{
					filepath.Join(home, ".zshrc"),
					filepath.Join(home, "missing.txt"),
				}
			},
			wantRels: []string{"dot_zshrc"},
			hasErr:   true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			home := t.TempDir()
			repo := t.TempDir()
			paths := test.setup(t, home, repo)

			err := cmdAdd(repo, home, paths)

			if test.hasErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
			}

			for _, rel := range test.wantRels {
				assertFile(t, filepath.Join(repo, rel), wantContent, wantPerm)
			}
		})
	}
}
