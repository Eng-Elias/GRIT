package cache

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

var ErrCacheMiss = errors.New("cache: miss")

const (
	CoreAnalysisTTL        = 24 * time.Hour
	ComplexityAnalysisTTL  = 24 * time.Hour
	ChurnAnalysisTTL       = 24 * time.Hour
	ContributorAnalysisTTL = 48 * time.Hour
	JobTTL                 = 1 * time.Hour
	ActiveJobTTL           = 10 * time.Minute
)

type Cache struct {
	client *redis.Client
}

func New(redisURL string) (*Cache, error) {
	opts, err := redis.ParseURL(redisURL)
	if err != nil {
		return nil, fmt.Errorf("cache: invalid redis URL: %w", err)
	}
	client := redis.NewClient(opts)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("cache: redis ping failed: %w", err)
	}

	return &Cache{client: client}, nil
}

func (c *Cache) Ping(ctx context.Context) error {
	return c.client.Ping(ctx).Err()
}

func (c *Cache) Close() error {
	return c.client.Close()
}

func coreKey(owner, repo, sha string) string {
	return fmt.Sprintf("%s/%s:%s:core", owner, repo, sha)
}

func jobKey(jobID string) string {
	return fmt.Sprintf("job:%s", jobID)
}

func activeKey(owner, repo, sha string) string {
	return fmt.Sprintf("active:%s/%s:%s", owner, repo, sha)
}

func complexityKey(owner, repo, sha string) string {
	return fmt.Sprintf("%s/%s:%s:complexity", owner, repo, sha)
}

func activeComplexityKey(owner, repo, sha string) string {
	return fmt.Sprintf("active:%s/%s:%s:complexity", owner, repo, sha)
}

func (c *Cache) GetAnalysis(ctx context.Context, owner, repo, sha string) ([]byte, error) {
	if sha == "" {
		return c.findAnalysis(ctx, owner, repo)
	}
	data, err := c.client.Get(ctx, coreKey(owner, repo, sha)).Bytes()
	if errors.Is(err, redis.Nil) {
		return nil, ErrCacheMiss
	}
	if err != nil {
		return nil, fmt.Errorf("cache: get analysis: %w", err)
	}
	return data, nil
}

func (c *Cache) findAnalysis(ctx context.Context, owner, repo string) ([]byte, error) {
	pattern := fmt.Sprintf("%s/%s:*:core", owner, repo)
	var cursor uint64
	for {
		keys, next, err := c.client.Scan(ctx, cursor, pattern, 100).Result()
		if err != nil {
			return nil, fmt.Errorf("cache: scan analysis: %w", err)
		}
		if len(keys) > 0 {
			data, err := c.client.Get(ctx, keys[0]).Bytes()
			if errors.Is(err, redis.Nil) {
				return nil, ErrCacheMiss
			}
			if err != nil {
				return nil, fmt.Errorf("cache: get analysis: %w", err)
			}
			return data, nil
		}
		cursor = next
		if cursor == 0 {
			break
		}
	}
	return nil, ErrCacheMiss
}

func (c *Cache) SetAnalysis(ctx context.Context, owner, repo, sha string, data []byte) error {
	return c.client.Set(ctx, coreKey(owner, repo, sha), data, CoreAnalysisTTL).Err()
}

func (c *Cache) DeleteAnalysis(ctx context.Context, owner, repo string) error {
	pattern := fmt.Sprintf("%s/%s:*:core", owner, repo)
	iter := c.client.Scan(ctx, 0, pattern, 100).Iterator()
	for iter.Next(ctx) {
		c.client.Del(ctx, iter.Val())
	}
	return iter.Err()
}

func (c *Cache) GetJob(ctx context.Context, jobID string) ([]byte, error) {
	data, err := c.client.Get(ctx, jobKey(jobID)).Bytes()
	if errors.Is(err, redis.Nil) {
		return nil, ErrCacheMiss
	}
	if err != nil {
		return nil, fmt.Errorf("cache: get job: %w", err)
	}
	return data, nil
}

func (c *Cache) SetJob(ctx context.Context, jobID string, job interface{}) error {
	data, err := json.Marshal(job)
	if err != nil {
		return fmt.Errorf("cache: marshal job: %w", err)
	}
	return c.client.Set(ctx, jobKey(jobID), data, JobTTL).Err()
}

func (c *Cache) GetActiveJob(ctx context.Context, owner, repo, sha string) (string, error) {
	jobID, err := c.client.Get(ctx, activeKey(owner, repo, sha)).Result()
	if errors.Is(err, redis.Nil) {
		return "", ErrCacheMiss
	}
	if err != nil {
		return "", fmt.Errorf("cache: get active job: %w", err)
	}
	return jobID, nil
}

