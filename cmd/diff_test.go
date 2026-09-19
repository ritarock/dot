package cmd

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_cmdDiff(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		repoFiles map[string]string // nil means the repo directory does not exist
		homeFiles map[string]string
		hasErr    bool
	}{
		{
			name:      "succeeds when nothing changed",
			repoFiles: map[string]string{"dot_zshrc": "same"},
			homeFiles: map[string]string{".zshrc": "same"},
		},
		{
			name:      "succeeds when file differs",
			repoFiles: map[string]string{"dot_zshrc": "new"},
			homeFiles: map[string]string{".zshrc": "old"},
		},
		{
			name:      "succeeds when file is missing in home",
			repoFiles: map[string]string{"dot_config/nvim/init.lua": "lua"},
		},
		{
			name:      "ignores files only in home",
			repoFiles: map[string]string{"dot_zshrc": "same"},
			homeFiles: map[string]string{".zshrc": "same", ".bashrc": "bash"},
		},
		{
			name:   "fails when repo does not exist",
			hasErr: true,
		},
		{
			name:      "fails when home has directory with same name",
			repoFiles: map[string]string{"dot_config": "x"},
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
