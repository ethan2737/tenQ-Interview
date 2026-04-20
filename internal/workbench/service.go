package workbench

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"tenq-interview/internal/agent"
	"tenq-interview/internal/audio"
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

var titleNumberPattern = regexp.MustCompile(`\d+`)

type DocumentSummary struct {
	Path          string   `json:"path"`
	RelativePath  string   `json:"relativePath"`
	Title         string   `json:"title"`
	Status        string   `json:"status"`
	FromCache     bool     `json:"fromCache"`
	Error         string   `json:"error,omitempty"`
	Encoding      string   `json:"encoding,omitempty"`
	Provider      string   `json:"provider,omitempty"`
	Model         string   `json:"model,omitempty"`
	CardAnswer    string   `json:"cardAnswer,omitempty"`
	MemoryOutline []string `json:"memoryOutline,omitempty"`
	SourceTexts   []string `json:"sourceTexts,omitempty"`
	Notes         string   `json:"notes,omitempty"`
	PromptVersion string   `json:"promptVersion,omitempty"`
	CacheKey      string   `json:"cacheKey,omitempty"`
}

type ImportResult struct {
	Target    string            `json:"target"`
	Total     int               `json:"total"`
	Ready     int               `json:"ready"`
	Failed    int               `json:"failed"`
	Documents []DocumentSummary `json:"documents"`
}

type AgentOption struct {
	Value   string `json:"value"`
	Label   string `json:"label"`
	Enabled bool   `json:"enabled"`
}

type AgentSettings struct {
	DefaultProvider string        `json:"defaultProvider"`
	Options         []AgentOption `json:"options"`
}

type AudioGenerationResult struct {
	OutputPath       string `json:"outputPath"`
	TotalEntries     int    `json:"totalEntries"`
	GeneratedEntries int    `json:"generatedEntries"`
	SkippedEntries   int    `json:"skippedEntries"`
	GeneratedAt      string `json:"generatedAt"`
	Backend          string `json:"backend"`
}

type AudioGenerationStatus struct {
	State            string `json:"state"`
	Message          string `json:"message"`
	Stage            string `json:"stage,omitempty"`
	CurrentQuestion  string `json:"currentQuestion,omitempty"`
	TotalEntries     int    `json:"totalEntries"`
	CompletedEntries int    `json:"completedEntries"`
	OutputPath       string `json:"outputPath,omitempty"`
	StartedAt        string `json:"startedAt,omitempty"`
	UpdatedAt        string `json:"updatedAt,omitempty"`
	FinishedAt       string `json:"finishedAt,omitempty"`
	Error            string `json:"error,omitempty"`
	CanCancel        bool   `json:"canCancel"`
	Backend          string `json:"backend,omitempty"`
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
	processor       *pipeline.Processor
	ruleVersions    cache.RuleVersions
	store           *cache.Store
	cachePath       string
	defaultProvider string
	summarizers     map[string]documentSummarizer
	audioJobs       *audio.JobManager
}

type interviewAudioGenerator interface {
	GenerateFromCache() (audio.Result, error)
}

var newInterviewAudioGenerator = func(config audio.GeneratorConfig) interviewAudioGenerator {
	return audio.NewGenerator(config)
}

type documentSummarizer interface {
	Summarize(ctx context.Context, req agent.SummarizeRequest) (agent.SummarizeResponse, error)
	ProviderModel() string
	PromptVersion() string
}

func NewService() *Service {
	return newService(cache.NewStore(), "")
}

func NewServiceWithCache(cachePath string) (*Service, error) {
	store, err := cache.LoadStore(cachePath)
	if err != nil {
		return nil, err
	}

	return newService(store, cachePath), nil
}

func NewServiceWithOptions(cachePath string, configRoots ...string) (*Service, error) {
	store, err := cache.LoadStore(cachePath)
	if err != nil {
		return nil, err
	}

	service := newService(store, cachePath)
	if err := service.configureAgent(configRoots...); err != nil {
		return nil, err
	}
	return service, nil
}

type markdownExportEntry struct {
	OrderKey string `json:"orderKey,omitempty"`
	Title    string `json:"title"`
	Question string `json:"question"`
	Answer   string `json:"answer"`
}

