package cmd

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_dirs(t *testing.T) {
	homeEnv := "HOME"

	tests := []struct {
		name     string
		home     string
		dotDir   string
		wantRepo func(home string) string
		hasErr   bool
	}{
		{
			name:   "falls back to home/dotfiles when DOT_DIR is empty",
			home:   "/tmp/testhome",
			dotDir: "",
			wantRepo: func(home string) string {
				return filepath.Join(home, "dotfiles")
			},
			hasErr: false,
		},
		{
			name:   "uses DOT_DIR when set",
			home:   "/tmp/testhome",
			dotDir: "/tmp/custom/dot",
			wantRepo: func(home string) string {
				return "/tmp/custom/dot"
			},
			hasErr: false,
		},
		{
			name:     "fails when home directory is unavailable",
			home:     "",
			dotDir:   "/tmp/custom/dot",
			wantRepo: nil,
			hasErr:   true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Setenv(homeEnv, test.home)
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
			assert.Equal(t, test.wantRepo(test.home), repo)
		})
	}
}

func Test_copyFile(t *testing.T) {
	t.Parallel()
	const wantContent = "hello, world\n"
	const wantPerm = os.FileMode(0o640)

	writeSrc := func(t *testing.T, path string) {
		t.Helper()
		require.NoError(t, os.MkdirAll(filepath.Dir(path), 0o755))
		require.NoError(t, os.WriteFile(path, []byte(wantContent), wantPerm))
		require.NoError(t, os.Chmod(path, wantPerm))
	}

	tests := []struct {
		name   string
		setup  func(t *testing.T, dir string) (src, dst string)
		hasErr bool
	}{
		{
			name: "copies file when parent directory already exists",
			setup: func(t *testing.T, dir string) (string, string) {
				src := filepath.Join(dir, "src.txt")
				writeSrc(t, src)
				return src, filepath.Join(dir, "dst.txt")
			},
			hasErr: false,
		},
		{
			name: "copies file when parent directory does not exist",
			setup: func(t *testing.T, dir string) (string, string) {
				src := filepath.Join(dir, "src.txt")
				writeSrc(t, src)
				return src, filepath.Join(dir, "out", "dst.txt")
			},
			hasErr: false,
		},
		{
			name: "copies file when parent directory needs nested creation",
			setup: func(t *testing.T, dir string) (string, string) {
				src := filepath.Join(dir, "src.txt")
				writeSrc(t, src)
				return src, filepath.Join(dir, "a", "b", "dst.txt")
			},
			hasErr: false,
		},
		{
			name: "overwrites existing destination file",
			setup: func(t *testing.T, dir string) (string, string) {
				src := filepath.Join(dir, "src.txt")
				writeSrc(t, src)
				dst := filepath.Join(dir, "dst.txt")
				require.NoError(t, os.WriteFile(dst, []byte("old content"), 0o600))
				require.NoError(t, os.Chmod(dst, 0o600))
				return src, filepath.Join(dir, "a", "b", "dst.txt")
			},
			hasErr: false,
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

			got, err := os.ReadFile(dst)
			require.NoError(t, err)
			assert.Equal(t, wantContent, string(got))

			info, err := os.Stat(dst)
			require.NoError(t, err)
			assert.Equal(t, wantPerm, info.Mode().Perm())

			srcData, err := os.ReadFile(src)
			require.NoError(t, err)
			assert.Equal(t, wantContent, string(srcData))
		})
	}
}

