package workbench

import (
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path/filepath"
	"sort"

	"tenq-interview/internal/cache"
	"tenq-interview/internal/importer"
	"tenq-interview/internal/library"
	"tenq-interview/internal/parser"
	"tenq-interview/internal/pipeline"
)

const (
	StatusPending = "pending"
	StatusReady   = "ready"
	StatusFailed  = "failed"
)

var defaultRuleVersions = cache.RuleVersions{
	ParserVersion:    "v1",
	SegmentVersion:   "v1",
	GeneratorVersion: "v1",
}

type DocumentSummary struct {
	Path         string   `json:"path"`
	RelativePath string   `json:"relativePath"`
	Title        string   `json:"title"`
	Status       string   `json:"status"`
	FromCache    bool     `json:"fromCache"`
	Error        string   `json:"error,omitempty"`
	Encoding     string   `json:"encoding,omitempty"`
	CardAnswer   string   `json:"cardAnswer,omitempty"`
	SourceTexts  []string `json:"sourceTexts,omitempty"`
	CacheKey     string   `json:"cacheKey,omitempty"`
}

type ImportResult struct {
	Target    string            `json:"target"`
	Total     int               `json:"total"`
	Ready     int               `json:"ready"`
	Failed    int               `json:"failed"`
	Documents []DocumentSummary `json:"documents"`
}

type DocumentPreview struct {
	Path             string `json:"path"`
	Title            string `json:"title"`
	Encoding         string `json:"encoding"`
	Fingerprint      string `json:"fingerprint"`
	NormalizedBody   string `json:"normalizedBody"`
	SuspectedGarbled bool   `json:"suspectedGarbled"`
	Warning          string `json:"warning,omitempty"`
}

type Service struct {
	processor    *pipeline.Processor
	ruleVersions cache.RuleVersions
	store        *cache.Store
	cachePath    string
}

func NewService() *Service {
	return &Service{
		processor:    pipeline.NewProcessor(),
		ruleVersions: defaultRuleVersions,
		store:        cache.NewStore(),
	}
}

func NewServiceWithCache(cachePath string) (*Service, error) {
	store, err := cache.LoadStore(cachePath)
	if err != nil {
		return nil, err
	}

	return &Service{
		processor:    pipeline.NewProcessor(),
		ruleVersions: defaultRuleVersions,
		store:        store,
		cachePath:    cachePath,
	}, nil
}

func (s *Service) ImportPath(target string) (ImportResult, error) {
	result, err := s.PrepareImport(target)
	if err != nil {
		return ImportResult{}, err
	}

	for index := range result.Documents {
		summary, processErr := s.ProcessDocument(result.Documents[index].Path, result.Documents[index].RelativePath)
		if processErr != nil {
			return ImportResult{}, processErr
		}
		result.Documents[index] = summary
		switch summary.Status {
		case StatusReady:
			result.Ready++
		case StatusFailed:
			result.Failed++
		}
	}
	return result, nil
}

func (s *Service) PrepareImport(target string) (ImportResult, error) {
	entries, err := library.ScanMarkdownPaths(target)
	if err != nil {
		return ImportResult{}, err
	}

	result := ImportResult{
		Target:    target,
		Total:     len(entries),
		Documents: make([]DocumentSummary, 0, len(entries)),
	}
	for _, entry := range entries {
		result.Documents = append(result.Documents, DocumentSummary{
			Path:         entry.Path,
			RelativePath: entry.RelativePath,
			Title:        trimExtension(filepath.Base(entry.Path)),
			Status:       StatusPending,
		})
	}
	return result, nil
}

func (s *Service) ProcessDocument(path string, relativePath string) (DocumentSummary, error) {
	summary := DocumentSummary{
		Path:         path,
		RelativePath: relativePath,
		Title:        trimExtension(filepath.Base(path)),
	}

	raw, readErr := os.ReadFile(path)
	if readErr != nil {
		summary.Status = StatusFailed
		summary.Error = readErr.Error()
		return summary, nil
	}

	cacheKey := cache.BuildCacheKey(path, fingerprint(raw), s.ruleVersions)
	if cached, ok := s.store.Get(cacheKey); ok {
		summary.Title = cached.Title
		summary.Status = StatusReady
		summary.FromCache = true
		summary.Encoding = cached.Encoding
		summary.CardAnswer = cached.CardAnswer
		summary.SourceTexts = cached.SourceTexts
		summary.CacheKey = cacheKey
		return summary, nil
	}

	processed, processErr := s.processor.ProcessFile(path)
	if processErr != nil {
		summary.Status = StatusFailed
		summary.Error = processErr.Error()
		summary.CacheKey = cacheKey
		return summary, nil
	}

	summary.Title = processed.Title
	summary.Status = StatusReady
	summary.Encoding = processed.Encoding
	summary.CardAnswer = processed.Card.Answer
	summary.SourceTexts = processed.Card.Sources
	summary.CacheKey = cacheKey
	s.store.Put(cacheKey, cache.Entry{
		Key:         cacheKey,
		Path:        path,
		Title:       processed.Title,
		Encoding:    processed.Encoding,
		CardAnswer:  processed.Card.Answer,
		SourceTexts: processed.Card.Sources,
	})

	if s.cachePath != "" {
		if err := s.store.Save(s.cachePath); err != nil {
			return DocumentSummary{}, err
		}
	}

	return summary, nil
}

func (s *Service) PreviewDocument(path string) (DocumentPreview, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return DocumentPreview{}, err
	}

	normalized, encoding, err := importer.NormalizeMarkdownBytes(raw)
	if err != nil {
		return DocumentPreview{}, err
	}

	parsed, err := parser.ParseMarkdown(path, normalized)
	if err != nil {
		return DocumentPreview{}, err
	}

	suspected, warning := importer.DetectLikelyGarbledText(parsed.Title + "\n" + parsed.Body)

	return DocumentPreview{
		Path:             path,
		Title:            parsed.Title,
		Encoding:         encoding,
		Fingerprint:      fingerprint(raw),
		NormalizedBody:   parsed.Body,
		SuspectedGarbled: suspected,
		Warning:          warning,
	}, nil
}

func (s *Service) ListImportedDocuments() (ImportResult, error) {
	entries := s.store.List()
	byPath := make(map[string]DocumentSummary, len(entries))

	for _, entry := range entries {
		byPath[entry.Path] = DocumentSummary{
			Path:         entry.Path,
			RelativePath: filepath.Base(entry.Path),
			Title:        entry.Title,
			Status:       StatusReady,
			FromCache:    true,
			Encoding:     entry.Encoding,
			CardAnswer:   entry.CardAnswer,
			SourceTexts:  entry.SourceTexts,
			CacheKey:     entry.Key,
		}
	}

	documents := make([]DocumentSummary, 0, len(byPath))
	for _, document := range byPath {
		documents = append(documents, document)
	}

	sort.Slice(documents, func(i int, j int) bool {
		if documents[i].Title == documents[j].Title {
			return documents[i].Path < documents[j].Path
		}
		return documents[i].Title < documents[j].Title
	})

	return ImportResult{
		Target:    "累计导入",
		Total:     len(documents),
		Ready:     len(documents),
		Failed:    0,
		Documents: documents,
	}, nil
}

func (s *Service) ClearImportedDocuments() error {
	s.store.Clear()
	if s.cachePath == "" {
		return nil
	}
	return s.store.Save(s.cachePath)
}

func fingerprint(raw []byte) string {
	sum := sha256.Sum256(raw)
	return hex.EncodeToString(sum[:])
}

func trimExtension(name string) string {
	return name[:len(name)-len(filepath.Ext(name))]
}
