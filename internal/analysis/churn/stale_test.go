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

	"github.com/grit-app/grit/internal/models"
)

func TestIsSourceExtension(t *testing.T) {
	tests := []struct {
		path string
		want bool
	}{
		{"main.go", true},
		{"app.py", true},
		{"index.js", true},
		{"lib.ts", true},
		{"component.tsx", true},
		{"Main.java", true},
		{"script.sh", true},
		{"query.sql", true},
		{"schema.proto", true},
		{"README.md", false},
		{"config.yaml", false},
		{"data.json", false},
		{"go.mod", false},
		{"go.sum", false},
		{"package-lock.lock", false},
		{"image.png", false},
		{"font.woff", false},
		{"notes.txt", false},
		{"noext", false},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			assert.Equal(t, tt.want, IsSourceExtension(tt.path))
		})
	}
}

func TestMonthsInactive(t *testing.T) {
	now := time.Date(2026, 4, 1, 0, 0, 0, 0, time.UTC)

	tests := []struct {
		name         string
		lastModified time.Time
		want         int
	}{
		{"6 months ago", now.AddDate(0, -6, 0), 6},
		{"12 months ago", now.AddDate(-1, 0, 0), 12},
		{"1 day ago", now.AddDate(0, 0, -1), 0},
		{"zero time", time.Time{}, 0},
		{"exactly 30 days", now.AddDate(0, 0, -30), 1},
		{"exactly 60 days", now.AddDate(0, 0, -60), 2},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := MonthsInactive(tt.lastModified, now)
			assert.Equal(t, tt.want, got)
		})
	}
}

// createStaleTestRepo creates a repo with:
// - main.go (recent commit, references "unused")
// - unused.go (old commit, source ext, NOT referenced → stale)
// - old_config.yaml (old commit, non-source ext → excluded)
// - referenced.go (old commit, source ext, referenced by main.go → not stale)
func createStaleTestRepo(t *testing.T) *git.Repository {
	t.Helper()
	dir := t.TempDir()

	repo, err := git.PlainInit(dir, false)
	require.NoError(t, err)

	wt, err := repo.Worktree()
	require.NoError(t, err)

	oldTime := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC) // ~15 months ago from Apr 2026

	// Commit 1: add unused.go and referenced.go and old_config.yaml (old)
	require.NoError(t, os.WriteFile(filepath.Join(dir, "unused.go"), []byte("package unused\n"), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "referenced.go"), []byte("package referenced\n"), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "old_config.yaml"), []byte("key: value\n"), 0644))
	_, err = wt.Add("unused.go")
	require.NoError(t, err)
	_, err = wt.Add("referenced.go")
	require.NoError(t, err)
	_, err = wt.Add("old_config.yaml")
	require.NoError(t, err)
	_, err = wt.Commit("old commit", &git.CommitOptions{
		Author: &object.Signature{Name: "test", Email: "t@t.com", When: oldTime},
	})
	require.NoError(t, err)

	// Commit 2: add main.go that references "referenced" (recent)
	mainContent := "package main\nimport \"referenced\"\nfunc main() {}\n"
	require.NoError(t, os.WriteFile(filepath.Join(dir, "main.go"), []byte(mainContent), 0644))
	_, err = wt.Add("main.go")
	require.NoError(t, err)
	_, err = wt.Commit("recent commit", &git.CommitOptions{
		Author: &object.Signature{Name: "test", Email: "t@t.com", When: time.Date(2026, 3, 15, 0, 0, 0, 0, time.UTC)},
	})
	require.NoError(t, err)

	return repo
}

func TestFindStaleFiles_ThreeConditions(t *testing.T) {
	repo := createStaleTestRepo(t)
	now := time.Date(2026, 4, 1, 0, 0, 0, 0, time.UTC)
	oldTime := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)

	files := []models.FileChurn{
		{Path: "main.go", Churn: 1, LastModified: time.Date(2026, 3, 15, 0, 0, 0, 0, time.UTC)},
		{Path: "unused.go", Churn: 1, LastModified: oldTime},
		{Path: "referenced.go", Churn: 1, LastModified: oldTime},
		{Path: "old_config.yaml", Churn: 1, LastModified: oldTime},
	}

	stale := FindStaleFiles(files, repo, now)

	// Only unused.go should be stale:
	// - main.go: recent → excluded by condition 1
	// - old_config.yaml: non-source ext → excluded by condition 2
	// - referenced.go: referenced by main.go → excluded by condition 3
	// - unused.go: old + source + not referenced → STALE
	assert.Len(t, stale, 1)
	assert.Equal(t, "unused.go", stale[0].Path)
	assert.GreaterOrEqual(t, stale[0].MonthsInactive, 6)
}

func TestFindStaleFiles_Empty(t *testing.T) {
	stale := FindStaleFiles(nil, nil, time.Now())
	assert.Nil(t, stale)
}

func TestFindStaleFiles_AllRecent(t *testing.T) {
	now := time.Now()
	files := []models.FileChurn{
		{Path: "a.go", Churn: 5, LastModified: now.AddDate(0, -1, 0)},
		{Path: "b.py", Churn: 3, LastModified: now.AddDate(0, -2, 0)},
	}
	stale := FindStaleFiles(files, nil, now)
	assert.Nil(t, stale)
}
