package library

import (
	"errors"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type Entry struct {
	Path         string
	RelativePath string
}

func ScanMarkdownPaths(target string) ([]Entry, error) {
	info, err := os.Stat(target)
	if err != nil {
		return nil, err
	}

	if !info.IsDir() {
		if !isMarkdownFile(target) {
			return nil, errors.New("target is not a markdown file")
		}
		return []Entry{{
			Path:         target,
			RelativePath: filepath.Base(target),
		}}, nil
	}

	var entries []Entry
	err = filepath.WalkDir(target, func(path string, dirEntry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if dirEntry.IsDir() {
			return nil
		}
		if !isMarkdownFile(path) {
			return nil
		}

		relativePath, err := filepath.Rel(target, path)
		if err != nil {
			return err
		}

		entries = append(entries, Entry{
			Path:         path,
			RelativePath: filepath.ToSlash(relativePath),
		})
		return nil
	})
	if err != nil {
		return nil, err
	}

	sort.Slice(entries, func(i, j int) bool {
		return entries[i].RelativePath < entries[j].RelativePath
	})
	return entries, nil
}

func isMarkdownFile(path string) bool {
	ext := strings.ToLower(filepath.Ext(path))
	return ext == ".md"
}
