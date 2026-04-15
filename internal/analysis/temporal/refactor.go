package temporal

import (
	"sort"

	"github.com/grit-app/grit/internal/models"
)

// DetectRefactorPeriods identifies consecutive weeks where net LOC change is negative
// AND commit count exceeds the median weekly commit count. Groups them into periods.
func DetectRefactorPeriods(weeks []models.WeeklyActivity) []models.RefactorPeriod {
	if len(weeks) == 0 {
		return nil
	}

	medianCommits := computeMedianCommits(weeks)

	// Identify refactor weeks
	type refactorWeek struct {
		index     int
		netChange int
	}

	var refactorWeeks []refactorWeek
	for i, w := range weeks {
		netLOC := w.Additions - w.Deletions
		if netLOC < 0 && w.Commits > medianCommits {
			refactorWeeks = append(refactorWeeks, refactorWeek{index: i, netChange: netLOC})
		}
	}

	if len(refactorWeeks) == 0 {
		return nil
	}

	// Group consecutive refactor weeks into periods
	var periods []models.RefactorPeriod
	start := refactorWeeks[0]
	prev := refactorWeeks[0]
	totalNet := start.netChange
	weekCount := 1

	for j := 1; j < len(refactorWeeks); j++ {
		curr := refactorWeeks[j]
		if curr.index == prev.index+1 {
			// Consecutive
			totalNet += curr.netChange
			weekCount++
			prev = curr
		} else {
			// Gap — close current period
			periods = append(periods, models.RefactorPeriod{
				Start:        weeks[start.index].WeekStart,
				End:          weeks[prev.index].WeekStart.AddDate(0, 0, 6),
				NetLOCChange: totalNet,
				Weeks:        weekCount,
			})
			start = curr
			prev = curr
			totalNet = curr.netChange
			weekCount = 1
		}
	}

	// Close last period
	periods = append(periods, models.RefactorPeriod{
		Start:        weeks[start.index].WeekStart,
		End:          weeks[prev.index].WeekStart.AddDate(0, 0, 6),
		NetLOCChange: totalNet,
		Weeks:        weekCount,
	})

	return periods
}

// computeMedianCommits returns the median commit count from weekly activity.
func computeMedianCommits(weeks []models.WeeklyActivity) int {
	counts := make([]int, len(weeks))
	for i, w := range weeks {
		counts[i] = w.Commits
	}

	sort.Ints(counts)

	n := len(counts)
	if n%2 == 0 {
		return (counts[n/2-1] + counts[n/2]) / 2
	}
	return counts[n/2]
}
