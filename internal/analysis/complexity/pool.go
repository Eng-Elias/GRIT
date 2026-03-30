package complexity

import (
	"context"
	"log/slog"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"

	"github.com/grit-app/grit/internal/models"
)

// FileInput represents a file to be analyzed by the worker pool.
type FileInput struct {
	Path string // relative path within clone dir
	LOC  int    // lines of code from core analysis
}

// Pool manages concurrent file parsing using a bounded number of goroutines.
type Pool struct {
	cloneDir string
	workers  int
}

// NewPool creates a new worker pool bounded by runtime.NumCPU().
func NewPool(cloneDir string) *Pool {
	return &Pool{
		cloneDir: cloneDir,
		workers:  runtime.NumCPU(),
	}
}

// Run processes all files concurrently and returns collected results.
// Per-file errors are logged and skipped; the pool never aborts on individual failures.
func (p *Pool) Run(ctx context.Context, files []FileInput) []models.FileComplexity {
	if len(files) == 0 {
		return nil
	}

	filesCh := make(chan FileInput, len(files))
	resultsCh := make(chan *models.FileComplexity, len(files))

	var wg sync.WaitGroup
	for i := 0; i < p.workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for fi := range filesCh {
				select {
				case <-ctx.Done():
					return
				default:
				}

				result := p.processFile(ctx, fi)
				if result != nil {
					resultsCh <- result
				}
			}
		}()
	}

	// Feed files into the channel.
	for _, f := range files {
		filesCh <- f
	}
	close(filesCh)

	// Wait for all workers then close results.
	go func() {
		wg.Wait()
		close(resultsCh)
	}()

	// Collect results.
	var results []models.FileComplexity
	for r := range resultsCh {
		results = append(results, *r)
	}
	return results
}

func (p *Pool) processFile(ctx context.Context, fi FileInput) *models.FileComplexity {
	absPath := filepath.Join(p.cloneDir, fi.Path)
	source, err := os.ReadFile(absPath)
	if err != nil {
		slog.Warn("complexity pool: failed to read file", "path", fi.Path, "error", err)
		return nil
	}

	ext := strings.ToLower(filepath.Ext(fi.Path))
	cfg := GetLanguageConfig(ext)
	if cfg == nil {
		return nil
	}

	return ParseFile(ctx, fi.Path, source, fi.LOC)
}
