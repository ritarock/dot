package cmd

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_sync(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		act       action
		repoFiles map[string]string // nil means the repo directory does not exist
		homeFiles map[string]string
		yes       bool
		stdin     string
		wantRepo  map[string]string
		wantHome  map[string]string
		hasErr    bool
	}{
		{
			name:      "apply copies changed file to home when yes",
			act:       apply,
			repoFiles: map[string]string{".zshrc": "new"},
			homeFiles: map[string]string{".zshrc": "old"},
			yes:       true,
			wantRepo:  map[string]string{".zshrc": "new"},
			wantHome:  map[string]string{".zshrc": "new"},
		},
		{
			name:      "apply creates nested file missing in home",
			act:       apply,
			repoFiles: map[string]string{".config/nvim/init.lua": "lua"},
			yes:       true,
			wantRepo:  map[string]string{".config/nvim/init.lua": "lua"},
			wantHome:  map[string]string{".config/nvim/init.lua": "lua"},
		},
		{
			name:      "apply copies when confirmed with y",
			act:       apply,
			repoFiles: map[string]string{".zshrc": "new"},
			homeFiles: map[string]string{".zshrc": "old"},
			stdin:     "y\n",
			wantRepo:  map[string]string{".zshrc": "new"},
			wantHome:  map[string]string{".zshrc": "new"},
		},
		{
			name:      "apply does nothing when confirmation is declined",
			act:       apply,
			repoFiles: map[string]string{".zshrc": "new"},
			homeFiles: map[string]string{".zshrc": "old"},
			stdin:     "n\n",
			wantRepo:  map[string]string{".zshrc": "new"},
			wantHome:  map[string]string{".zshrc": "old"},
		},
		{
			name:      "apply does nothing when stdin is empty",
			act:       apply,
			repoFiles: map[string]string{".zshrc": "new"},
			homeFiles: map[string]string{".zshrc": "old"},
			stdin:     "",
			wantRepo:  map[string]string{".zshrc": "new"},
			wantHome:  map[string]string{".zshrc": "old"},
		},
		{
			name:      "apply does nothing when already updated",
			act:       apply,
			repoFiles: map[string]string{".zshrc": "same"},
			homeFiles: map[string]string{".zshrc": "same"},
			wantRepo:  map[string]string{".zshrc": "same"},
			wantHome:  map[string]string{".zshrc": "same"},
		},
		{
			name:      "apply ignores files only in home",
			act:       apply,
			repoFiles: map[string]string{".zshrc": "new"},
			homeFiles: map[string]string{".zshrc": "old", ".bashrc": "bash"},
			yes:       true,
			wantRepo:  map[string]string{".zshrc": "new"},
			wantHome:  map[string]string{".zshrc": "new", ".bashrc": "bash"},
		},
		{
			name:      "update copies changed file to repo without confirmation",
			act:       update,
			repoFiles: map[string]string{".zshrc": "old"},
			homeFiles: map[string]string{".zshrc": "new"},
			stdin:     "",
			wantRepo:  map[string]string{".zshrc": "new"},
			wantHome:  map[string]string{".zshrc": "new"},
		},
		{
			name:      "update skips file missing in home",
			act:       update,
			repoFiles: map[string]string{".zshrc": "x", ".vimrc": "old"},
			homeFiles: map[string]string{".vimrc": "new"},
			wantRepo:  map[string]string{".zshrc": "x", ".vimrc": "new"},
			wantHome:  map[string]string{".vimrc": "new"},
		},
		{
			name:      "update does nothing when already updated",
			act:       update,
			repoFiles: map[string]string{".zshrc": "same"},
			homeFiles: map[string]string{".zshrc": "same"},
			wantRepo:  map[string]string{".zshrc": "same"},
			wantHome:  map[string]string{".zshrc": "same"},
		},
		{
			name:   "apply fails when repo does not exist",
			act:    apply,
			yes:    true,
			hasErr: true,
		},
		{
			name:   "update fails when repo does not exist",
			act:    update,
			hasErr: true,
		},
		{
			name:      "apply fails when home has directory with same name",
			act:       apply,
			repoFiles: map[string]string{".config": "x"},
			homeFiles: map[string]string{".config/nvim/init.lua": "lua"},
			yes:       true,
			hasErr:    true,
		},
		{
			name:      "update fails when home has directory with same name",
			act:       update,
			repoFiles: map[string]string{".config": "x"},
			homeFiles: map[string]string{".config/nvim/init.lua": "lua"},
			hasErr:    true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			repo := filepath.Join(t.TempDir(), "repo")
			home := t.TempDir()
			writeFiles(t, repo, test.repoFiles)
			writeFiles(t, home, test.homeFiles)

			err := sync(repo, home, test.yes, strings.NewReader(test.stdin), test.act)

			if test.hasErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, test.wantRepo, readFiles(t, repo))
			assert.Equal(t, test.wantHome, readFiles(t, home))
		})
	}
}

func Test_confirm(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		stdin string
		want  bool
	}{
		{
			name:  "returns true for y",
			stdin: "y\n",
			want:  true,
		},
		{
			name:  "returns true for yes",
			stdin: "yes\n",
			want:  true,
		},
		{
			name:  "returns true for uppercase Y",
			stdin: "Y\n",
			want:  true,
		},
		{
			name:  "returns true for uppercase YES",
			stdin: "YES\n",
			want:  true,
		},
		{
			name:  "returns true for y surrounded by spaces",
			stdin: "  y  \n",
			want:  true,
		},
		{
			name:  "returns true for y without newline",
			stdin: "y",
			want:  true,
		},
		{
			name:  "returns false for n",
			stdin: "n\n",
			want:  false,
		},
		{
			name:  "returns false for no",
			stdin: "no\n",
			want:  false,
		},
		{
			name:  "returns false for empty line",
			stdin: "\n",
			want:  false,
		},
		{
			name:  "returns false when stdin is empty",
			stdin: "",
			want:  false,
		},
		{
			name:  "returns false for other input",
			stdin: "yy\n",
			want:  false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			got := confirm("prompt", strings.NewReader(test.stdin))

			assert.Equal(t, test.want, got)
		})
	}
}
