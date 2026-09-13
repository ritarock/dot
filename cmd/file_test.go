package cmd

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_dirs(t *testing.T) {
	tests := []struct {
		name     string
		home     string
		dotDir   string
		wantRepo string
		hasErr   bool
	}{
		{
			name:     "falls back to home/dotfiles when DOT_DIR is empty",
			home:     "/tmp/testhome",
			dotDir:   "",
			wantRepo: filepath.Join("/tmp/testhome", "dotfiles"),
		},
		{
			name:     "uses DOT_DIR when set",
			home:     "/tmp/testhome",
			dotDir:   "/tmp/custom/dot",
			wantRepo: "/tmp/custom/dot",
		},
		{
			name:   "fails when home directory is unavailable",
			home:   "",
			dotDir: "/tmp/custom/dot",
			hasErr: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Setenv("HOME", test.home)
			t.Setenv("DOT_DIR", test.dotDir)

			repo, home, err := dirs()

			if test.hasErr {
				assert.Error(t, err)
				assert.Empty(t, repo)
				assert.Empty(t, home)
				return
			}
			require.NoError(t, err)

			assert.Equal(t, test.home, home)
			assert.Equal(t, test.wantRepo, repo)
		})
	}
}

func Test_copyFile(t *testing.T) {
	t.Parallel()
	const wantContent = "hello, world\n"
	const wantPerm = os.FileMode(0o640)

	tests := []struct {
		name   string
		setup  func(t *testing.T, dir string) (src, dst string)
		hasErr bool
	}{
		{
			name: "copies file when parent directory already exists",
			setup: func(t *testing.T, dir string) (string, string) {
				src := filepath.Join(dir, "src.txt")
				writeFile(t, src, wantContent, wantPerm)
				return src, filepath.Join(dir, "dst.txt")
			},
		},
		{
			name: "copies file when parent directory does not exist",
			setup: func(t *testing.T, dir string) (string, string) {
				src := filepath.Join(dir, "src.txt")
				writeFile(t, src, wantContent, wantPerm)
				return src, filepath.Join(dir, "out", "dst.txt")
			},
		},
		{
			name: "copies file when parent directory needs nested creation",
			setup: func(t *testing.T, dir string) (string, string) {
				src := filepath.Join(dir, "src.txt")
				writeFile(t, src, wantContent, wantPerm)
				return src, filepath.Join(dir, "a", "b", "dst.txt")
			},
		},
		{
			name: "overwrites existing destination file",
			setup: func(t *testing.T, dir string) (string, string) {
				src := filepath.Join(dir, "src.txt")
				writeFile(t, src, wantContent, wantPerm)
				dst := filepath.Join(dir, "dst.txt")
				writeFile(t, dst, "old content", 0o600)
				return src, dst
			},
		},
		{
			name: "fails when source does not exist",
			setup: func(t *testing.T, dir string) (string, string) {
				return filepath.Join(dir, "missing.txt"), filepath.Join(dir, "out", "dst.txt")
			},
			hasErr: true,
		},
		{
			name: "fails when source is a directory",
			setup: func(t *testing.T, dir string) (string, string) {
				src := filepath.Join(dir, "srcdir")
				require.NoError(t, os.Mkdir(src, 0o755))
				return src, filepath.Join(dir, "out", "dst.txt")
			},
			hasErr: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			dir := t.TempDir()
			src, dst := test.setup(t, dir)

			err := copyFile(src, dst)

			if test.hasErr {
				assert.Error(t, err)
				assert.NoFileExists(t, dst)
				return
			}
			require.NoError(t, err)

			assertFile(t, dst, wantContent, wantPerm)
			assertFile(t, src, wantContent, wantPerm)
		})
	}
}

