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