func (c *Cache) SetActiveJob(ctx context.Context, owner, repo, sha, jobID string) error {
	return c.client.Set(ctx, activeKey(owner, repo, sha), jobID, ActiveJobTTL).Err()
}

func (c *Cache) DeleteActiveJob(ctx context.Context, owner, repo, sha string) error {
	return c.client.Del(ctx, activeKey(owner, repo, sha)).Err()
}

func (c *Cache) GetComplexity(ctx context.Context, owner, repo, sha string) ([]byte, error) {
	if sha == "" {
		return c.findComplexity(ctx, owner, repo)
	}
	data, err := c.client.Get(ctx, complexityKey(owner, repo, sha)).Bytes()
	if errors.Is(err, redis.Nil) {
		return nil, ErrCacheMiss
	}
	if err != nil {
		return nil, fmt.Errorf("cache: get complexity: %w", err)
	}
	return data, nil
}

func (c *Cache) findComplexity(ctx context.Context, owner, repo string) ([]byte, error) {
	pattern := fmt.Sprintf("%s/%s:*:complexity", owner, repo)
	var cursor uint64
	for {
		keys, next, err := c.client.Scan(ctx, cursor, pattern, 100).Result()
		if err != nil {
			return nil, fmt.Errorf("cache: scan complexity: %w", err)
		}
		if len(keys) > 0 {
			data, err := c.client.Get(ctx, keys[0]).Bytes()
			if errors.Is(err, redis.Nil) {
				return nil, ErrCacheMiss
			}
			if err != nil {
				return nil, fmt.Errorf("cache: get complexity: %w", err)
			}
			return data, nil
		}
		cursor = next
		if cursor == 0 {
			break
		}
	}
	return nil, ErrCacheMiss
}

func (c *Cache) SetComplexity(ctx context.Context, owner, repo, sha string, data []byte) error {
	return c.client.Set(ctx, complexityKey(owner, repo, sha), data, ComplexityAnalysisTTL).Err()
}

func (c *Cache) DeleteComplexity(ctx context.Context, owner, repo string) error {
	pattern := fmt.Sprintf("%s/%s:*:complexity", owner, repo)
	iter := c.client.Scan(ctx, 0, pattern, 100).Iterator()
	for iter.Next(ctx) {
		c.client.Del(ctx, iter.Val())
	}
	return iter.Err()
}

func (c *Cache) GetActiveComplexityJob(ctx context.Context, owner, repo, sha string) (string, error) {
	jobID, err := c.client.Get(ctx, activeComplexityKey(owner, repo, sha)).Result()
	if errors.Is(err, redis.Nil) {
		return "", ErrCacheMiss
	}
	if err != nil {
		return "", fmt.Errorf("cache: get active complexity job: %w", err)
	}
	return jobID, nil
}

func (c *Cache) SetActiveComplexityJob(ctx context.Context, owner, repo, sha, jobID string) error {
	return c.client.Set(ctx, activeComplexityKey(owner, repo, sha), jobID, ActiveJobTTL).Err()
}

func (c *Cache) DeleteActiveComplexityJob(ctx context.Context, owner, repo, sha string) error {
	return c.client.Del(ctx, activeComplexityKey(owner, repo, sha)).Err()
}

func contributorKey(owner, repo, sha string) string {
	return fmt.Sprintf("%s/%s:%s:contributors", owner, repo, sha)
}

func activeBlameKey(owner, repo, sha string) string {
	return fmt.Sprintf("active:%s/%s:%s:blame", owner, repo, sha)
}

func (c *Cache) GetContributors(ctx context.Context, owner, repo, sha string) ([]byte, error) {
	if sha == "" {
		return c.findContributors(ctx, owner, repo)
	}
	data, err := c.client.Get(ctx, contributorKey(owner, repo, sha)).Bytes()
	if errors.Is(err, redis.Nil) {
		return nil, ErrCacheMiss
	}
	if err != nil {
		return nil, fmt.Errorf("cache: get contributors: %w", err)
	}
	return data, nil
}

func (c *Cache) findContributors(ctx context.Context, owner, repo string) ([]byte, error) {
	pattern := fmt.Sprintf("%s/%s:*:contributors", owner, repo)
	var cursor uint64
	for {
		keys, next, err := c.client.Scan(ctx, cursor, pattern, 100).Result()
		if err != nil {
			return nil, fmt.Errorf("cache: scan contributors: %w", err)
		}
		if len(keys) > 0 {
			data, err := c.client.Get(ctx, keys[0]).Bytes()
			if errors.Is(err, redis.Nil) {
				return nil, ErrCacheMiss
			}
			if err != nil {
				return nil, fmt.Errorf("cache: get contributors: %w", err)
			}
			return data, nil
		}
		cursor = next
		if cursor == 0 {
			break
		}
	}
	return nil, ErrCacheMiss
}

