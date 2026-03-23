package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoad_Defaults(t *testing.T) {
	os.Clearenv()
	cfg, err := Load()
	require.NoError(t, err)

	assert.Equal(t, 8080, cfg.Port)
	assert.Equal(t, "redis://localhost:6379", cfg.RedisURL)
	assert.Equal(t, "nats://localhost:4222", cfg.NATSURL)
	assert.Equal(t, "", cfg.GitHubToken)
	assert.Equal(t, "/tmp/grit-clones", cfg.CloneDir)
	assert.Equal(t, int64(51200), cfg.CloneSizeThresholdKB)
}

func TestLoad_CustomValues(t *testing.T) {
	tests := []struct {
		name     string
		envVars  map[string]string
		validate func(t *testing.T, cfg *Config)
	}{
		{
			name:    "custom port",
			envVars: map[string]string{"PORT": "9090"},
			validate: func(t *testing.T, cfg *Config) {
				assert.Equal(t, 9090, cfg.Port)
			},
		},
		{
			name:    "custom redis URL",
			envVars: map[string]string{"REDIS_URL": "redis://myhost:6380"},
			validate: func(t *testing.T, cfg *Config) {
				assert.Equal(t, "redis://myhost:6380", cfg.RedisURL)
			},
		},
		{
			name:    "custom NATS URL",
			envVars: map[string]string{"NATS_URL": "nats://myhost:4223"},
			validate: func(t *testing.T, cfg *Config) {
				assert.Equal(t, "nats://myhost:4223", cfg.NATSURL)
			},
		},
		{
			name:    "GitHub token",
			envVars: map[string]string{"GITHUB_TOKEN": "ghp_test123"},
			validate: func(t *testing.T, cfg *Config) {
				assert.Equal(t, "ghp_test123", cfg.GitHubToken)
			},
		},
		{
			name:    "custom clone dir",
			envVars: map[string]string{"CLONE_DIR": "/var/grit"},
			validate: func(t *testing.T, cfg *Config) {
				assert.Equal(t, "/var/grit", cfg.CloneDir)
			},
		},
		{
			name:    "custom clone threshold",
			envVars: map[string]string{"CLONE_SIZE_THRESHOLD_KB": "102400"},
			validate: func(t *testing.T, cfg *Config) {
				assert.Equal(t, int64(102400), cfg.CloneSizeThresholdKB)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Clearenv()
			for k, v := range tt.envVars {
				os.Setenv(k, v)
			}
			cfg, err := Load()
			require.NoError(t, err)
			tt.validate(t, cfg)
		})
	}
}

func TestLoad_InvalidPort(t *testing.T) {
	os.Clearenv()
	os.Setenv("PORT", "not-a-number")
	_, err := Load()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid PORT")
}

func TestLoad_InvalidThreshold(t *testing.T) {
	os.Clearenv()
	os.Setenv("CLONE_SIZE_THRESHOLD_KB", "abc")
	_, err := Load()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid CLONE_SIZE_THRESHOLD_KB")
}