func Test_managedFiles(t *testing.T) {
	t.Parallel()

	mkFiles := func(t *testing.T, repo string, rels ...string) {
		t.Helper()
		for _, rel := range rels {
			path := filepath.Join(repo, filepath.FromSlash(rel))
			require.NoError(t, os.MkdirAll(filepath.Dir(path), 0o755))
			require.NoError(t, os.WriteFile(path, []byte("x"), 0o644))
		}
	}

	tests := []struct {
		name   string
		setup  func(t *testing.T) string
		want   []string
		hasErr bool
	}{
		{
			name: "returns files sorted",
			setup: func(t *testing.T) string {
				repo := t.TempDir()
				mkFiles(t, repo, ".zshrc", ".gitconfig", ".vimrc")
				return repo
			},
			want: []string{".gitconfig", ".vimrc", ".zshrc"},
		},
		{
			name: "returns nested files with relative paths",
			setup: func(t *testing.T) string {
				repo := t.TempDir()
				mkFiles(t, repo, ".config/nvim/init.lua", ".config/git/ignore", ".zshrc")
				return repo
			},
			want: []string{".config/git/ignore", ".config/nvim/init.lua", ".zshrc"},
		},
		{
			name: "skips .git directory",
			setup: func(t *testing.T) string {
				repo := t.TempDir()
				mkFiles(t, repo, ".zshrc", ".git/config", ".git/objects/ab/cdef")
				return repo
			},
			want: []string{".zshrc"},
		},
		{
			name: "skips nested .git directory",
			setup: func(t *testing.T) string {
				repo := t.TempDir()
				mkFiles(t, repo, ".config/sub/.git/config", ".config/sub/file.txt")
				return repo
			},
			want: []string{".config/sub/file.txt"},
		},
		{
			name: "does not skip .git when it is a file",
			setup: func(t *testing.T) string {
				repo := t.TempDir()
				mkFiles(t, repo, ".git", ".zshrc")
				return repo
			},
			want: []string{".git", ".zshrc"},
		},
		{
			name: "returns nil when repo has only directories",
			setup: func(t *testing.T) string {
				repo := t.TempDir()
				require.NoError(t, os.MkdirAll(filepath.Join(repo, ".config", "nvim"), 0o755))
				return repo
			},
			want: nil,
		},
		{
			name: "returns nil when repo is empty",
			setup: func(t *testing.T) string {
				return t.TempDir()
			},
			want: nil,
		},
		{
			name: "fails when repo does not exist",
			setup: func(t *testing.T) string {
				return filepath.Join(t.TempDir(), "missing")
			},
			hasErr: true,
		},
		{
			name: "returns dot when repo is a regular file",
			setup: func(t *testing.T) string {
				dir := t.TempDir()
				path := filepath.Join(dir, "notadir")
				require.NoError(t, os.WriteFile(path, []byte("x"), 0o644))
				return path
			},
			want: []string{"."},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			repo := test.setup(t)

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

	mkFile := func(t *testing.T, path, content string) {
		t.Helper()
		require.NoError(t, os.WriteFile(path, []byte(content), 0o644))
	}

	tests := []struct {
		name   string
		setup  func(t *testing.T) (a, b string)
		want   bool
		hasErr bool
	}{
		{
			name: "returns false when contents are same",
			setup: func(t *testing.T) (string, string) {
				dir := t.TempDir()
				a, b := filepath.Join(dir, "a"), filepath.Join(dir, "b")
				mkFile(t, a, "same")
				mkFile(t, b, "same")
				return a, b
			},
			want: false,
		},
		{
			name: "returns true when contents differ",
			setup: func(t *testing.T) (string, string) {
				dir := t.TempDir()
				a, b := filepath.Join(dir, "a"), filepath.Join(dir, "b")
				mkFile(t, a, "new")
				mkFile(t, b, "old")
				return a, b
			},
			want: true,
		},
		{
			name: "returns true when b does not exist",
			setup: func(t *testing.T) (string, string) {
				dir := t.TempDir()
				a := filepath.Join(dir, "a")
				mkFile(t, a, "x")
				return a, filepath.Join(dir, "missing")
			},
			want: true,
		},
		{
			name: "fails when a does not exist",
			setup: func(t *testing.T) (string, string) {
				dir := t.TempDir()
				b := filepath.Join(dir, "b")
				mkFile(t, b, "x")
				return filepath.Join(dir, "missing"), b
			},
			hasErr: true,
		},
		{
			name: "fails when b is a directory",
			setup: func(t *testing.T) (string, string) {
				dir := t.TempDir()
				a := filepath.Join(dir, "a")
				mkFile(t, a, "x")
				return a, t.TempDir()
			},
			hasErr: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			a, b := test.setup(t)

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