func (c *Cache) SetContributors(ctx context.Context, owner, repo, sha string, data []byte) error {
	return c.client.Set(ctx, contributorKey(owner, repo, sha), data, ContributorAnalysisTTL).Err()
}

func (c *Cache) DeleteContributors(ctx context.Context, owner, repo string) error {
	pattern := fmt.Sprintf("%s/%s:*:contributors", owner, repo)
	iter := c.client.Scan(ctx, 0, pattern, 100).Iterator()
	for iter.Next(ctx) {
		c.client.Del(ctx, iter.Val())
	}
	return iter.Err()
}

func (c *Cache) GetActiveBlameJob(ctx context.Context, owner, repo, sha string) (string, error) {
	jobID, err := c.client.Get(ctx, activeBlameKey(owner, repo, sha)).Result()
	if errors.Is(err, redis.Nil) {
		return "", ErrCacheMiss
	}
	if err != nil {
		return "", fmt.Errorf("cache: get active blame job: %w", err)
	}
	return jobID, nil
}

func (c *Cache) SetActiveBlameJob(ctx context.Context, owner, repo, sha, jobID string) error {
	return c.client.Set(ctx, activeBlameKey(owner, repo, sha), jobID, ActiveJobTTL).Err()
}

func (c *Cache) DeleteActiveBlameJob(ctx context.Context, owner, repo, sha string) error {
	return c.client.Del(ctx, activeBlameKey(owner, repo, sha)).Err()
}

func churnKey(owner, repo, sha string) string {
	return fmt.Sprintf("%s/%s:%s:churn", owner, repo, sha)
}

func activeChurnKey(owner, repo, sha string) string {
	return fmt.Sprintf("active:%s/%s:%s:churn", owner, repo, sha)
}

func (c *Cache) GetChurn(ctx context.Context, owner, repo, sha string) ([]byte, error) {
	if sha == "" {
		return c.findChurn(ctx, owner, repo)
	}
	data, err := c.client.Get(ctx, churnKey(owner, repo, sha)).Bytes()
	if errors.Is(err, redis.Nil) {
		return nil, ErrCacheMiss
	}
	if err != nil {
		return nil, fmt.Errorf("cache: get churn: %w", err)
	}
	return data, nil
}

func (c *Cache) findChurn(ctx context.Context, owner, repo string) ([]byte, error) {
	pattern := fmt.Sprintf("%s/%s:*:churn", owner, repo)
	var cursor uint64
	for {
		keys, next, err := c.client.Scan(ctx, cursor, pattern, 100).Result()
		if err != nil {
			return nil, fmt.Errorf("cache: scan churn: %w", err)
		}
		if len(keys) > 0 {
			data, err := c.client.Get(ctx, keys[0]).Bytes()
			if errors.Is(err, redis.Nil) {
				return nil, ErrCacheMiss
			}
			if err != nil {
				return nil, fmt.Errorf("cache: get churn: %w", err)
			}
			return data, nil
		}
		cursor = next
		if cursor == 0 {
			break
		}
	}
	return nil, ErrCacheMiss
}

func (c *Cache) SetChurn(ctx context.Context, owner, repo, sha string, data []byte) error {
	return c.client.Set(ctx, churnKey(owner, repo, sha), data, ChurnAnalysisTTL).Err()
}

func (c *Cache) DeleteChurn(ctx context.Context, owner, repo string) error {
	pattern := fmt.Sprintf("%s/%s:*:churn", owner, repo)
	iter := c.client.Scan(ctx, 0, pattern, 100).Iterator()
	for iter.Next(ctx) {
		c.client.Del(ctx, iter.Val())
	}
	return iter.Err()
}

func (c *Cache) GetActiveChurnJob(ctx context.Context, owner, repo, sha string) (string, error) {
	jobID, err := c.client.Get(ctx, activeChurnKey(owner, repo, sha)).Result()
	if errors.Is(err, redis.Nil) {
		return "", ErrCacheMiss
	}
	if err != nil {
		return "", fmt.Errorf("cache: get active churn job: %w", err)
	}
	return jobID, nil
}

func (c *Cache) SetActiveChurnJob(ctx context.Context, owner, repo, sha, jobID string) error {
	return c.client.Set(ctx, activeChurnKey(owner, repo, sha), jobID, ActiveJobTTL).Err()
}

func (c *Cache) DeleteActiveChurnJob(ctx context.Context, owner, repo, sha string) error {
	return c.client.Del(ctx, activeChurnKey(owner, repo, sha)).Err()
}
