package cmd

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

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
				writeFiles(t, dir, map[string]string{"dot_zshrc": "x", "dot_gitconfig": "x", "dot_vimrc": "x"})
				return dir
			},
			want: []string{"dot_gitconfig", "dot_vimrc", "dot_zshrc"},
		},
		{
			name: "returns nested files with relative paths",
			setup: func(t *testing.T, dir string) string {
				writeFiles(t, dir, map[string]string{"dot_config/nvim/init.lua": "x", "dot_config/git/ignore": "x", "dot_zshrc": "x"})
				return dir
			},
			want: []string{"dot_config/git/ignore", "dot_config/nvim/init.lua", "dot_zshrc"},
		},
		{
			name: "skips dot-prefixed directories",
			setup: func(t *testing.T, dir string) string {
				writeFiles(t, dir, map[string]string{"dot_zshrc": "x", ".git/config": "x", ".git/objects/ab/cdef": "x", ".github/workflows/ci.yml": "x"})
				return dir
			},
			want: []string{"dot_zshrc"},
		},
		{
			name: "skips nested dot-prefixed directory",
			setup: func(t *testing.T, dir string) string {
				writeFiles(t, dir, map[string]string{"dot_config/sub/.git/config": "x", "dot_config/sub/file.txt": "x"})
				return dir
			},
			want: []string{"dot_config/sub/file.txt"},
		},
		{
			name: "skips dot-prefixed files",
			setup: func(t *testing.T, dir string) string {
				writeFiles(t, dir, map[string]string{".git": "x", ".gitignore": "x", "dot_zshrc": "x"})
				return dir
			},
			want: []string{"dot_zshrc"},
		},
		{
			name: "skips runner.yaml at repo root",
			setup: func(t *testing.T, dir string) string {
				writeFiles(t, dir, map[string]string{runnerFile: "x", "dot_zshrc": "x"})
				return dir
			},
			want: []string{"dot_zshrc"},
		},
		{
			name: "does not skip runner.yaml in a subdirectory",
			setup: func(t *testing.T, dir string) string {
				writeFiles(t, dir, map[string]string{"dot_config/" + runnerFile: "x", "dot_zshrc": "x"})
				return dir
			},
			want: []string{"dot_config/" + runnerFile, "dot_zshrc"},
		},
		{
			name: "does not skip repo root starting with dot",
			setup: func(t *testing.T, dir string) string {
				repo := filepath.Join(dir, ".dotfiles")
				writeFiles(t, repo, map[string]string{"dot_zshrc": "x"})
				return repo
			},
			want: []string{"dot_zshrc"},
		},
		{
			name: "returns nil when repo has only directories",
			setup: func(t *testing.T, dir string) string {
				require.NoError(t, os.MkdirAll(filepath.Join(dir, "dot_config", "nvim"), 0o755))
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

func Test_encodeRel(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		rel  string
		want string
	}{
		{name: "returns empty for empty path", rel: "", want: ""},
		{name: "leaves non-dot file unchanged", rel: "zshrc", want: "zshrc"},
		{name: "encodes dot file", rel: ".zshrc", want: "dot_zshrc"},
		{name: "encodes only leading dot directory", rel: ".config/nvim/init.lua", want: "dot_config/nvim/init.lua"},
		{name: "encodes every dot element", rel: ".config/.hidden/.x", want: "dot_config/dot_hidden/dot_x"},
		{name: "ignores dot in the middle of element", rel: "a.b/c.txt", want: "a.b/c.txt"},
		{name: "leaves dot unchanged", rel: ".", want: "."},
		{name: "leaves dot dot unchanged", rel: "..", want: ".."},
		{name: "keeps dot dot and encodes following element", rel: "../.zshrc", want: "../dot_zshrc"},
		{name: "strips only one leading dot", rel: "..foo", want: "dot_.foo"},
		{name: "leaves element already prefixed with dot_ unchanged", rel: "dot_foo", want: "dot_foo"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			got := encodeRel(filepath.FromSlash(test.rel))

			assert.Equal(t, filepath.FromSlash(test.want), got)
		})
	}
}

func Test_decodeRel(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		rel  string
		want string
	}{
		{name: "returns empty for empty path", rel: "", want: ""},
		{name: "leaves non-prefixed file unchanged", rel: "zshrc", want: "zshrc"},
		{name: "decodes prefixed file", rel: "dot_zshrc", want: ".zshrc"},
		{name: "decodes only prefixed directory", rel: "dot_config/nvim/init.lua", want: ".config/nvim/init.lua"},
		{name: "decodes every prefixed element", rel: "dot_config/dot_hidden/dot_x", want: ".config/.hidden/.x"},
		{name: "ignores prefix not at element start", rel: "mydot_file", want: "mydot_file"},
		{name: "decodes bare prefix to dot", rel: "dot_", want: "."},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			got := decodeRel(filepath.FromSlash(test.rel))

			assert.Equal(t, filepath.FromSlash(test.want), got)
		})
	}
}

func Test_encodeDecodeRel_roundTrip(t *testing.T) {
	t.Parallel()

	tests := []string{
		".zshrc",
		".config/nvim/init.lua",
		".config/.hidden/.x",
		"a.b/c.txt",
		"../.zshrc",
		"..foo",
	}

	for _, rel := range tests {
		t.Run(rel, func(t *testing.T) {
			t.Parallel()

			rel := filepath.FromSlash(rel)

			assert.Equal(t, rel, decodeRel(encodeRel(rel)))
		})
	}
}
