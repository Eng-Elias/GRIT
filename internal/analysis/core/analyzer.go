package core

import (
	"context"
	"sort"
	"time"

	"github.com/grit-app/grit/internal/clone"
	"github.com/grit-app/grit/internal/github"
	gritmetrics "github.com/grit-app/grit/internal/metrics"
	"github.com/grit-app/grit/internal/models"
)

type Analyzer struct {
	cloneDir    string
	thresholdKB int64
}

func NewAnalyzer(cloneDir string, thresholdKB int64) *Analyzer {
	return &Analyzer{
		cloneDir:    cloneDir,
		thresholdKB: thresholdKB,
	}
}

type ProgressCallback func(step string, status models.SubJobStatus)

func (a *Analyzer) Analyze(ctx context.Context, owner, repo, token string, onProgress ProgressCallback) (*models.AnalysisResult, error) {
	start := time.Now()
	defer func() {
		gritmetrics.AnalysisDurationSeconds.Observe(time.Since(start).Seconds())
	}()

	ghClient := github.NewClient(token)

	if onProgress != nil {
		onProgress("metadata_fetch", models.SubJobRunning)
	}
	meta, sizeKB, defaultBranch, err := ghClient.FetchMetadata(ctx, owner, repo)
	if err != nil {
		if onProgress != nil {
			onProgress("metadata_fetch", models.SubJobFailed)
		}
		return nil, err
	}
	if onProgress != nil {
		onProgress("metadata_fetch", models.SubJobCompleted)
	}

	if onProgress != nil {
		onProgress("clone", models.SubJobRunning)
	}
	cloneResult, err := clone.Clone(ctx, owner, repo, token, sizeKB, a.cloneDir, a.thresholdKB)
	if err != nil {
		if onProgress != nil {
			onProgress("clone", models.SubJobFailed)
		}
		return nil, err
	}
	if onProgress != nil {
		onProgress("clone", models.SubJobCompleted)
	}

	ref, err := cloneResult.Repo.Head()
	if err != nil {
		return nil, err
	}
	sha := ref.Hash().String()

	if onProgress != nil {
		onProgress("file_walk", models.SubJobRunning)
	}
	files, err := Walk(cloneResult.Repo)
	if err != nil {
		if onProgress != nil {
			onProgress("file_walk", models.SubJobFailed)
		}
		return nil, err
	}
	if onProgress != nil {
		onProgress("file_walk", models.SubJobCompleted)
	}

	if onProgress != nil {
		onProgress("commit_activity_fetch", models.SubJobRunning)
	}
	commitActivity, err := ghClient.FetchCommitActivity(ctx, owner, repo)
	if err != nil {
		if onProgress != nil {
			onProgress("commit_activity_fetch", models.SubJobFailed)
		}
		return nil, err
	}
	if onProgress != nil {
		onProgress("commit_activity_fetch", models.SubJobCompleted)
	}

	languages := aggregateLanguages(files)

	var totalLines, totalCode, totalComment, totalBlank int
	for _, f := range files {
		totalLines += f.TotalLines
		totalCode += f.CodeLines
		totalComment += f.CommentLines
		totalBlank += f.BlankLines
	}

	result := &models.AnalysisResult{
		Repository: models.Repository{
			Owner:         owner,
			Name:          repo,
			FullName:      owner + "/" + repo,
			DefaultBranch: defaultBranch,
			LatestSHA:     sha,
			SizeKB:        sizeKB,
		},
		Metadata:          *meta,
		CommitActivity:    *commitActivity,
		Files:             files,
		Languages:         languages,
		TotalFiles:        len(files),
		TotalLines:        totalLines,
		TotalCodeLines:    totalCode,
		TotalCommentLines: totalComment,
		TotalBlankLines:   totalBlank,
		AnalyzedAt:        time.Now().UTC(),
		Cached:            false,
		CachedAt:          nil,
	}

	return result, nil
}

func aggregateLanguages(files []models.FileStats) []models.LanguageBreakdown {
	langMap := make(map[string]*models.LanguageBreakdown)

	var grandTotal int
	for _, f := range files {
		grandTotal += f.TotalLines
	}

	for _, f := range files {
		lb, ok := langMap[f.Language]
		if !ok {
			lb = &models.LanguageBreakdown{Language: f.Language}
			langMap[f.Language] = lb
		}
		lb.FileCount++
		lb.TotalLines += f.TotalLines
		lb.CodeLines += f.CodeLines
		lb.CommentLines += f.CommentLines
		lb.BlankLines += f.BlankLines
	}

	languages := make([]models.LanguageBreakdown, 0, len(langMap))
	for _, lb := range langMap {
		if grandTotal > 0 {
			lb.Percentage = float64(lb.TotalLines) / float64(grandTotal) * 100.0
		}
		languages = append(languages, *lb)
	}

	sort.Slice(languages, func(i, j int) bool {
		return languages[i].TotalLines > languages[j].TotalLines
	})

	return languages
}
