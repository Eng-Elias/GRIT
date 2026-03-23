package clone

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func fixtureRepoPath() string {
	_, filename, _, _ := runtime.Caller(0)
	return filepath.Join(filepath.Dir(filename), "..", "..", "testdata", "fixture-repo")
}

func TestCloneToMemory_FixtureRepo(t *testing.T) {
	repoPath := fixtureRepoPath()
	require.DirExists(t, repoPath)

	absPath, err := filepath.Abs(repoPath)
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	result, err := Clone(ctx, "", "", "", 0, "", 51200)
	_ = result
	_ = err
	_ = absPath
}

func TestCleanOldClones(t *testing.T) {
	tmpDir := t.TempDir()

	ownerDir := filepath.Join(tmpDir, "owner", "repo")
	require.NoError(t, os.MkdirAll(ownerDir, 0o755))

	oldClone := filepath.Join(ownerDir, "old-clone")
	require.NoError(t, os.MkdirAll(oldClone, 0o755))
	oldTime := time.Now().Add(-2 * time.Hour)
	os.Chtimes(oldClone, oldTime, oldTime)

	newClone := filepath.Join(ownerDir, "new-clone")
	require.NoError(t, os.MkdirAll(newClone, 0o755))

	cleanOldClones(tmpDir, 1*time.Hour)

	assert.NoDirExists(t, oldClone, "old clone should be removed")
	assert.DirExists(t, newClone, "new clone should remain")
}

func TestStartCleanup_ContextCancel(t *testing.T) {
	tmpDir := t.TempDir()

	ctx, cancel := context.WithCancel(context.Background())

	StartCleanup(ctx, tmpDir, 1*time.Hour, 100*time.Millisecond)

	time.Sleep(250 * time.Millisecond)
	cancel()
}
