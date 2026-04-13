package blame

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/grit-app/grit/internal/models"
)

func TestAggregate_EmailCaseDedup(t *testing.T) {
	now := time.Now()
	results := map[string]*FileBlameResult{
		"main.go": {
			Path:     "main.go",
			Language: "Go",
			Lines: []LineBlameDatum{
				{AuthorName: "Alice", AuthorEmail: "alice@example.com", Date: now},
				{AuthorName: "ALICE", AuthorEmail: "alice@example.com", Date: now.Add(-time.Hour)},
				{AuthorName: "Bob", AuthorEmail: "bob@example.com", Date: now},
			},
		},
	}

	agg := Aggregate(results)
	require.Len(t, agg.Authors, 2)
	assert.Equal(t, 3, agg.TotalLines)

	// Alice owns 2 lines, Bob owns 1.
	assert.Equal(t, "alice@example.com", agg.Authors[0].Email)
	assert.Equal(t, 2, agg.Authors[0].TotalLinesOwned)
	assert.Equal(t, "bob@example.com", agg.Authors[1].Email)
	assert.Equal(t, 1, agg.Authors[1].TotalLinesOwned)
}

func TestAggregate_MostRecentNameResolution(t *testing.T) {
	old := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	recent := time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC)

	results := map[string]*FileBlameResult{
		"main.go": {
			Path:     "main.go",
			Language: "Go",
			Lines: []LineBlameDatum{
				{AuthorName: "Old Name", AuthorEmail: "alice@example.com", Date: old},
				{AuthorName: "New Name", AuthorEmail: "alice@example.com", Date: recent},
			},
		},
	}

	agg := Aggregate(results)
	require.Len(t, agg.Authors, 1)
	assert.Equal(t, "New Name", agg.Authors[0].Name)
}

func TestAggregate_OwnershipPercentSumsTo100(t *testing.T) {
	now := time.Now()
	results := map[string]*FileBlameResult{
		"main.go": {
			Path:     "main.go",
			Language: "Go",
			Lines: []LineBlameDatum{
				{AuthorName: "Alice", AuthorEmail: "alice@example.com", Date: now},
				{AuthorName: "Alice", AuthorEmail: "alice@example.com", Date: now},
				{AuthorName: "Alice", AuthorEmail: "alice@example.com", Date: now},
				{AuthorName: "Bob", AuthorEmail: "bob@example.com", Date: now},
				{AuthorName: "Bob", AuthorEmail: "bob@example.com", Date: now},
				{AuthorName: "Carol", AuthorEmail: "carol@example.com", Date: now},
				{AuthorName: "Carol", AuthorEmail: "carol@example.com", Date: now},
				{AuthorName: "Carol", AuthorEmail: "carol@example.com", Date: now},
				{AuthorName: "Carol", AuthorEmail: "carol@example.com", Date: now},
				{AuthorName: "Dave", AuthorEmail: "dave@example.com", Date: now},
			},
		},
	}

	agg := Aggregate(results)
	var total float64
	for _, a := range agg.Authors {
		total += a.OwnershipPercent
	}
	assert.InDelta(t, 100.0, total, 0.5)
}

func TestAggregate_PrimaryLanguagesTop3(t *testing.T) {
	now := time.Now()
	results := map[string]*FileBlameResult{
		"main.go": {
			Path: "main.go", Language: "Go",
			Lines: []LineBlameDatum{
				{AuthorName: "Alice", AuthorEmail: "alice@example.com", Date: now},
				{AuthorName: "Alice", AuthorEmail: "alice@example.com", Date: now},
				{AuthorName: "Alice", AuthorEmail: "alice@example.com", Date: now},
				{AuthorName: "Alice", AuthorEmail: "alice@example.com", Date: now},
				{AuthorName: "Alice", AuthorEmail: "alice@example.com", Date: now},
			},
		},
		"app.py": {
			Path: "app.py", Language: "Python",
			Lines: []LineBlameDatum{
				{AuthorName: "Alice", AuthorEmail: "alice@example.com", Date: now},
				{AuthorName: "Alice", AuthorEmail: "alice@example.com", Date: now},
				{AuthorName: "Alice", AuthorEmail: "alice@example.com", Date: now},
			},
		},
		"index.js": {
			Path: "index.js", Language: "JavaScript",
			Lines: []LineBlameDatum{
				{AuthorName: "Alice", AuthorEmail: "alice@example.com", Date: now},
				{AuthorName: "Alice", AuthorEmail: "alice@example.com", Date: now},
			},
		},
		"Main.java": {
			Path: "Main.java", Language: "Java",
			Lines: []LineBlameDatum{
				{AuthorName: "Alice", AuthorEmail: "alice@example.com", Date: now},
			},
		},
	}

	agg := Aggregate(results)
	require.Len(t, agg.Authors, 1)
	assert.Len(t, agg.Authors[0].PrimaryLanguages, 3)
	assert.Equal(t, "Go", agg.Authors[0].PrimaryLanguages[0])
	assert.Equal(t, "Python", agg.Authors[0].PrimaryLanguages[1])
	assert.Equal(t, "JavaScript", agg.Authors[0].PrimaryLanguages[2])
}

