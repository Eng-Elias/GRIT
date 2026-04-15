package temporal

import (
	"testing"
	"time"

	"github.com/grit-app/grit/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBucketToISOWeek(t *testing.T) {
	// Monday 2026-04-13 is ISO week 2026-W16
	monday := time.Date(2026, 4, 13, 0, 0, 0, 0, time.UTC)
	year, week := monday.ISOWeek()
	assert.Equal(t, 2026, year)
	assert.Equal(t, 16, week)

	label := isoWeekLabel(year, week)
	assert.Equal(t, "2026-W16", label)
}

func TestISOWeekLabel(t *testing.T) {
	tests := []struct {
		year int
		week int
		want string
	}{
		{2026, 1, "2026-W01"},
		{2026, 52, "2026-W52"},
		{2025, 5, "2025-W05"},
	}
	for _, tt := range tests {
		assert.Equal(t, tt.want, isoWeekLabel(tt.year, tt.week))
	}
}

func TestWeekStartFromISOWeek(t *testing.T) {
	// ISO week 2026-W16 starts on Monday 2026-04-13
	ws := weekStartFromISO(2026, 16)
	assert.Equal(t, time.Monday, ws.Weekday())
	assert.Equal(t, 2026, ws.Year())
	assert.Equal(t, time.April, ws.Month())
	assert.Equal(t, 13, ws.Day())
}

func TestComputeCommitCadence(t *testing.T) {
	// Create 8 weeks of data: weeks with varying commit counts
	weeks := []models.WeeklyActivity{
		{Week: "2026-W09", WeekStart: time.Date(2026, 2, 23, 0, 0, 0, 0, time.UTC), Commits: 7},
		{Week: "2026-W10", WeekStart: time.Date(2026, 3, 2, 0, 0, 0, 0, time.UTC), Commits: 14},
		{Week: "2026-W11", WeekStart: time.Date(2026, 3, 9, 0, 0, 0, 0, time.UTC), Commits: 7},
		{Week: "2026-W12", WeekStart: time.Date(2026, 3, 16, 0, 0, 0, 0, time.UTC), Commits: 21},
		{Week: "2026-W13", WeekStart: time.Date(2026, 3, 23, 0, 0, 0, 0, time.UTC), Commits: 7},
		{Week: "2026-W14", WeekStart: time.Date(2026, 3, 30, 0, 0, 0, 0, time.UTC), Commits: 14},
		{Week: "2026-W15", WeekStart: time.Date(2026, 4, 6, 0, 0, 0, 0, time.UTC), Commits: 7},
		{Week: "2026-W16", WeekStart: time.Date(2026, 4, 13, 0, 0, 0, 0, time.UTC), Commits: 21},
	}

	cadence := computeCommitCadence(weeks)
	require.NotEmpty(t, cadence)

	// First window: W09-W12, total = 7+14+7+21 = 49 commits, 28 days → 49/28 ≈ 1.75
	assert.InDelta(t, 49.0/28.0, cadence[0].CommitsPerDay, 0.01)
}

func TestComputeCommitCadenceTooFewWeeks(t *testing.T) {
	weeks := []models.WeeklyActivity{
		{Week: "2026-W15", WeekStart: time.Date(2026, 4, 6, 0, 0, 0, 0, time.UTC), Commits: 10},
	}
	cadence := computeCommitCadence(weeks)
	assert.Empty(t, cadence)
}

func TestRankAuthors(t *testing.T) {
	authorMap := map[string]*authorAccum{
		"alice@ex.com": {name: "Alice", additions: 500, deletions: 200, weeks: map[weekKey]weekAccum{
			{2026, 15}: {additions: 500, deletions: 200, commits: 5},
		}},
		"bob@ex.com": {name: "Bob", additions: 100, deletions: 50, weeks: map[weekKey]weekAccum{
			{2026, 15}: {additions: 100, deletions: 50, commits: 2},
		}},
	}

	result := rankAuthors(authorMap, 10)
	require.Len(t, result, 2)
	assert.Equal(t, "alice@ex.com", result[0].Email)
	assert.Equal(t, 500, result[0].TotalAdditions)
	assert.Equal(t, "bob@ex.com", result[1].Email)
}

func TestRankAuthorsTopN(t *testing.T) {
	authorMap := make(map[string]*authorAccum)
	for i := 0; i < 15; i++ {
		email := string(rune('a'+i)) + "@ex.com"
		authorMap[email] = &authorAccum{
			name:      string(rune('A' + i)),
			additions: (15 - i) * 100,
			deletions: 0,
			weeks:     map[weekKey]weekAccum{},
		}
	}

	result := rankAuthors(authorMap, 10)
	assert.Len(t, result, 10)
}