type MarkdownExportDocument struct {
	Title  string `json:"title"`
	Answer string `json:"answer"`
}

func newService(store *cache.Store, cachePath string) *Service {
	return &Service{
		processor:    pipeline.NewProcessor(),
		ruleVersions: defaultRuleVersions,
		store:        store,
		cachePath:    cachePath,
		summarizers:  map[string]documentSummarizer{},
		audioJobs:    audio.NewJobManager(audio.GeneratorConfig{CachePath: cachePath}),
	}
}

func (s *Service) configureAgent(configRoots ...string) error {
	cfg, err := agent.LoadConfigFromEnv(configRoots...)
	if err != nil {
		var missingKeyErr bool
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}
		switch err.Error() {
		case "deepseek api key is required", "modelscope api key is required":
			missingKeyErr = true
		}
		if missingKeyErr {
			return nil
		}
		return err
	}

	s.defaultProvider = string(cfg.DefaultProvider)
	for _, providerName := range []agent.ProviderName{agent.ProviderDeepSeek, agent.ProviderModelScope} {
		provider, providerErr := agent.NewProvider(providerName, cfg)
		if providerErr != nil {
			continue
		}
		s.summarizers[string(providerName)] = agent.NewSummarizer(provider, agent.PromptVersion)
	}
	return nil
}

