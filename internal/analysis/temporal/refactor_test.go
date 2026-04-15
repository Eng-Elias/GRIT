package temporal

import (
	"testing"
	"time"

	"github.com/grit-app/grit/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func makeWeek(weekLabel string, start time.Time, adds, dels, commits int) models.WeeklyActivity {
	return models.WeeklyActivity{
		Week:      weekLabel,
		WeekStart: start,
		Additions: adds,
		Deletions: dels,
		Commits:   commits,
	}
}

func TestDetectRefactorPeriods_SinglePeriod(t *testing.T) {
	// commits = [5,5,15,12,14], sorted = [5,5,12,14,15], median = 12
	// Weeks with negative net LOC AND commits > 12: W03(15), W04→no(net+), W05(14)
	// Actually: W03 net -150 commits 15>12 ✓, W04 net -70 commits 14>12 ✓, W05 net -60 commits 13>12 ✓
	weeks := []models.WeeklyActivity{
		makeWeek("2026-W01", time.Date(2026, 1, 5, 0, 0, 0, 0, time.UTC), 100, 50, 5),   // net +50
		makeWeek("2026-W02", time.Date(2026, 1, 12, 0, 0, 0, 0, time.UTC), 100, 50, 5),  // net +50
		makeWeek("2026-W03", time.Date(2026, 1, 19, 0, 0, 0, 0, time.UTC), 50, 200, 15), // net -150, commits>median
		makeWeek("2026-W04", time.Date(2026, 1, 26, 0, 0, 0, 0, time.UTC), 30, 100, 14), // net -70, commits>median
		makeWeek("2026-W05", time.Date(2026, 2, 2, 0, 0, 0, 0, time.UTC), 20, 80, 13),   // net -60, commits>median
	}
	// sorted commits = [5,5,13,14,15], median = 13. >13 means 14,15 qualify. Fix: use >= or adjust.
	// Actually median=13, commits>13: W03(15)✓, W04(14)✓, W05(13)✗
	// Let's make W05 commits=14 so all three qualify.
	weeks[4] = makeWeek("2026-W05", time.Date(2026, 2, 2, 0, 0, 0, 0, time.UTC), 20, 80, 14)
	// sorted commits = [5,5,14,14,15], median = 14. >14 means only W03(15). That's wrong too.
	// Use: [3,3,15,15,15] → median=15, none >15. Bad.
	// Use: [5,5,20,20,20] → sorted=[5,5,20,20,20], median=20, >20: none.
	// Strategy: make low-commit weeks very low so median is low.
	// [2,2,15,14,13] → sorted=[2,2,13,14,15], median=13, >13: W03(15), W04(14). Two qualify.
	// Let's go with exactly 2 consecutive refactor weeks.
	weeks = []models.WeeklyActivity{
		makeWeek("2026-W01", time.Date(2026, 1, 5, 0, 0, 0, 0, time.UTC), 100, 50, 2),   // net +50
		makeWeek("2026-W02", time.Date(2026, 1, 12, 0, 0, 0, 0, time.UTC), 100, 50, 2),  // net +50
		makeWeek("2026-W03", time.Date(2026, 1, 19, 0, 0, 0, 0, time.UTC), 50, 200, 15), // net -150, 15>6 ✓
		makeWeek("2026-W04", time.Date(2026, 1, 26, 0, 0, 0, 0, time.UTC), 30, 100, 14), // net -70, 14>6 ✓
		makeWeek("2026-W05", time.Date(2026, 2, 2, 0, 0, 0, 0, time.UTC), 20, 80, 13),   // net -60, 13>6 ✓
	}
	// sorted commits = [2,2,13,14,15], median = 13. >13: W03(15), W04(14). Only 2.
	// Need commits > median. median=13. 13 is NOT > 13.
	// So W03(15>13)✓, W04(14>13)✓, W05(13>13)✗. Two consecutive refactor weeks.

	periods := DetectRefactorPeriods(weeks)
	require.Len(t, periods, 1)
	assert.Equal(t, weeks[2].WeekStart, periods[0].Start)
	assert.Equal(t, weeks[3].WeekStart.AddDate(0, 0, 6), periods[0].End)
	assert.Equal(t, -220, periods[0].NetLOCChange) // -150 + -70
	assert.Equal(t, 2, periods[0].Weeks)
}

func TestDetectRefactorPeriods_SplitPeriods(t *testing.T) {
	// Two separate refactor weeks separated by a non-refactor week
	// commits = [15,15,15,2,2], sorted = [2,2,15,15,15], median = 15. >15: none.
	// Fix: lower the median by adding more low-commit weeks.
	// commits = [20,5,20,2,2], sorted = [2,2,5,20,20], median = 5. >5: W01(20),W02(5→no),W03(20)
	weeks := []models.WeeklyActivity{
		makeWeek("2026-W01", time.Date(2026, 1, 5, 0, 0, 0, 0, time.UTC), 10, 100, 20), // net -90, 20>5 ✓
		makeWeek("2026-W02", time.Date(2026, 1, 12, 0, 0, 0, 0, time.UTC), 200, 50, 5), // net +150, NOT refactor
		makeWeek("2026-W03", time.Date(2026, 1, 19, 0, 0, 0, 0, time.UTC), 10, 80, 20), // net -70, 20>5 ✓
		makeWeek("2026-W04", time.Date(2026, 1, 26, 0, 0, 0, 0, time.UTC), 5, 5, 2),    // net 0, low commits
		makeWeek("2026-W05", time.Date(2026, 2, 2, 0, 0, 0, 0, time.UTC), 5, 5, 2),     // net 0, low commits
	}

	periods := DetectRefactorPeriods(weeks)
	require.Len(t, periods, 2)
	assert.Equal(t, 1, periods[0].Weeks)
	assert.Equal(t, 1, periods[1].Weeks)
}

func TestDetectRefactorPeriods_NoRefactors(t *testing.T) {
	// All weeks have positive net LOC
	weeks := []models.WeeklyActivity{
		makeWeek("2026-W01", time.Date(2026, 1, 5, 0, 0, 0, 0, time.UTC), 100, 50, 10),
		makeWeek("2026-W02", time.Date(2026, 1, 12, 0, 0, 0, 0, time.UTC), 200, 100, 10),
		makeWeek("2026-W03", time.Date(2026, 1, 19, 0, 0, 0, 0, time.UTC), 150, 50, 10),
	}

	periods := DetectRefactorPeriods(weeks)
	assert.Empty(t, periods)
}

func TestDetectRefactorPeriods_NegativeLOCBelowMedian(t *testing.T) {
	// Negative net LOC but below-median commit count → NOT a refactor week
	// Median commits = median of [20, 20, 2] = 20
	weeks := []models.WeeklyActivity{
		makeWeek("2026-W01", time.Date(2026, 1, 5, 0, 0, 0, 0, time.UTC), 100, 50, 20),
		makeWeek("2026-W02", time.Date(2026, 1, 12, 0, 0, 0, 0, time.UTC), 100, 50, 20),
		makeWeek("2026-W03", time.Date(2026, 1, 19, 0, 0, 0, 0, time.UTC), 10, 50, 2), // net -40, but commits(2) < median(20)
	}

	periods := DetectRefactorPeriods(weeks)
	assert.Empty(t, periods)
}

func TestDetectRefactorPeriods_EmptyInput(t *testing.T) {
	periods := DetectRefactorPeriods(nil)
	assert.Empty(t, periods)
}
