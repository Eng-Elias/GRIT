package models

import "time"

// MonthlySnapshot is a point-in-time LOC measurement at a monthly boundary.
type MonthlySnapshot struct {
	Date       time.Time      `json:"date"`
	TotalLOC   int            `json:"total_loc"`
	ByLanguage map[string]int `json:"by_language"`
}

// WeeklyActivity is aggregated additions and deletions for a single ISO week.
type WeeklyActivity struct {
	Week      string    `json:"week"`
	WeekStart time.Time `json:"week_start"`
	Additions int       `json:"additions"`
	Deletions int       `json:"deletions"`
	Commits   int       `json:"commits"`
}

// AuthorWeeklyActivity is per-author weekly activity for a top contributor.
type AuthorWeeklyActivity struct {
	Email          string           `json:"email"`
	Name           string           `json:"name"`
	TotalAdditions int              `json:"total_additions"`
	TotalDeletions int              `json:"total_deletions"`
	Weeks          []WeeklyActivity `json:"weeks"`
}

// CommitCadence is a rolling 4-week window of commit frequency.
type CommitCadence struct {
	WindowStart   time.Time `json:"window_start"`
	WindowEnd     time.Time `json:"window_end"`
	CommitsPerDay float64   `json:"commits_per_day"`
}

// PRMergeTime holds pull request merge time percentile statistics.
type PRMergeTime struct {
	MedianHours float64 `json:"median_hours"`
	P75Hours    float64 `json:"p75_hours"`
	P95Hours    float64 `json:"p95_hours"`
	SampleSize  int     `json:"sample_size"`
}

// VelocityMetrics is a container for all velocity-related data.
type VelocityMetrics struct {
	WeeklyActivity []WeeklyActivity       `json:"weekly_activity"`
	AuthorActivity []AuthorWeeklyActivity `json:"author_activity"`
	CommitCadence  []CommitCadence        `json:"commit_cadence"`
	PRMergeTime    *PRMergeTime           `json:"pr_merge_time"`
}

// RefactorPeriod is a contiguous span of refactor weeks.
type RefactorPeriod struct {
	Start        time.Time `json:"start"`
	End          time.Time `json:"end"`
	NetLOCChange int       `json:"net_loc_change"`
	Weeks        int       `json:"weeks"`
}

// TemporalResult is the complete temporal analysis output cached in Redis.
type TemporalResult struct {
	Repository      Repository        `json:"repository"`
	Period          string            `json:"period"`
	LOCOverTime     []MonthlySnapshot `json:"loc_over_time"`
	Velocity        VelocityMetrics   `json:"velocity"`
	RefactorPeriods []RefactorPeriod  `json:"refactor_periods"`
	TotalMonths     int               `json:"total_months"`
	TotalWeeks      int               `json:"total_weeks"`
	AnalyzedAt      time.Time         `json:"analyzed_at"`
	Cached          bool              `json:"cached"`
	CachedAt        *time.Time        `json:"cached_at"`
}

// TemporalSummary is abbreviated temporal data embedded in the main analysis response.
type TemporalSummary struct {
	Status             string  `json:"status"`
	CurrentLOC         int     `json:"current_loc"`
	LOCTrend6mPercent  float64 `json:"loc_trend_6m_percent"`
	AvgWeeklyCommits   float64 `json:"avg_weekly_commits"`
	TemporalURL        string  `json:"temporal_url"`
}
