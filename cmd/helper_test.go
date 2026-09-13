package cmd

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func writeFile(t *testing.T, path, content string, perm os.FileMode) {
	t.Helper()
	require.NoError(t, os.MkdirAll(filepath.Dir(path), 0o755))
	require.NoError(t, os.WriteFile(path, []byte(content), perm))
	require.NoError(t, os.Chmod(path, perm))
}

func writeFiles(t *testing.T, dir string, files map[string]string) {
	t.Helper()
	for rel, content := range files {
		writeFile(t, filepath.Join(dir, filepath.FromSlash(rel)), content, 0o644)
	}
}

func assertFile(t *testing.T, path, wantContent string, wantPerm os.FileMode) {
	t.Helper()
	got, err := os.ReadFile(path)
	require.NoError(t, err)
	assert.Equal(t, wantContent, string(got))

	info, err := os.Stat(path)
	require.NoError(t, err)
	assert.Equal(t, wantPerm, info.Mode().Perm())
}
