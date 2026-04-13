package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	JobsCompletedTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "grit_jobs_completed_total",
		Help: "Total analysis jobs completed",
	})

	JobsFailedTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "grit_jobs_failed_total",
		Help: "Total analysis jobs failed",
	})

	CacheHitTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "grit_cache_hit_total",
		Help: "Total cache hits",
	})

	CacheMissTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "grit_cache_miss_total",
		Help: "Total cache misses",
	})

	GitHubAPIRequestsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "grit_github_api_requests_total",
		Help: "GitHub API requests by endpoint",
	}, []string{"endpoint"})

	CloneDurationSeconds = promauto.NewHistogram(prometheus.HistogramOpts{
		Name:    "grit_clone_duration_seconds",
		Help:    "Time to clone repository",
		Buckets: []float64{0.5, 1, 2, 5, 10, 30, 60, 120, 300},
	})

	AnalysisDurationSeconds = promauto.NewHistogram(prometheus.HistogramOpts{
		Name:    "grit_analysis_duration_seconds",
		Help:    "End-to-end analysis duration",
		Buckets: []float64{1, 5, 10, 30, 60, 120, 300},
	})

	ComplexityAnalysisDuration = promauto.NewHistogram(prometheus.HistogramOpts{
		Name:    "grit_complexity_analysis_duration_seconds",
		Help:    "End-to-end complexity analysis duration",
		Buckets: []float64{1, 5, 10, 30, 60, 120, 300},
	})

	ChurnAnalysisDuration = promauto.NewHistogram(prometheus.HistogramOpts{
		Name:    "grit_churn_analysis_duration_seconds",
		Help:    "End-to-end churn analysis duration",
		Buckets: []float64{1, 5, 10, 30, 60, 120, 300},
	})

	BlameAnalysisDuration = promauto.NewHistogram(prometheus.HistogramOpts{
		Name:    "grit_blame_analysis_duration_seconds",
		Help:    "End-to-end blame/contributor analysis duration",
		Buckets: []float64{1, 5, 10, 30, 60, 120, 300, 600},
	})

	BlameJobsCompletedTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "grit_blame_jobs_completed_total",
		Help: "Total blame/contributor analysis jobs completed",
	})
)
