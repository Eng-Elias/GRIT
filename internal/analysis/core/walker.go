package core

import (
	"bytes"
	"io"
	"path/filepath"
	"strings"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"

	"github.com/grit-app/grit/internal/models"
)

func Walk(repo *git.Repository) ([]models.FileStats, error) {
	ref, err := repo.Head()
	if err != nil {
		return nil, err
	}

	commit, err := repo.CommitObject(ref.Hash())
	if err != nil {
		return nil, err
	}

	tree, err := commit.Tree()
	if err != nil {
		return nil, err
	}

	var files []models.FileStats

	err = tree.Files().ForEach(func(f *object.File) error {
		if shouldSkip(f.Name) {
			return nil
		}

		reader, err := f.Reader()
		if err != nil {
			return nil
		}
		defer reader.Close()

		content, err := io.ReadAll(reader)
		if err != nil {
			return nil
		}

		ext := strings.ToLower(filepath.Ext(f.Name))
		filename := filepath.Base(f.Name)
		lang := LookupLanguage(ext, filename)

		fs := models.FileStats{
			Path:     f.Name,
			Language: lang.Name,
			ByteSize: f.Size,
		}

		if !IsBinary(content[:min(len(content), 512)]) {
			counts := CountLines(bytes.NewReader(content), lang)
			fs.TotalLines = counts.Total
			fs.CodeLines = counts.Code
			fs.CommentLines = counts.Comment
			fs.BlankLines = counts.Blank
		}

		files = append(files, fs)
		return nil
	})

	if err != nil {
		return nil, err
	}

	return files, nil
}

func shouldSkip(path string) bool {
	lower := strings.ToLower(filepath.Base(path))
	skipExts := []string{".png", ".jpg", ".jpeg", ".gif", ".bmp", ".ico", ".svg",
		".webp", ".mp3", ".mp4", ".wav", ".avi", ".mov",
		".zip", ".tar", ".gz", ".bz2", ".7z", ".rar",
		".pdf", ".doc", ".docx", ".xls", ".xlsx",
		".exe", ".dll", ".so", ".dylib", ".bin",
		".woff", ".woff2", ".ttf", ".eot", ".otf",
		".lock"}

	ext := strings.ToLower(filepath.Ext(lower))
	for _, skip := range skipExts {
		if ext == skip {
			return true
		}
	}
	return false
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
