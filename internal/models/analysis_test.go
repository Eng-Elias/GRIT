package models

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAnalysisResult_JSONSerialization(t *testing.T) {
	now := time.Date(2026, 3, 23, 12, 0, 0, 0, time.UTC)
	cachedAt := time.Date(2026, 3, 23, 12, 0, 0, 0, time.UTC)

	result := AnalysisResult{
		Repository: Repository{
			Owner:         "facebook",
			Name:          "react",
			FullName:      "facebook/react",
			DefaultBranch: "main",
			LatestSHA:     "abc123",
		},
		Metadata: GitHubMetadata{
			Name:            "react",
			Description:     "A JavaScript library",
			Stars:           230000,
			PrimaryLanguage: "JavaScript",
			LicenseName:     "MIT License",
			LicenseSPDX:     "MIT",
			Topics:          []string{"react", "javascript"},
		},
		CommitActivity: CommitActivity{
			WeeklyCounts: []int{10, 20, 30},
			TotalCommits: 18500,
		},
		Files: []FileStats{
			{
				Path:         "src/index.js",
				Language:     "JavaScript",
				TotalLines:   100,
				CodeLines:    80,
				CommentLines: 10,
				BlankLines:   10,
				ByteSize:     3000,
			},
		},
		Languages: []LanguageBreakdown{
			{
				Language:     "JavaScript",
				FileCount:    1,
				TotalLines:   100,
				CodeLines:    80,
				CommentLines: 10,
				BlankLines:   10,
				Percentage:   100.0,
			},
		},
		TotalFiles:        1,
		TotalLines:        100,
		TotalCodeLines:    80,
		TotalCommentLines: 10,
		TotalBlankLines:   10,
		AnalyzedAt:        now,
		Cached:            true,
		CachedAt:          &cachedAt,
	}

	data, err := json.Marshal(result)
	require.NoError(t, err)

	var m map[string]interface{}
	err = json.Unmarshal(data, &m)
	require.NoError(t, err)

	assert.Equal(t, float64(100), m["total_lines"])
	assert.Equal(t, float64(80), m["total_code_lines"])
	assert.Equal(t, float64(10), m["total_comment_lines"])
	assert.Equal(t, float64(10), m["total_blank_lines"])
	assert.Equal(t, true, m["cached"])
	assert.NotNil(t, m["cached_at"])

	repo := m["repository"].(map[string]interface{})
	assert.Equal(t, "facebook", repo["owner"])
	assert.Equal(t, "react", repo["name"])
	assert.Equal(t, "facebook/react", repo["full_name"])
}

func TestFileStats_JSONKeys(t *testing.T) {
	fs := FileStats{
		Path:         "main.go",
		Language:     "Go",
		TotalLines:   50,
		CodeLines:    40,
		CommentLines: 5,
		BlankLines:   5,
		ByteSize:     1500,
	}

	data, err := json.Marshal(fs)
	require.NoError(t, err)

	var m map[string]interface{}
	err = json.Unmarshal(data, &m)
	require.NoError(t, err)

	assert.Contains(t, m, "path")
	assert.Contains(t, m, "language")
	assert.Contains(t, m, "total_lines")
	assert.Contains(t, m, "code_lines")
	assert.Contains(t, m, "comment_lines")
	assert.Contains(t, m, "blank_lines")
	assert.Contains(t, m, "byte_size")
}

func TestLanguageBreakdown_JSONKeys(t *testing.T) {
	lb := LanguageBreakdown{
		Language:     "Go",
		FileCount:    10,
		TotalLines:   500,
		CodeLines:    400,
		CommentLines: 50,
		BlankLines:   50,
		Percentage:   65.5,
	}

	data, err := json.Marshal(lb)
	require.NoError(t, err)

	var m map[string]interface{}
	err = json.Unmarshal(data, &m)
	require.NoError(t, err)

	assert.Contains(t, m, "language")
	assert.Contains(t, m, "file_count")
	assert.Contains(t, m, "percentage")
	assert.Equal(t, 65.5, m["percentage"])
}

func TestAnalysisResult_NullCachedAt(t *testing.T) {
	result := AnalysisResult{
		Cached:   false,
		CachedAt: nil,
	}

	data, err := json.Marshal(result)
	require.NoError(t, err)
	assert.Contains(t, string(data), `"cached_at":null`)
}
