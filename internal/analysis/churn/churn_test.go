package churn

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// helper: create a temp git repo with N commits, each touching "file.txt".
func createTestRepo(t *testing.T, commitCount int) *git.Repository {
	t.Helper()
	dir := t.TempDir()

	repo, err := git.PlainInit(dir, false)
	require.NoError(t, err)

	wt, err := repo.Worktree()
	require.NoError(t, err)

	for i := 0; i < commitCount; i++ {
		fname := "file.txt"
		if i%3 == 0 {
			fname = "other.go"
		}
		fpath := filepath.Join(dir, fname)
		err := os.WriteFile(fpath, []byte("line "+string(rune('A'+i%26))+"\n"), 0644)
		require.NoError(t, err)

		_, err = wt.Add(fname)
		require.NoError(t, err)

		_, err = wt.Commit("commit "+string(rune('A'+i%26)), &git.CommitOptions{
			Author: &object.Signature{
				Name:  "test",
				Email: "test@test.com",
				When:  time.Now().Add(-time.Duration(commitCount-i) * 24 * time.Hour),
			},
		})
		require.NoError(t, err)
	}

	return repo
}

func TestWalkCommitLog_BasicCounting(t *testing.T) {
	repo := createTestRepo(t, 10)

	result, err := WalkCommitLog(repo)
	require.NoError(t, err)

	assert.Equal(t, 10, result.TotalCommits)
	assert.NotEmpty(t, result.Files)

	// file.txt is modified in commits where i%3 != 0 → 7 times (i=1,2,4,5,7,8, and one more)
	// other.go is modified in commits where i%3 == 0 → 4 times (i=0,3,6,9)
	churnMap := make(map[string]int)
	for _, f := range result.Files {
		churnMap[f.Path] = f.Churn
	}
	assert.Contains(t, churnMap, "file.txt")
	assert.Contains(t, churnMap, "other.go")
	assert.Equal(t, 10, churnMap["file.txt"]+churnMap["other.go"])
}

func TestWalkCommitLog_EmptyRepo(t *testing.T) {
	dir := t.TempDir()
	repo, err := git.PlainInit(dir, false)
	require.NoError(t, err)

	result, err := WalkCommitLog(repo)
	require.NoError(t, err)
	assert.Empty(t, result.Files)
	assert.Equal(t, 0, result.TotalCommits)
}

func TestWalkCommitLog_SortedByChurnDescending(t *testing.T) {
	repo := createTestRepo(t, 20)

	result, err := WalkCommitLog(repo)
	require.NoError(t, err)

	for i := 1; i < len(result.Files); i++ {
		assert.GreaterOrEqual(t, result.Files[i-1].Churn, result.Files[i].Churn,
			"files should be sorted by churn descending")
	}
}

func TestWalkCommitLog_LastModifiedTracked(t *testing.T) {
	repo := createTestRepo(t, 5)

	result, err := WalkCommitLog(repo)
	require.NoError(t, err)

	for _, f := range result.Files {
		assert.False(t, f.LastModified.IsZero(), "last_modified should be set for %s", f.Path)
	}
}

func TestWalkCommitLog_WindowDates(t *testing.T) {
	repo := createTestRepo(t, 5)

	result, err := WalkCommitLog(repo)
	require.NoError(t, err)

	assert.False(t, result.CommitWindowStart.IsZero())
	assert.False(t, result.CommitWindowEnd.IsZero())
	assert.True(t, result.CommitWindowEnd.After(result.CommitWindowStart) || result.CommitWindowEnd.Equal(result.CommitWindowStart))
}