func TestAggregate_IsActiveFlag(t *testing.T) {
	recent := time.Now().Add(-30 * 24 * time.Hour) // 1 month ago — active
	old := time.Now().Add(-365 * 24 * time.Hour)   // 1 year ago — inactive

	results := map[string]*FileBlameResult{
		"main.go": {
			Path: "main.go", Language: "Go",
			Lines: []LineBlameDatum{
				{AuthorName: "Active", AuthorEmail: "active@example.com", Date: recent},
				{AuthorName: "Inactive", AuthorEmail: "inactive@example.com", Date: old},
			},
		},
	}

	agg := Aggregate(results)
	require.Len(t, agg.Authors, 2)

	for _, a := range agg.Authors {
		if a.Email == "active@example.com" {
			assert.True(t, a.IsActive, "recent author should be active")
		}
		if a.Email == "inactive@example.com" {
			assert.False(t, a.IsActive, "old author should be inactive")
		}
	}
}

func TestAggregate_FilesTouched(t *testing.T) {
	now := time.Now()
	results := map[string]*FileBlameResult{
		"a.go": {
			Path: "a.go", Language: "Go",
			Lines: []LineBlameDatum{
				{AuthorName: "Alice", AuthorEmail: "alice@example.com", Date: now},
			},
		},
		"b.go": {
			Path: "b.go", Language: "Go",
			Lines: []LineBlameDatum{
				{AuthorName: "Alice", AuthorEmail: "alice@example.com", Date: now},
			},
		},
		"c.py": {
			Path: "c.py", Language: "Python",
			Lines: []LineBlameDatum{
				{AuthorName: "Alice", AuthorEmail: "alice@example.com", Date: now},
				{AuthorName: "Bob", AuthorEmail: "bob@example.com", Date: now},
			},
		},
	}

	agg := Aggregate(results)
	for _, a := range agg.Authors {
		if a.Email == "alice@example.com" {
			assert.Equal(t, 3, a.FilesTouched)
		}
		if a.Email == "bob@example.com" {
			assert.Equal(t, 1, a.FilesTouched)
		}
	}
}

func TestAggregate_FirstLastCommitDate(t *testing.T) {
	d1 := time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)
	d2 := time.Date(2025, 6, 1, 0, 0, 0, 0, time.UTC)
	d3 := time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC)

	results := map[string]*FileBlameResult{
		"main.go": {
			Path: "main.go", Language: "Go",
			Lines: []LineBlameDatum{
				{AuthorName: "Alice", AuthorEmail: "alice@example.com", Date: d2},
				{AuthorName: "Alice", AuthorEmail: "alice@example.com", Date: d1},
				{AuthorName: "Alice", AuthorEmail: "alice@example.com", Date: d3},
			},
		},
	}

	agg := Aggregate(results)
	require.Len(t, agg.Authors, 1)
	assert.Equal(t, d1, agg.Authors[0].FirstCommitDate)
	assert.Equal(t, d3, agg.Authors[0].LastCommitDate)
}

func TestAggregate_EmptyInput(t *testing.T) {
	agg := Aggregate(map[string]*FileBlameResult{})
	assert.Empty(t, agg.Authors)
	assert.Equal(t, 0, agg.TotalLines)
}

func TestAggregate_SortedByLinesOwnedDesc(t *testing.T) {
	now := time.Now()
	results := map[string]*FileBlameResult{
		"main.go": {
			Path: "main.go", Language: "Go",
			Lines: []LineBlameDatum{
				{AuthorName: "Carol", AuthorEmail: "carol@example.com", Date: now},
				{AuthorName: "Alice", AuthorEmail: "alice@example.com", Date: now},
				{AuthorName: "Alice", AuthorEmail: "alice@example.com", Date: now},
				{AuthorName: "Alice", AuthorEmail: "alice@example.com", Date: now},
				{AuthorName: "Bob", AuthorEmail: "bob@example.com", Date: now},
				{AuthorName: "Bob", AuthorEmail: "bob@example.com", Date: now},
			},
		},
	}

	agg := Aggregate(results)
	require.Len(t, agg.Authors, 3)
	assert.Equal(t, "alice@example.com", agg.Authors[0].Email)
	assert.Equal(t, "bob@example.com", agg.Authors[1].Email)
	assert.Equal(t, "carol@example.com", agg.Authors[2].Email)
}

// ---- T012: Bus Factor Tests ----

func TestComputeBusFactor_SingleDominant(t *testing.T) {
	authors := makeAuthors([]float64{85, 10, 5})
	bf, keyPeople := ComputeBusFactor(authors)
	assert.Equal(t, 1, bf)
	require.Len(t, keyPeople, 1)
	assert.Equal(t, 85.0, keyPeople[0].OwnershipPercent)
}

func TestComputeBusFactor_ThreeWaySplit(t *testing.T) {
	authors := makeAuthors([]float64{30, 28, 25, 10, 7})
	bf, keyPeople := ComputeBusFactor(authors)
	assert.Equal(t, 3, bf)
	require.Len(t, keyPeople, 3)
}

