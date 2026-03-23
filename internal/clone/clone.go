package clone

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"time"

	"github.com/go-git/go-billy/v5/memfs"
	"github.com/go-git/go-billy/v5/osfs"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/go-git/go-git/v5/storage/filesystem"
	"github.com/go-git/go-git/v5/storage/memory"

	gritmetrics "github.com/grit-app/grit/internal/metrics"
)

type CloneResult struct {
	Repo    *git.Repository
	DiskDir string
}

func Clone(ctx context.Context, owner, repo, token string, sizeKB int64, cloneDir string, thresholdKB int64) (*CloneResult, error) {
	start := time.Now()
	defer func() {
		gritmetrics.CloneDurationSeconds.Observe(time.Since(start).Seconds())
	}()

	url := fmt.Sprintf("https://github.com/%s/%s.git", owner, repo)

	opts := &git.CloneOptions{
		URL:          url,
		Depth:        1,
		SingleBranch: true,
	}

	if token != "" {
		opts.Auth = &http.BasicAuth{
			Username: "x-access-token",
			Password: token,
		}
	}

	if sizeKB > 0 && sizeKB > thresholdKB {
		return cloneToDisk(ctx, owner, repo, opts, cloneDir)
	}
	return cloneToMemory(ctx, opts)
}

func cloneToMemory(ctx context.Context, opts *git.CloneOptions) (*CloneResult, error) {
	storer := memory.NewStorage()
	fs := memfs.New()

	r, err := git.CloneContext(ctx, storer, fs, opts)
	if err != nil {
		return nil, fmt.Errorf("clone: memory clone failed: %w", err)
	}

	return &CloneResult{Repo: r}, nil
}

func cloneToDisk(ctx context.Context, owner, repo string, opts *git.CloneOptions, cloneDir string) (*CloneResult, error) {
	dir := filepath.Join(cloneDir, owner, repo, fmt.Sprintf("%d", time.Now().UnixNano()))

	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, fmt.Errorf("clone: create dir: %w", err)
	}

	dotGit := filepath.Join(dir, ".git")
	if err := os.MkdirAll(dotGit, 0o755); err != nil {
		return nil, fmt.Errorf("clone: create .git dir: %w", err)
	}

	storer := filesystem.NewStorage(osfs.New(dotGit), nil)
	wt := osfs.New(dir)

	r, err := git.CloneContext(ctx, storer, wt, opts)
	if err != nil {
		os.RemoveAll(dir)
		return nil, fmt.Errorf("clone: disk clone failed: %w", err)
	}

	return &CloneResult{Repo: r, DiskDir: dir}, nil
}

func StartCleanup(ctx context.Context, cloneDir string, maxAge time.Duration, interval time.Duration) {
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				cleanOldClones(cloneDir, maxAge)
			}
		}
	}()
}

func cleanOldClones(cloneDir string, maxAge time.Duration) {
	entries, err := os.ReadDir(cloneDir)
	if err != nil {
		return
	}

	cutoff := time.Now().Add(-maxAge)
	var removed int

	for _, ownerEntry := range entries {
		if !ownerEntry.IsDir() {
			continue
		}
		ownerPath := filepath.Join(cloneDir, ownerEntry.Name())
		repoEntries, err := os.ReadDir(ownerPath)
		if err != nil {
			continue
		}
		for _, repoEntry := range repoEntries {
			if !repoEntry.IsDir() {
				continue
			}
			repoPath := filepath.Join(ownerPath, repoEntry.Name())
			cloneEntries, err := os.ReadDir(repoPath)
			if err != nil {
				continue
			}
			for _, cloneEntry := range cloneEntries {
				if !cloneEntry.IsDir() {
					continue
				}
				info, err := cloneEntry.Info()
				if err != nil {
					continue
				}
				if info.ModTime().Before(cutoff) {
					clonePath := filepath.Join(repoPath, cloneEntry.Name())
					if err := os.RemoveAll(clonePath); err == nil {
						removed++
					}
				}
			}
		}
	}

	if removed > 0 {
		slog.Info("clone cleanup completed", "removed", removed, "dir", cloneDir)
	}
}
