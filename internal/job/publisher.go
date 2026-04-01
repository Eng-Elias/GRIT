package job

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/nats-io/nats.go"

	"github.com/grit-app/grit/internal/cache"
	"github.com/grit-app/grit/internal/models"
)

const (
	StreamName        = "GRIT"
	Subject           = "grit.jobs.analysis"
	ComplexitySubject = "grit.jobs.complexity"
	ChurnSubject      = "grit.jobs.churn"
)

type JobPayload struct {
	JobID string `json:"job_id"`
	Owner string `json:"owner"`
	Repo  string `json:"repo"`
	SHA   string `json:"sha"`
	Token string `json:"token,omitempty"`
}

type Publisher struct {
	js    nats.JetStreamContext
	cache *cache.Cache
}

func NewPublisher(js nats.JetStreamContext, c *cache.Cache) *Publisher {
	return &Publisher{js: js, cache: c}
}

func EnsureStream(js nats.JetStreamContext) error {
	_, err := js.StreamInfo(StreamName)
	if err != nil {
		_, err = js.AddStream(&nats.StreamConfig{
			Name:       StreamName,
			Subjects:   []string{"grit.jobs.>"},
			Retention:  nats.WorkQueuePolicy,
			MaxAge:     1 * time.Hour,
			Discard:    nats.DiscardOld,
			Duplicates: 5 * time.Minute,
		})
		if err != nil {
			return fmt.Errorf("job: create stream: %w", err)
		}
	}
	return nil
}

func (p *Publisher) PublishComplexity(ctx context.Context, owner, repo, sha, token string) (string, error) {
	existing, err := p.cache.GetActiveComplexityJob(ctx, owner, repo, sha)
	if err == nil && existing != "" {
		return existing, nil
	}

	jobID := uuid.New().String()

	payload := JobPayload{
		JobID: jobID,
		Owner: owner,
		Repo:  repo,
		SHA:   sha,
		Token: token,
	}

	data, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("job: marshal complexity payload: %w", err)
	}

	msg := &nats.Msg{
		Subject: ComplexitySubject,
		Data:    data,
		Header:  nats.Header{},
	}
	msg.Header.Set("Nats-Msg-Id", fmt.Sprintf("%s/%s:%s:complexity", owner, repo, sha))

	_, err = p.js.PublishMsg(msg, nats.Context(ctx))
	if err != nil {
		return "", fmt.Errorf("job: publish complexity: %w", err)
	}

	job := models.AnalysisJob{
		JobID:     jobID,
		Owner:     owner,
		Repo:      repo,
		SHA:       sha,
		Status:    models.JobStatusQueued,
		Progress:  models.NewJobProgress(),
		CreatedAt: time.Now().UTC(),
	}

	if err := p.cache.SetJob(ctx, jobID, &job); err != nil {
		return "", fmt.Errorf("job: store complexity job state: %w", err)
	}

	if err := p.cache.SetActiveComplexityJob(ctx, owner, repo, sha, jobID); err != nil {
		return "", fmt.Errorf("job: store active complexity job: %w", err)
	}

	return jobID, nil
}

func (p *Publisher) PublishChurn(ctx context.Context, owner, repo, sha, token string) (string, error) {
	existing, err := p.cache.GetActiveChurnJob(ctx, owner, repo, sha)
	if err == nil && existing != "" {
		return existing, nil
	}

	jobID := uuid.New().String()

	payload := JobPayload{
		JobID: jobID,
		Owner: owner,
		Repo:  repo,
		SHA:   sha,
		Token: token,
	}

	data, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("job: marshal churn payload: %w", err)
	}

	msg := &nats.Msg{
		Subject: ChurnSubject,
		Data:    data,
		Header:  nats.Header{},
	}
	msg.Header.Set("Nats-Msg-Id", fmt.Sprintf("%s/%s:%s:churn", owner, repo, sha))

	_, err = p.js.PublishMsg(msg, nats.Context(ctx))
	if err != nil {
		return "", fmt.Errorf("job: publish churn: %w", err)
	}

	job := models.AnalysisJob{
		JobID:     jobID,
		Owner:     owner,
		Repo:      repo,
		SHA:       sha,
		Status:    models.JobStatusQueued,
		Progress:  models.NewJobProgress(),
		CreatedAt: time.Now().UTC(),
	}

	if err := p.cache.SetJob(ctx, jobID, &job); err != nil {
		return "", fmt.Errorf("job: store churn job state: %w", err)
	}

	if err := p.cache.SetActiveChurnJob(ctx, owner, repo, sha, jobID); err != nil {
		return "", fmt.Errorf("job: store active churn job: %w", err)
	}

	return jobID, nil
}

func (p *Publisher) Publish(ctx context.Context, owner, repo, sha, token string) (string, error) {
	existing, err := p.cache.GetActiveJob(ctx, owner, repo, sha)
	if err == nil && existing != "" {
		return existing, nil
	}

	jobID := uuid.New().String()

	payload := JobPayload{
		JobID: jobID,
		Owner: owner,
		Repo:  repo,
		SHA:   sha,
		Token: token,
	}

	data, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("job: marshal payload: %w", err)
	}

	msg := &nats.Msg{
		Subject: Subject,
		Data:    data,
		Header:  nats.Header{},
	}
	msg.Header.Set("Nats-Msg-Id", fmt.Sprintf("%s/%s:%s", owner, repo, sha))

	_, err = p.js.PublishMsg(msg, nats.Context(ctx))
	if err != nil {
		return "", fmt.Errorf("job: publish: %w", err)
	}

	job := models.AnalysisJob{
		JobID:     jobID,
		Owner:     owner,
		Repo:      repo,
		SHA:       sha,
		Status:    models.JobStatusQueued,
		Progress:  models.NewJobProgress(),
		CreatedAt: time.Now().UTC(),
	}

	if err := p.cache.SetJob(ctx, jobID, &job); err != nil {
		return "", fmt.Errorf("job: store job state: %w", err)
	}

	if err := p.cache.SetActiveJob(ctx, owner, repo, sha, jobID); err != nil {
		return "", fmt.Errorf("job: store active job: %w", err)
	}

	return jobID, nil
}
