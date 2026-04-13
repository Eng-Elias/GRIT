package blame

import (
	"context"
	"runtime"
	"sync"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"

	"github.com/grit-app/grit/internal/models"
)

const blameTimeout = 10 * time.Minute

// Analyzer orchestrates concurrent git blame across all source files.
type Analyzer struct{}

// NewAnalyzer creates a new blame analyzer.
func NewAnalyzer() *Analyzer {
	return &Analyzer{}
}

// Analyze runs git blame on the provided source files using a goroutine pool,
// aggregates results, computes bus factor and per-file contributors.
// A 10-minute hard timeout is applied; partial results are returned on timeout.
func (a *Analyzer) Analyze(ctx context.Context, repo *git.Repository, commitHash plumbing.Hash, files []string) (*models.ContributorResult, error) {
	ctx, cancel := context.WithTimeout(ctx, blameTimeout)
	defer cancel()

	numWorkers := runtime.NumCPU()
	if numWorkers > 4 {
		numWorkers = 4
	}

	var mu sync.Mutex
	fileResults := make(map[string]*FileBlameResult)

	fileCh := make(chan string, len(files))
	for _, f := range files {
		fileCh <- f
	}
	close(fileCh)

	var wg sync.WaitGroup
	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for path := range fileCh {
				if ctx.Err() != nil {
					return
				}
				result, err := BlameFile(ctx, repo, commitHash, path)
				if err != nil {
					continue
				}
				mu.Lock()
				fileResults[path] = result
				mu.Unlock()
			}
		}()
	}
	wg.Wait()

	// Aggregate results.
	aggResult := Aggregate(fileResults)
	busFactor, keyPeople := ComputeBusFactor(aggResult.Authors)
	fileContribs := ComputeFileContributors(fileResults)

	return &models.ContributorResult{
		Authors:            aggResult.Authors,
		BusFactor:          busFactor,
		KeyPeople:          keyPeople,
		FileContributors:   fileContribs,
		TotalLinesAnalyzed: aggResult.TotalLines,
		TotalFilesAnalyzed: len(fileResults),
		Partial:            ctx.Err() != nil,
		AnalyzedAt:         time.Now().UTC(),
	}, nil
}
