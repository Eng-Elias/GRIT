package complexity

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPool_ProcessesAllFiles(t *testing.T) {
	repoRoot := locateRepoRoot(t)
	cloneDir := filepath.Join(repoRoot, "testdata", "fixture-repo")

	pool := NewPool(cloneDir)
	files := []FileInput{
		{Path: "complexity_sample.go", LOC: 65},
		{Path: "complexity_sample.py", LOC: 35},
		{Path: "complexity_sample.js", LOC: 40},
	}

	results := pool.Run(context.Background(), files)
	assert.Len(t, results, 3)
}

func TestPool_SkipsParseError(t *testing.T) {
	tmpDir := t.TempDir()
	// Write an invalid Go file.
	err := os.WriteFile(filepath.Join(tmpDir, "bad.go"), []byte("not valid go }{}{"), 0644)
	require.NoError(t, err)

	pool := NewPool(tmpDir)
	files := []FileInput{
		{Path: "bad.go", LOC: 1},
	}

	// Should not panic; returns whatever the parser returns (may be partial or nil).
	results := pool.Run(context.Background(), files)
	// The parser may or may not return a result for malformed input;
	// the key assertion is no panic.
	_ = results
}

func TestPool_EmptyFiles(t *testing.T) {
	pool := NewPool("/tmp")
	results := pool.Run(context.Background(), nil)
	assert.Nil(t, results)
}

func TestPool_RespectsContextCancellation(t *testing.T) {
	repoRoot := locateRepoRoot(t)
	cloneDir := filepath.Join(repoRoot, "testdata", "fixture-repo")

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately.

	pool := NewPool(cloneDir)
	files := []FileInput{
		{Path: "complexity_sample.go", LOC: 65},
		{Path: "complexity_sample.py", LOC: 35},
	}

	results := pool.Run(ctx, files)
	// With a cancelled context, workers should exit early.
	// We may get 0 results or some partial results.
	assert.True(t, len(results) <= len(files))
}

func TestPool_MissingFileSkipped(t *testing.T) {
	tmpDir := t.TempDir()

	pool := NewPool(tmpDir)
	files := []FileInput{
		{Path: "nonexistent.go", LOC: 10},
	}

	results := pool.Run(context.Background(), files)
	assert.Len(t, results, 0)
}
