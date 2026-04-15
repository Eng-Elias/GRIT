package temporal

import (
	"sort"
	"testing"
	"time"

	"github.com/grit-app/grit/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMonthBoundaries(t *testing.T) {
	now := time.Date(2026, 4, 15, 0, 0, 0, 0, time.UTC)

	tests := []struct {
		name     string
		period   string
		now      time.Time
		wantLen  int
		wantFirst time.Time
		wantLast  time.Time
	}{
		{
			name:      "3y period produces 36 boundaries",
			period:    "3y",
			now:       now,
			wantLen:   36,
			wantFirst: time.Date(2023, 5, 1, 0, 0, 0, 0, time.UTC),
			wantLast:  time.Date(2026, 4, 1, 0, 0, 0, 0, time.UTC),
		},
		{
			name:      "2y period produces 24 boundaries",
			period:    "2y",
			now:       now,
			wantLen:   24,
			wantFirst: time.Date(2024, 5, 1, 0, 0, 0, 0, time.UTC),
			wantLast:  time.Date(2026, 4, 1, 0, 0, 0, 0, time.UTC),
		},
		{
			name:      "1y period produces 12 boundaries",
			period:    "1y",
			now:       now,
			wantLen:   12,
			wantFirst: time.Date(2025, 5, 1, 0, 0, 0, 0, time.UTC),
			wantLast:  time.Date(2026, 4, 1, 0, 0, 0, 0, time.UTC),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			boundaries := monthBoundaries(tt.now, tt.period)
			require.Len(t, boundaries, tt.wantLen)
			assert.Equal(t, tt.wantFirst, boundaries[0])
			assert.Equal(t, tt.wantLast, boundaries[len(boundaries)-1])

			// Verify chronological order
			for i := 1; i < len(boundaries); i++ {
				assert.True(t, boundaries[i].After(boundaries[i-1]))
			}
		})
	}
}

func TestTopLanguages(t *testing.T) {
	tests := []struct {
		name      string
		langMap   map[string]int
		wantLen   int
		wantOther int
	}{
		{
			name: "10 or fewer languages, no Other",
			langMap: map[string]int{
				"Go":     500,
				"Python": 300,
				"Rust":   200,
			},
			wantLen:   3,
			wantOther: 0,
		},
		{
			name: "more than 10 languages groups rest as Other",
			langMap: map[string]int{
				"Go": 500, "Python": 400, "Rust": 300, "Java": 200,
				"TypeScript": 150, "C": 120, "C++": 100, "Ruby": 80,
				"PHP": 60, "Perl": 40, "Lua": 20, "Dart": 10,
			},
			wantLen:   11, // top 10 + Other
			wantOther: 30, // Lua(20) + Dart(10)
		},
		{
			name:      "empty map",
			langMap:   map[string]int{},
			wantLen:   0,
			wantOther: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := topLanguages(tt.langMap, 10)
			assert.Len(t, result, tt.wantLen)
			if tt.wantOther > 0 {
				assert.Equal(t, tt.wantOther, result["Other"])
			} else {
				_, hasOther := result["Other"]
				assert.False(t, hasOther)
			}
		})
	}
}

func TestSnapshotsAreSortedChronologically(t *testing.T) {
	snapshots := []models.MonthlySnapshot{
		{Date: time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC), TotalLOC: 300},
		{Date: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC), TotalLOC: 100},
		{Date: time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC), TotalLOC: 200},
	}

	sort.Slice(snapshots, func(i, j int) bool {
		return snapshots[i].Date.Before(snapshots[j].Date)
	})

	assert.Equal(t, 100, snapshots[0].TotalLOC)
	assert.Equal(t, 200, snapshots[1].TotalLOC)
	assert.Equal(t, 300, snapshots[2].TotalLOC)
}

func TestParsePeriodMonths(t *testing.T) {
	tests := []struct {
		period string
		want   int
	}{
		{"1y", 12},
		{"2y", 24},
		{"3y", 36},
		{"invalid", 36},
		{"", 36},
	}

	for _, tt := range tests {
		t.Run(tt.period, func(t *testing.T) {
			assert.Equal(t, tt.want, parsePeriodMonths(tt.period))
		})
	}
}
