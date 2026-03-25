package core

import (
	"testing"

	"github.com/go-git/go-git/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIntegration_WalkAndCount_FixtureRepo(t *testing.T) {
	repoPath := fixtureRepoPath()

	repo, err := git.PlainOpen(repoPath)
	require.NoError(t, err, "fixture repo must exist at %s", repoPath)

	files, err := Walk(repo)
	require.NoError(t, err)

	require.GreaterOrEqual(t, len(files), 7, "fixture has at least 7 text files")

	languages := aggregateLanguages(files)
	require.NotEmpty(t, languages)

	langMap := make(map[string]int)
	for _, lb := range languages {
		langMap[lb.Language] = lb.TotalLines
	}

	assert.Contains(t, langMap, "Go", "should detect Go language")
	assert.Contains(t, langMap, "Python", "should detect Python language")
	assert.Contains(t, langMap, "JavaScript", "should detect JavaScript language")
	assert.Contains(t, langMap, "CSS", "should detect CSS language")
	assert.Contains(t, langMap, "YAML", "should detect YAML language")

	var totalLines, totalCode, totalComment, totalBlank int
	for _, f := range files {
		totalLines += f.TotalLines
		totalCode += f.CodeLines
		totalComment += f.CommentLines
		totalBlank += f.BlankLines
	}

	assert.Greater(t, totalLines, 50, "fixture repo should have >50 total lines")
	assert.Greater(t, totalCode, 20, "fixture repo should have >20 code lines")
	assert.Greater(t, totalComment, 5, "fixture repo should have >5 comment lines")
	assert.Greater(t, totalBlank, 3, "fixture repo should have >3 blank lines")

	assert.Equal(t, totalLines, totalCode+totalComment+totalBlank,
		"total = code + comment + blank")

	filePathMap := make(map[string]bool)
	for _, f := range files {
		filePathMap[f.Path] = true
		assert.NotEmpty(t, f.Language)
		assert.Greater(t, f.ByteSize, int64(0))
	}

	assert.True(t, filePathMap["main.go"])
	assert.True(t, filePathMap["utils.go"])
	assert.True(t, filePathMap["app.py"])
	assert.True(t, filePathMap["helpers.py"])
	assert.True(t, filePathMap["index.js"])
	assert.True(t, filePathMap["style.css"])
	assert.True(t, filePathMap["config.yaml"])

	var totalPercentage float64
	for _, lb := range languages {
		assert.Greater(t, lb.FileCount, 0)
		assert.Greater(t, lb.Percentage, 0.0)
		totalPercentage += lb.Percentage
	}
	assert.InDelta(t, 100.0, totalPercentage, 0.5, "language percentages should sum to ~100")
}

func TestIntegration_GoFileComments(t *testing.T) {
	repoPath := fixtureRepoPath()

	repo, err := git.PlainOpen(repoPath)
	require.NoError(t, err)

	files, err := Walk(repo)
	require.NoError(t, err)

	for _, f := range files {
		if f.Path == "main.go" {
			assert.Greater(t, f.CommentLines, 0, "main.go has // and /* */ comments")
			assert.Greater(t, f.CodeLines, 0, "main.go has code lines")
			assert.Greater(t, f.BlankLines, 0, "main.go has blank lines")
			return
		}
	}
	t.Fatal("main.go not found in fixture repo")
}

func TestIntegration_PythonFileComments(t *testing.T) {
	repoPath := fixtureRepoPath()

	repo, err := git.PlainOpen(repoPath)
	require.NoError(t, err)

	files, err := Walk(repo)
	require.NoError(t, err)

	for _, f := range files {
		if f.Path == "app.py" {
			assert.Greater(t, f.CommentLines, 0, "app.py has # and \"\"\" comments")
			assert.Greater(t, f.CodeLines, 0, "app.py has code lines")
			return
		}
	}
	t.Fatal("app.py not found in fixture repo")
}
