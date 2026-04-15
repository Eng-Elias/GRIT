package config

import (
	"fmt"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	Port                 int
	RedisURL             string
	NATSURL              string
	GitHubToken          string
	GeminiAPIKey         string
	CloneDir             string
	CloneSizeThresholdKB int64
}

func Load() (*Config, error) {
	_ = godotenv.Load()

	cfg := &Config{
		Port:                 8080,
		RedisURL:             "redis://localhost:6379",
		NATSURL:              "nats://localhost:4222",
		CloneDir:             "/tmp/grit-clones",
		CloneSizeThresholdKB: 51200,
	}

	if v := os.Getenv("PORT"); v != "" {
		p, err := strconv.Atoi(v)
		if err != nil {
			return nil, fmt.Errorf("invalid PORT: %w", err)
		}
		cfg.Port = p
	}

	if v := os.Getenv("REDIS_URL"); v != "" {
		cfg.RedisURL = v
	}

	if v := os.Getenv("NATS_URL"); v != "" {
		cfg.NATSURL = v
	}

	if v := os.Getenv("GITHUB_TOKEN"); v != "" {
		cfg.GitHubToken = v
	}

	if v := os.Getenv("GEMINI_API_KEY"); v != "" {
		cfg.GeminiAPIKey = v
	}

	if v := os.Getenv("CLONE_DIR"); v != "" {
		cfg.CloneDir = v
	}

	if v := os.Getenv("CLONE_SIZE_THRESHOLD_KB"); v != "" {
		t, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid CLONE_SIZE_THRESHOLD_KB: %w", err)
		}
		cfg.CloneSizeThresholdKB = t
	}

	return cfg, nil
}
