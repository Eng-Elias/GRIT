package models

import "time"

type JobStatus string

const (
	JobStatusQueued    JobStatus = "queued"
	JobStatusRunning   JobStatus = "running"
	JobStatusCompleted JobStatus = "completed"
	JobStatusFailed    JobStatus = "failed"
)

type SubJobStatus string

const (
	SubJobPending   SubJobStatus = "pending"
	SubJobRunning   SubJobStatus = "running"
	SubJobCompleted SubJobStatus = "completed"
	SubJobFailed    SubJobStatus = "failed"
)

type JobProgress struct {
	Clone              SubJobStatus `json:"clone"`
	FileWalk           SubJobStatus `json:"file_walk"`
	MetadataFetch      SubJobStatus `json:"metadata_fetch"`
	CommitActivityFetch SubJobStatus `json:"commit_activity_fetch"`
}

type AnalysisJob struct {
	JobID       string      `json:"job_id"`
	Owner       string      `json:"owner"`
	Repo        string      `json:"repo"`
	SHA         string      `json:"sha"`
	Token       string      `json:"-"`
	Status      JobStatus   `json:"status"`
	Progress    JobProgress `json:"progress"`
	CreatedAt   time.Time   `json:"created_at"`
	CompletedAt *time.Time  `json:"completed_at,omitempty"`
	Error       string      `json:"error,omitempty"`
	ResultURL   string      `json:"result_url,omitempty"`
}

func NewJobProgress() JobProgress {
	return JobProgress{
		Clone:              SubJobPending,
		FileWalk:           SubJobPending,
		MetadataFetch:      SubJobPending,
		CommitActivityFetch: SubJobPending,
	}
}
