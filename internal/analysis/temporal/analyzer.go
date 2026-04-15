package temporal

import (
	"context"
	"time"

	"github.com/go-git/go-git/v5"

	gh "github.com/grit-app/grit/internal/github"
	"github.com/grit-app/grit/internal/models"
)

// AnalyzeOptions configures the temporal analysis.
type AnalyzeOptions struct {
	Period string // "1y", "2y", "3y"
	Owner  string
	Repo   string
	Token  string
}

// Analyze runs the full temporal analysis: LOC over time, velocity, refactor detection,
// and PR merge time.
func Analyze(ctx context.Context, repo *git.Repository, opts AnalyzeOptions) (*models.TemporalResult, error) {
	period := opts.Period
	if period == "" {
		period = "3y"
	}

	// 1. LOC over time
	snapshots, err := ComputeLOCOverTime(repo, period)
	if err != nil {
		return nil, err
	}

	// 2. Velocity metrics
	weeklyActivity, authorActivity, cadence, err := ComputeVelocity(repo)
	if err != nil {
		return nil, err
	}

	// 3. Refactor detection (derived from weekly velocity)
	refactorPeriods := DetectRefactorPeriods(weeklyActivity)
	if refactorPeriods == nil {
		refactorPeriods = []models.RefactorPeriod{}
	}

	// 4. PR merge time (optional, gracefully degrades)
	var prMergeTime *models.PRMergeTime
	if opts.Token != "" {
		ghClient := gh.NewClient(opts.Token)
		prMergeTime, _ = ghClient.FetchPRMergeTimes(ctx, opts.Owner, opts.Repo)
	}

	// Ensure nil slices become empty arrays for JSON
	if snapshots == nil {
		snapshots = []models.MonthlySnapshot{}
	}
	if weeklyActivity == nil {
		weeklyActivity = []models.WeeklyActivity{}
	}
	if authorActivity == nil {
		authorActivity = []models.AuthorWeeklyActivity{}
	}
	if cadence == nil {
		cadence = []models.CommitCadence{}
	}

	now := time.Now().UTC()
	result := &models.TemporalResult{
		Repository: models.Repository{
			Owner:    opts.Owner,
			Name:     opts.Repo,
			FullName: opts.Owner + "/" + opts.Repo,
		},
		Period:      period,
		LOCOverTime: snapshots,
		Velocity: models.VelocityMetrics{
			WeeklyActivity: weeklyActivity,
			AuthorActivity: authorActivity,
			CommitCadence:  cadence,
			PRMergeTime:    prMergeTime,
		},
		RefactorPeriods: refactorPeriods,
		TotalMonths:     len(snapshots),
		TotalWeeks:      len(weeklyActivity),
		AnalyzedAt:      now,
	}

	return result, nil
}