func (s *Service) ImportPath(target string) (ImportResult, error) {
	result, err := s.PrepareImport(target)
	if err != nil {
		return ImportResult{}, err
	}

	for index := range result.Documents {
		summary, processErr := s.ProcessDocument(result.Documents[index].Path, result.Documents[index].RelativePath, s.defaultProvider)
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

func (s *Service) ProcessDocument(path string, relativePath string, provider string) (DocumentSummary, error) {
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

	providerName := provider
	if providerName == "" {
		providerName = s.defaultProvider
	}
	modelName, promptVersion := s.cacheScope(providerName)
	cacheKey := cache.BuildCacheKey(path, fingerprint(raw), s.ruleVersions, providerName, modelName, promptVersion)
	if cached, ok := s.store.Get(cacheKey); ok {
		summary.Title = cached.Title
		summary.Status = StatusReady
		summary.FromCache = true
		summary.Encoding = cached.Encoding
		summary.Provider = cached.Provider
		summary.Model = cached.Model
		summary.CardAnswer = cached.CardAnswer
		summary.MemoryOutline = cached.MemoryOutline
		summary.SourceTexts = cached.SourceTexts
		summary.Notes = cached.Notes
		summary.PromptVersion = cached.PromptVersion
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
	summary.Provider = providerName
	summary.Model = modelName
	summary.PromptVersion = promptVersion
	if providerName != "" {
		agentSummary, err := s.summarizeDocument(providerName, processed)
		if err != nil {
			if s.summarizers[providerName] == nil {
				summary.Status = StatusFailed
				summary.Error = err.Error()
				summary.CacheKey = cacheKey
				return summary, nil
			}
			summary.CardAnswer = processed.Card.Answer
			summary.SourceTexts = processed.Card.Sources
			summary.Notes = "agent unavailable, fell back to rule-based summary"
		} else {
			summary.Provider = agentSummary.Provider
			summary.Model = agentSummary.Model
			summary.CardAnswer = agentSummary.StandardAnswer
			summary.MemoryOutline = agentSummary.MemoryOutline
			summary.SourceTexts = agentSummary.SourceQuotes
			summary.Notes = agentSummary.Notes
		}
	} else {
		summary.CardAnswer = processed.Card.Answer
		summary.SourceTexts = processed.Card.Sources
	}
	summary.CacheKey = cacheKey
	s.store.Put(cacheKey, cache.Entry{
		Key:           cacheKey,
		Path:          path,
		Title:         processed.Title,
		Encoding:      processed.Encoding,
		Provider:      summary.Provider,
		Model:         summary.Model,
		CardAnswer:    summary.CardAnswer,
		MemoryOutline: summary.MemoryOutline,
		SourceTexts:   summary.SourceTexts,
		Notes:         summary.Notes,
		PromptVersion: summary.PromptVersion,
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
			Path:          entry.Path,
			RelativePath:  filepath.Base(entry.Path),
			Title:         entry.Title,
			Status:        StatusReady,
			FromCache:     true,
			Encoding:      entry.Encoding,
			Provider:      entry.Provider,
			Model:         entry.Model,
			CardAnswer:    entry.CardAnswer,
			MemoryOutline: entry.MemoryOutline,
			SourceTexts:   entry.SourceTexts,
			Notes:         entry.Notes,
			PromptVersion: entry.PromptVersion,
			CacheKey:      entry.Key,
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

func (s *Service) ExportDocumentMarkdown(title string, answer string, outputPath string) error {
	return s.ExportDocumentsMarkdown([]MarkdownExportDocument{
		{
			Title:  title,
			Answer: answer,
		},
	}, outputPath)
}

func (s *Service) ExportDocumentsMarkdown(documents []MarkdownExportDocument, outputPath string) error {
	if strings.TrimSpace(outputPath) == "" {
		return errors.New("output path is required")
	}

	if len(documents) == 0 {
		return errors.New("at least one document is required")
	}

	newEntries := make([]markdownExportEntry, 0, len(documents))
	for _, document := range documents {
		entry, err := newMarkdownExportEntry(document.Title, document.Answer)
		if err != nil {
			return err
		}
		newEntries = append(newEntries, entry)
	}

	entries, err := loadMarkdownExportEntries(outputPath)
	if err != nil {
		return err
	}

	entryByOrderKey := make(map[string]markdownExportEntry, len(entries)+len(newEntries))
	for _, item := range entries {
		entryByOrderKey[item.OrderKey] = item
	}
	for _, item := range newEntries {
		entryByOrderKey[item.OrderKey] = item
	}

	filtered := make([]markdownExportEntry, 0, len(entryByOrderKey))
	for _, item := range entryByOrderKey {
		filtered = append(filtered, item)
	}

	sort.Slice(filtered, func(i int, j int) bool {
		if compareOrderingKeys(filtered[i].OrderKey, filtered[j].OrderKey) == 0 {
			return filtered[i].Title < filtered[j].Title
		}
		return compareOrderingKeys(filtered[i].OrderKey, filtered[j].OrderKey) < 0
	})

	if err := os.MkdirAll(filepath.Dir(outputPath), 0o755); err != nil {
		return fmt.Errorf("create export directory: %w", err)
	}

	return os.WriteFile(outputPath, buildMarkdownExportDocument(filtered), 0o600)
}

func ExportMarkdownDocumentPath(targetDir string) (string, error) {
	trimmed := strings.TrimSpace(targetDir)
	if trimmed == "" {
		return "", errors.New("export directory is required")
	}
	return filepath.Join(trimmed, "document.md"), nil
}

func (s *Service) AgentSettings() AgentSettings {
	options := []AgentOption{
		{Value: "deepseek", Label: "DeepSeek", Enabled: s.summarizers["deepseek"] != nil},
		{Value: "modelscope", Label: "魔塔", Enabled: s.summarizers["modelscope"] != nil},
	}
	return AgentSettings{
		DefaultProvider: s.defaultProvider,
		Options:         options,
	}
}

func (s *Service) GenerateInterviewAudioFromCache() (AudioGenerationResult, error) {
	if strings.TrimSpace(s.cachePath) == "" {
		return AudioGenerationResult{}, errors.New("cache path is not configured")
	}

	result, err := newInterviewAudioGenerator(audio.GeneratorConfig{
		CachePath: s.cachePath,
	}).GenerateFromCache()
	if err != nil {
		return AudioGenerationResult{}, err
	}

	return AudioGenerationResult{
		OutputPath:       result.OutputPath,
		TotalEntries:     result.TotalEntries,
		GeneratedEntries: result.GeneratedEntries,
		SkippedEntries:   result.SkippedEntries,
		GeneratedAt:      result.GeneratedAt,
		Backend:          result.Backend,
	}, nil
}

func (s *Service) StartInterviewAudioGenerationFromCache() (AudioGenerationStatus, error) {
	if strings.TrimSpace(s.cachePath) == "" {
		return AudioGenerationStatus{}, errors.New("cache path is not configured")
	}

	status, err := s.audioJobs.StartFromCache()
	if err != nil {
		return AudioGenerationStatus{}, err
	}
	return mapAudioGenerationStatus(status), nil
}

func (s *Service) AudioGenerationStatus() AudioGenerationStatus {
	return mapAudioGenerationStatus(s.audioJobs.Status())
}

func (s *Service) CancelInterviewAudioGeneration() (AudioGenerationStatus, error) {
	status, err := s.audioJobs.Cancel()
	if err != nil {
		return AudioGenerationStatus{}, err
	}
	return mapAudioGenerationStatus(status), nil
}

func mapAudioGenerationStatus(status audio.JobStatus) AudioGenerationStatus {
	return AudioGenerationStatus{
		State:            status.State,
		Message:          status.Message,
		Stage:            status.Stage,
		CurrentQuestion:  status.CurrentQuestion,
		TotalEntries:     status.TotalEntries,
		CompletedEntries: status.CompletedEntries,
		OutputPath:       status.OutputPath,
		StartedAt:        status.StartedAt,
		UpdatedAt:        status.UpdatedAt,
		FinishedAt:       status.FinishedAt,
		Error:            status.Error,
		CanCancel:        status.CanCancel,
		Backend:          status.Backend,
	}
}

func (s *Service) cacheScope(provider string) (string, string) {
	summarizer, ok := s.summarizers[provider]
	if !ok || summarizer == nil {
		return "rule-fallback", "rule-fallback"
	}
	return summarizer.ProviderModel(), summarizer.PromptVersion()
}

func (s *Service) summarizeDocument(provider string, processed pipeline.Result) (agent.SummarizeResponse, error) {
	summarizer, ok := s.summarizers[provider]
	if !ok || summarizer == nil {
		return agent.SummarizeResponse{}, errors.New("provider not configured")
	}

	candidateText := make([]string, 0, len(processed.Segments))
	for _, item := range processed.Segments {
		candidateText = append(candidateText, item.Text)
	}

	return summarizer.Summarize(context.Background(), agent.SummarizeRequest{
		Title:         processed.Title,
		Body:          processed.Body,
		CandidateText: candidateText,
	})
}

func fingerprint(raw []byte) string {
	sum := sha256.Sum256(raw)
	return hex.EncodeToString(sum[:])
}

func trimExtension(name string) string {
	return name[:len(name)-len(filepath.Ext(name))]
}

func newMarkdownExportEntry(title string, answer string) (markdownExportEntry, error) {
	trimmedTitle := strings.TrimSpace(title)
	if trimmedTitle == "" {
		return markdownExportEntry{}, errors.New("document title is required")
	}

	trimmedAnswer := strings.TrimSpace(answer)
	if trimmedAnswer == "" {
		return markdownExportEntry{}, errors.New("document answer is empty")
	}

	orderKey, err := extractOrderKey(trimmedTitle)
	if err != nil {
		return markdownExportEntry{}, err
	}

	return markdownExportEntry{
		OrderKey: orderKey,
		Title:    trimmedTitle,
		Question: trimmedTitle,
		Answer:   trimmedAnswer,
	}, nil
}

func loadMarkdownExportEntries(outputPath string) ([]markdownExportEntry, error) {
	raw, err := os.ReadFile(outputPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, nil
		}
		return nil, fmt.Errorf("read export file: %w", err)
	}

	content := strings.TrimSpace(string(raw))
	if content == "" {
		return nil, nil
	}

	var entries []markdownExportEntry
	for _, line := range strings.Split(content, "\n") {
		line = strings.TrimSpace(line)
		if !strings.HasPrefix(line, "<!-- TENQ_EXPORT_ENTRY ") || !strings.HasSuffix(line, " -->") {
			continue
		}

		payload := strings.TrimSuffix(strings.TrimPrefix(line, "<!-- TENQ_EXPORT_ENTRY "), " -->")
		decoded, err := base64.StdEncoding.DecodeString(payload)
		if err != nil {
			return nil, fmt.Errorf("decode export entry: %w", err)
		}

		var entry markdownExportEntry
		if err := json.Unmarshal(decoded, &entry); err != nil {
			return nil, fmt.Errorf("parse export entry: %w", err)
		}
		if entry.OrderKey == "" {
			orderKey, orderErr := extractOrderKey(entry.Title)
			if orderErr != nil {
				return nil, orderErr
			}
			entry.OrderKey = orderKey
		}
		entries = append(entries, entry)
	}

	if len(entries) > 0 {
		return entries, nil
	}

	return parseMarkdownExportEntries(content)
}

func buildMarkdownExportDocument(entries []markdownExportEntry) []byte {
	var builder bytes.Buffer
	builder.WriteString("# 面试整理导出\n\n")
	builder.WriteString("> 由 TenQ Interview 导出，所有题目按标题中的序号排序。\n\n")

	for index, entry := range entries {
		builder.WriteString("## ")
		builder.WriteString(fmt.Sprintf("%d. %s", displayOrderIndex(entry.OrderKey), entry.Title))
		builder.WriteString("\n\n")
		builder.WriteString("**问题**\n\n")
		builder.WriteString(entry.Question)
		builder.WriteString("\n\n")
		builder.WriteString("**答案**\n\n")
		builder.WriteString(entry.Answer)
		if index < len(entries)-1 {
			builder.WriteString("\n\n")
		}
	}

	builder.WriteString("\n")
	return builder.Bytes()
}

func extractOrderKey(title string) (string, error) {
	matches := titleNumberPattern.FindAllString(strings.TrimSpace(title), -1)
	if len(matches) == 0 {
		return "", errors.New("document title must contain an ordering number")
	}

	parts := make([]string, 0, len(matches))
	for _, match := range matches {
		var value int
		if _, err := fmt.Sscanf(match, "%d", &value); err != nil || value <= 0 {
			return "", errors.New("document title contains an invalid ordering number")
		}
		parts = append(parts, fmt.Sprintf("%d", value))
	}
	return strings.Join(parts, "."), nil
}

func compareOrderingKeys(left string, right string) int {
	leftParts := titleNumberPattern.FindAllString(left, -1)
	rightParts := titleNumberPattern.FindAllString(right, -1)

	limit := len(leftParts)
	if len(rightParts) < limit {
		limit = len(rightParts)
	}

	for index := 0; index < limit; index++ {
		var leftValue int
		var rightValue int
		fmt.Sscanf(leftParts[index], "%d", &leftValue)
		fmt.Sscanf(rightParts[index], "%d", &rightValue)
		switch {
		case leftValue < rightValue:
			return -1
		case leftValue > rightValue:
			return 1
		}
	}

	switch {
	case len(leftParts) < len(rightParts):
		return -1
	case len(leftParts) > len(rightParts):
		return 1
	default:
		return strings.Compare(left, right)
	}
}

func displayOrderIndex(orderKey string) int {
	parts := titleNumberPattern.FindAllString(orderKey, -1)
	if len(parts) == 0 {
		return 0
	}

	var value int
	fmt.Sscanf(parts[0], "%d", &value)
	return value
}

func parseMarkdownExportEntries(content string) ([]markdownExportEntry, error) {
	sections := strings.Split(content, "\n## ")
	if len(sections) == 0 {
		return nil, errors.New("existing export file is not a TenQ markdown export")
	}

	entries := make([]markdownExportEntry, 0, len(sections))
	for index, section := range sections {
		if index == 0 {
			headerIndex := strings.Index(section, "## ")
			if headerIndex == -1 {
				continue
			}
			section = section[headerIndex+3:]
		}

		section = strings.TrimSpace(section)
		if section == "" {
			continue
		}

		parts := strings.SplitN(section, "\n\n**问题**\n\n", 2)
		if len(parts) != 2 {
			continue
		}

		answerParts := strings.SplitN(parts[1], "\n\n**答案**\n\n", 2)
		if len(answerParts) != 2 {
			continue
		}

		question := strings.TrimSpace(answerParts[0])
		answer := strings.TrimSpace(answerParts[1])
		title := question
		if title == "" {
			title = strings.TrimSpace(parts[0])
		}

		entry, err := newMarkdownExportEntry(title, answer)
		if err != nil {
			return nil, err
		}
		if question != "" {
			entry.Question = question
		}
		entries = append(entries, entry)
	}

	if len(entries) == 0 {
		return nil, errors.New("existing export file is not a TenQ markdown export")
	}

	return entries, nil
}
