package core

import (
	"path/filepath"
	"runtime"
	"testing"

	"github.com/go-git/go-git/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func fixtureRepoPath() string {
	_, filename, _, _ := runtime.Caller(0)
	return filepath.Join(filepath.Dir(filename), "..", "..", "..", "testdata", "fixture-repo")
}

func TestWalk_FixtureRepo(t *testing.T) {
	repoPath := fixtureRepoPath()

	repo, err := git.PlainOpen(repoPath)
	require.NoError(t, err, "fixture repo must exist at %s", repoPath)

	files, err := Walk(repo)
	require.NoError(t, err)

	assert.GreaterOrEqual(t, len(files), 7, "fixture has at least 7 non-lock text files")

	langsSeen := make(map[string]bool)
	for _, f := range files {
		langsSeen[f.Language] = true
		assert.NotEmpty(t, f.Path)
		assert.NotEmpty(t, f.Language)
	}

	assert.True(t, langsSeen["Go"], "should detect Go files")
	assert.True(t, langsSeen["Python"], "should detect Python files")
	assert.True(t, langsSeen["JavaScript"], "should detect JavaScript files")
}

func TestWalk_BinarySkipping(t *testing.T) {
	repoPath := fixtureRepoPath()

	repo, err := git.PlainOpen(repoPath)
	require.NoError(t, err)

	files, err := Walk(repo)
	require.NoError(t, err)

	for _, f := range files {
		ext := filepath.Ext(f.Path)
		assert.NotEqual(t, ".png", ext, "should skip binary files")
		assert.NotEqual(t, ".exe", ext, "should skip binary files")
	}
}

func TestWalk_LineCounts(t *testing.T) {
	repoPath := fixtureRepoPath()

	repo, err := git.PlainOpen(repoPath)
	require.NoError(t, err)

	files, err := Walk(repo)
	require.NoError(t, err)

	for _, f := range files {
		if f.Language == "Go" || f.Language == "Python" || f.Language == "JavaScript" {
			assert.Greater(t, f.TotalLines, 0, "text files should have line counts: %s", f.Path)
			assert.Greater(t, f.CodeLines, 0, "code files should have code lines: %s", f.Path)
		}
	}
}