func Test_managedFiles(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		setup  func(t *testing.T, dir string) string
		want   []string
		hasErr bool
	}{
		{
			name: "returns files sorted",
			setup: func(t *testing.T, dir string) string {
				writeFiles(t, dir, map[string]string{".zshrc": "x", ".gitconfig": "x", ".vimrc": "x"})
				return dir
			},
			want: []string{".gitconfig", ".vimrc", ".zshrc"},
		},
		{
			name: "returns nested files with relative paths",
			setup: func(t *testing.T, dir string) string {
				writeFiles(t, dir, map[string]string{".config/nvim/init.lua": "x", ".config/git/ignore": "x", ".zshrc": "x"})
				return dir
			},
			want: []string{".config/git/ignore", ".config/nvim/init.lua", ".zshrc"},
		},
		{
			name: "skips .git directory",
			setup: func(t *testing.T, dir string) string {
				writeFiles(t, dir, map[string]string{".zshrc": "x", ".git/config": "x", ".git/objects/ab/cdef": "x"})
				return dir
			},
			want: []string{".zshrc"},
		},
		{
			name: "skips nested .git directory",
			setup: func(t *testing.T, dir string) string {
				writeFiles(t, dir, map[string]string{".config/sub/.git/config": "x", ".config/sub/file.txt": "x"})
				return dir
			},
			want: []string{".config/sub/file.txt"},
		},
		{
			name: "does not skip .git when it is a file",
			setup: func(t *testing.T, dir string) string {
				writeFiles(t, dir, map[string]string{".git": "x", ".zshrc": "x"})
				return dir
			},
			want: []string{".git", ".zshrc"},
		},
		{
			name: "returns nil when repo has only directories",
			setup: func(t *testing.T, dir string) string {
				require.NoError(t, os.MkdirAll(filepath.Join(dir, ".config", "nvim"), 0o755))
				return dir
			},
			want: nil,
		},
		{
			name: "returns nil when repo is empty",
			setup: func(t *testing.T, dir string) string {
				return dir
			},
			want: nil,
		},
		{
			name: "fails when repo does not exist",
			setup: func(t *testing.T, dir string) string {
				return filepath.Join(dir, "missing")
			},
			hasErr: true,
		},
		{
			name: "returns dot when repo is a regular file",
			setup: func(t *testing.T, dir string) string {
				path := filepath.Join(dir, "notadir")
				writeFile(t, path, "x", 0o644)
				return path
			},
			want: []string{"."},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			repo := test.setup(t, t.TempDir())

			got, err := managedFiles(repo)

			if test.hasErr {
				assert.Error(t, err)
				assert.Nil(t, got)
				return
			}
			require.NoError(t, err)

			var gotSlash []string
			for _, p := range got {
				gotSlash = append(gotSlash, filepath.ToSlash(p))
			}
			assert.Equal(t, test.want, gotSlash)
		})
	}
}

func Test_filesDiffer(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		setup  func(t *testing.T, dir string) (a, b string)
		want   bool
		hasErr bool
	}{
		{
			name: "returns false when contents are same",
			setup: func(t *testing.T, dir string) (string, string) {
				a, b := filepath.Join(dir, "a"), filepath.Join(dir, "b")
				writeFile(t, a, "same", 0o644)
				writeFile(t, b, "same", 0o644)
				return a, b
			},
			want: false,
		},
		{
			name: "returns true when contents differ",
			setup: func(t *testing.T, dir string) (string, string) {
				a, b := filepath.Join(dir, "a"), filepath.Join(dir, "b")
				writeFile(t, a, "new", 0o644)
				writeFile(t, b, "old", 0o644)
				return a, b
			},
			want: true,
		},
		{
			name: "returns true when b does not exist",
			setup: func(t *testing.T, dir string) (string, string) {
				a := filepath.Join(dir, "a")
				writeFile(t, a, "x", 0o644)
				return a, filepath.Join(dir, "missing")
			},
			want: true,
		},
		{
			name: "fails when a does not exist",
			setup: func(t *testing.T, dir string) (string, string) {
				b := filepath.Join(dir, "b")
				writeFile(t, b, "x", 0o644)
				return filepath.Join(dir, "missing"), b
			},
			hasErr: true,
		},
		{
			name: "fails when b is a directory",
			setup: func(t *testing.T, dir string) (string, string) {
				a := filepath.Join(dir, "a")
				writeFile(t, a, "x", 0o644)
				return a, dir
			},
			hasErr: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			a, b := test.setup(t, t.TempDir())

			got, err := filesDiffer(a, b)

			if test.hasErr {
				assert.Error(t, err)
				assert.False(t, got)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, test.want, got)
		})
	}
}
