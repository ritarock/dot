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

	writeFile := func(t *testing.T, path string) {
		t.Helper()
		require.NoError(t, os.MkdirAll(filepath.Dir(path), 0o755))
		require.NoError(t, os.WriteFile(path, []byte(wantContent), wantPerm))
		require.NoError(t, os.Chmod(path, wantPerm))
	}

	tests := []struct {
		name     string
		setup    func(t *testing.T, home, repo string) []string
		wantRels []string
		hasErr   bool
	}{
		{
			name: "adds a file directly under home",
			setup: func(t *testing.T, home, repo string) []string {
				writeFile(t, filepath.Join(home, ".zshrc"))
				return []string{filepath.Join(home, ".zshrc")}
			},
			wantRels: []string{".zshrc"},
		},
		{
			name: "adds a nested file preserving relative path",
			setup: func(t *testing.T, home, repo string) []string {
				writeFile(t, filepath.Join(home, ".config", "nvim", "init.lua"))
				return []string{filepath.Join(home, ".config", "nvim", "init.lua")}
			},
			wantRels: []string{filepath.Join(".config", "nvim", "init.lua")},
		},
		{
			name: "adds multiple files",
			setup: func(t *testing.T, home, repo string) []string {
				writeFile(t, filepath.Join(home, ".zshrc"))
				writeFile(t, filepath.Join(home, ".gitconfig"))
				return []string{
					filepath.Join(home, ".zshrc"),
					filepath.Join(home, ".gitconfig"),
				}
			},
			wantRels: []string{".zshrc", ".gitconfig"},
		},
		{
			name: "overwrites existing file in repo",
			setup: func(t *testing.T, home, repo string) []string {
				writeFile(t, filepath.Join(home, ".zshrc"))
				require.NoError(t, os.MkdirAll(repo, 0o755))
				require.NoError(t, os.WriteFile(filepath.Join(repo, ".zshrc"), []byte("old"), 0o600))
				return []string{filepath.Join(home, ".zshrc")}
			},
			wantRels: []string{".zshrc"},
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
				writeFile(t, outside)
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
				writeFile(t, filepath.Join(home, ".zshrc"))
				return []string{
					filepath.Join(home, ".zshrc"),
					filepath.Join(home, "missing.txt"),
				}
			},
			wantRels: []string{".zshrc"},
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
				dst := filepath.Join(repo, rel)

				got, err := os.ReadFile(dst)
				require.NoError(t, err)
				assert.Equal(t, wantContent, string(got))

				info, err := os.Stat(dst)
				require.NoError(t, err)
				assert.Equal(t, wantPerm, info.Mode().Perm())
			}
		})
	}
}
