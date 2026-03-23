package core

import (
	"bufio"
	"io"
	"strings"
)

type LineCounts struct {
	Total   int
	Code    int
	Comment int
	Blank   int
}

func IsBinary(data []byte) bool {
	for _, b := range data {
		if b == 0 {
			return true
		}
	}
	return false
}

func CountLines(r io.Reader, lang LanguageDef) LineCounts {
	var counts LineCounts
	inBlock := false

	scanner := bufio.NewScanner(r)
	buf := make([]byte, 0, 64*1024)
	scanner.Buffer(buf, 1024*1024)

	for scanner.Scan() {
		counts.Total++
		line := scanner.Text()
		trimmed := strings.TrimSpace(line)

		if trimmed == "" {
			counts.Blank++
			continue
		}

		if inBlock {
			counts.Comment++
			if lang.BlockCommentEnd != "" && strings.Contains(trimmed, lang.BlockCommentEnd) {
				inBlock = false
			}
			continue
		}

		if lang.BlockCommentStart != "" && strings.HasPrefix(trimmed, lang.BlockCommentStart) {
			counts.Comment++
			if lang.BlockCommentEnd != "" && !strings.Contains(trimmed[len(lang.BlockCommentStart):], lang.BlockCommentEnd) {
				inBlock = true
			}
			continue
		}

		isComment := false
		for _, prefix := range lang.LineCommentPrefixes {
			if strings.HasPrefix(trimmed, prefix) {
				isComment = true
				break
			}
		}

		if isComment {
			counts.Comment++
		} else {
			counts.Code++
		}
	}

	return counts
}