func TestComputeBusFactor_EqualDistribution(t *testing.T) {
	pcts := make([]float64, 10)
	for i := range pcts {
		pcts[i] = 10
	}
	authors := makeAuthors(pcts)
	bf, keyPeople := ComputeBusFactor(authors)
	assert.Equal(t, 8, bf)
	require.Len(t, keyPeople, 8)
}

func TestComputeBusFactor_Empty(t *testing.T) {
	bf, keyPeople := ComputeBusFactor(nil)
	assert.Equal(t, 0, bf)
	assert.Nil(t, keyPeople)
}

func TestComputeBusFactor_SingleAuthor(t *testing.T) {
	authors := makeAuthors([]float64{100})
	bf, keyPeople := ComputeBusFactor(authors)
	assert.Equal(t, 1, bf)
	require.Len(t, keyPeople, 1)
}

// ---- T014: FileContributors Tests ----

func TestComputeFileContributors_Top3Only(t *testing.T) {
	now := time.Now()
	lines := make([]LineBlameDatum, 0)
	// 5 authors: 50, 20, 15, 10, 5 lines
	for i := 0; i < 50; i++ {
		lines = append(lines, LineBlameDatum{AuthorName: "A1", AuthorEmail: "a1@x.com", Date: now})
	}
	for i := 0; i < 20; i++ {
		lines = append(lines, LineBlameDatum{AuthorName: "A2", AuthorEmail: "a2@x.com", Date: now})
	}
	for i := 0; i < 15; i++ {
		lines = append(lines, LineBlameDatum{AuthorName: "A3", AuthorEmail: "a3@x.com", Date: now})
	}
	for i := 0; i < 10; i++ {
		lines = append(lines, LineBlameDatum{AuthorName: "A4", AuthorEmail: "a4@x.com", Date: now})
	}
	for i := 0; i < 5; i++ {
		lines = append(lines, LineBlameDatum{AuthorName: "A5", AuthorEmail: "a5@x.com", Date: now})
	}

	results := map[string]*FileBlameResult{
		"main.go": {Path: "main.go", Language: "Go", Lines: lines},
	}

	fc := ComputeFileContributors(results)
	require.Len(t, fc, 1)
	assert.Len(t, fc[0].TopAuthors, 3)
	assert.Equal(t, 100, fc[0].TotalLines)
	assert.Equal(t, "a1@x.com", fc[0].TopAuthors[0].Email)
	assert.Equal(t, 50, fc[0].TopAuthors[0].LinesOwned)
}

func TestComputeFileContributors_SingleAuthor100Percent(t *testing.T) {
	now := time.Now()
	results := map[string]*FileBlameResult{
		"main.go": {
			Path: "main.go", Language: "Go",
			Lines: []LineBlameDatum{
				{AuthorName: "Alice", AuthorEmail: "alice@x.com", Date: now},
				{AuthorName: "Alice", AuthorEmail: "alice@x.com", Date: now},
			},
		},
	}

	fc := ComputeFileContributors(results)
	require.Len(t, fc, 1)
	require.Len(t, fc[0].TopAuthors, 1)
	assert.Equal(t, 100.0, fc[0].TopAuthors[0].OwnershipPercent)
}

func TestComputeFileContributors_TwoAuthors(t *testing.T) {
	now := time.Now()
	results := map[string]*FileBlameResult{
		"main.go": {
			Path: "main.go", Language: "Go",
			Lines: []LineBlameDatum{
				{AuthorName: "Alice", AuthorEmail: "alice@x.com", Date: now},
				{AuthorName: "Alice", AuthorEmail: "alice@x.com", Date: now},
				{AuthorName: "Alice", AuthorEmail: "alice@x.com", Date: now},
				{AuthorName: "Bob", AuthorEmail: "bob@x.com", Date: now},
			},
		},
	}

	fc := ComputeFileContributors(results)
	require.Len(t, fc, 1)
	require.Len(t, fc[0].TopAuthors, 2)
	assert.Equal(t, "alice@x.com", fc[0].TopAuthors[0].Email)
	assert.Equal(t, 75.0, fc[0].TopAuthors[0].OwnershipPercent)
	assert.Equal(t, "bob@x.com", fc[0].TopAuthors[1].Email)
	assert.Equal(t, 25.0, fc[0].TopAuthors[1].OwnershipPercent)
}

func TestComputeFileContributors_EmptyFileSkipped(t *testing.T) {
	results := map[string]*FileBlameResult{
		"empty.go": {Path: "empty.go", Language: "Go", Lines: []LineBlameDatum{}},
	}
	fc := ComputeFileContributors(results)
	assert.Empty(t, fc)
}

// helper to make Author slices from ownership percentages.
func makeAuthors(pcts []float64) []models.Author {
	authors := make([]models.Author, len(pcts))
	for i, p := range pcts {
		authors[i] = models.Author{
			Email:            fmt.Sprintf("author%d@example.com", i),
			Name:             fmt.Sprintf("Author %d", i),
			OwnershipPercent: p,
			TotalLinesOwned:  int(p * 10),
		}
	}
	return authors
}
