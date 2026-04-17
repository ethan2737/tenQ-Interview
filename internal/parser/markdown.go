package parser

import (
	"errors"
	"path/filepath"
	"strings"
)

type ParsedDocument struct {
	Path  string
	Title string
	Body  string
}

func ParseMarkdown(path string, markdown string) (ParsedDocument, error) {
	markdown = strings.TrimPrefix(markdown, "\uFEFF")
	lines := strings.Split(markdown, "\n")
	title := ""
	bodyStart := 0

	for index, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "#") {
			title = strings.TrimSpace(strings.TrimLeft(trimmed, "#"))
			title = strings.TrimPrefix(title, "\uFEFF")
			bodyStart = index + 1
			break
		}
	}

	if title == "" {
		base := filepath.Base(path)
		title = strings.TrimSuffix(base, filepath.Ext(base))
	}

	body := strings.TrimSpace(strings.Join(lines[bodyStart:], "\n"))
	if body == "" {
		return ParsedDocument{}, errors.New("markdown body is empty")
	}

	return ParsedDocument{
		Path:  path,
		Title: title,
		Body:  body,
	}, nil
}
