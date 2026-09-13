package cmd

import (
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
