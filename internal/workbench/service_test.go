package workbench

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"tenq-interview/internal/agent"
	"tenq-interview/internal/audio"

	"golang.org/x/text/encoding/simplifiedchinese"
)

type stubSummarizer struct {
	response agent.SummarizeResponse
	err      error
	model    string
	version  string
}

func (s stubSummarizer) Summarize(ctx context.Context, req agent.SummarizeRequest) (agent.SummarizeResponse, error) {
	if s.err != nil {
		return agent.SummarizeResponse{}, s.err
	}
	return s.response, nil
}

func (s stubSummarizer) ProviderModel() string { return s.model }
func (s stubSummarizer) PromptVersion() string { return s.version }

type stubInterviewAudioGenerator struct {
	result audio.Result
	err    error
}

func (s stubInterviewAudioGenerator) GenerateFromCache() (audio.Result, error) {
	return s.result, s.err
}

func TestImportPathBuildsDocumentSummaries(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	docsDir := filepath.Join(root, "docs")
	if err := os.MkdirAll(docsDir, 0o755); err != nil {
		t.Fatalf("failed to create docs dir: %v", err)
	}

	validPath := filepath.Join(docsDir, "gmp.md")
	emptyPath := filepath.Join(docsDir, "empty.md")
	if err := os.WriteFile(validPath, []byte("# Go 的 GMP 模型是什么？\n\nGMP 是 Go 的调度模型。它把大量 goroutine 映射到更少的线程上执行。"), 0o600); err != nil {
		t.Fatalf("failed to write valid fixture: %v", err)
	}
	if err := os.WriteFile(emptyPath, []byte("# 空文档"), 0o600); err != nil {
		t.Fatalf("failed to write empty fixture: %v", err)
	}

	service := NewService()
	result, err := service.ImportPath(docsDir)
	if err != nil {
		t.Fatalf("ImportPath returned error: %v", err)
	}

	if result.Total != 2 {
		t.Fatalf("expected total 2, got %d", result.Total)
	}
	if result.Ready != 1 {
		t.Fatalf("expected ready count 1, got %d", result.Ready)
	}
	if result.Failed != 1 {
		t.Fatalf("expected failed count 1, got %d", result.Failed)
	}
	if len(result.Documents) != 2 {
		t.Fatalf("expected 2 documents, got %d", len(result.Documents))
	}

	first := result.Documents[0]
	if first.Title == "" {
		t.Fatalf("expected first document title")
	}
	if first.Status != StatusFailed {
		t.Fatalf("expected first document to fail, got %s", first.Status)
	}
	if first.Error == "" {
		t.Fatalf("expected failed document error")
	}

	second := result.Documents[1]
	if second.Status != StatusReady {
		t.Fatalf("expected second document to be ready, got %s", second.Status)
	}
	if second.CardAnswer == "" {
		t.Fatalf("expected generated card answer")
	}
	if second.CacheKey == "" {
		t.Fatalf("expected cache key")
	}
	if second.Encoding == "" {
		t.Fatalf("expected encoding metadata")
	}
}

func TestPrepareImportListsPendingDocuments(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	docsDir := filepath.Join(root, "docs")
	if err := os.MkdirAll(docsDir, 0o755); err != nil {
		t.Fatalf("failed to create docs dir: %v", err)
	}

	firstPath := filepath.Join(docsDir, "gmp.md")
	secondPath := filepath.Join(docsDir, "memory.md")
	if err := os.WriteFile(firstPath, []byte("# GMP\n\nBody"), 0o600); err != nil {
		t.Fatalf("failed to write first fixture: %v", err)
	}
	if err := os.WriteFile(secondPath, []byte("# Memory\n\nBody"), 0o600); err != nil {
		t.Fatalf("failed to write second fixture: %v", err)
	}

	service := NewService()
	result, err := service.PrepareImport(docsDir)
	if err != nil {
		t.Fatalf("PrepareImport returned error: %v", err)
	}

	if result.Total != 2 {
		t.Fatalf("expected total 2, got %d", result.Total)
	}
	if len(result.Documents) != 2 {
		t.Fatalf("expected 2 documents, got %d", len(result.Documents))
	}
	if result.Documents[0].Status != StatusPending {
		t.Fatalf("expected pending status, got %s", result.Documents[0].Status)
	}
	if result.Documents[0].RelativePath == "" {
		t.Fatalf("expected relative path to be populated")
	}
}

func TestProcessDocumentUsesCacheOnSecondRun(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	cachePath := filepath.Join(root, "cache.json")
	docPath := filepath.Join(root, "gmp.md")
	content := "# Go 的 GMP 模型是什么？\n\nGMP 是 Go 的调度模型。"
	if err := os.WriteFile(docPath, []byte(content), 0o600); err != nil {
		t.Fatalf("failed to write fixture: %v", err)
	}

	service, err := NewServiceWithCache(cachePath)
	if err != nil {
		t.Fatalf("NewServiceWithCache returned error: %v", err)
	}

	first, err := service.ProcessDocument(docPath, "gmp.md", "")
	if err != nil {
		t.Fatalf("ProcessDocument returned error: %v", err)
	}
	if first.FromCache {
		t.Fatalf("expected first run to be fresh")
	}

	second, err := service.ProcessDocument(docPath, "gmp.md", "")
	if err != nil {
		t.Fatalf("ProcessDocument returned error: %v", err)
	}
	if !second.FromCache {
		t.Fatalf("expected second run to hit cache")
	}
}

func TestProcessDocumentCanRecoverAfterFailure(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	cachePath := filepath.Join(root, "cache.json")
	docPath := filepath.Join(root, "broken.md")
	if err := os.WriteFile(docPath, []byte("# 空文档"), 0o600); err != nil {
		t.Fatalf("failed to write broken fixture: %v", err)
	}

	service, err := NewServiceWithCache(cachePath)
	if err != nil {
		t.Fatalf("NewServiceWithCache returned error: %v", err)
	}

	first, err := service.ProcessDocument(docPath, "broken.md", "")
	if err != nil {
		t.Fatalf("ProcessDocument returned error: %v", err)
	}
	if first.Status != StatusFailed {
		t.Fatalf("expected first run to fail, got %s", first.Status)
	}

	if err := os.WriteFile(docPath, []byte("# 修复后文档\n\n现在已经有正文，可以重新生成题卡。"), 0o600); err != nil {
		t.Fatalf("failed to rewrite fixture: %v", err)
	}

	second, err := service.ProcessDocument(docPath, "broken.md", "")
	if err != nil {
		t.Fatalf("ProcessDocument returned error: %v", err)
	}
	if second.Status != StatusReady {
		t.Fatalf("expected second run to recover, got %s", second.Status)
	}
	if second.CardAnswer == "" {
		t.Fatalf("expected recovered document to produce an answer")
	}
}

func TestPreviewDocumentNormalizesGB18030Content(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	docPath := filepath.Join(root, "preview.md")
	raw := "# 预览文档\r\n\r\n第一行内容\r\n第二行内容"
	encoded, err := simplifiedchinese.GB18030.NewEncoder().Bytes([]byte(raw))
	if err != nil {
		t.Fatalf("failed to encode fixture: %v", err)
	}
	if err := os.WriteFile(docPath, encoded, 0o600); err != nil {
		t.Fatalf("failed to write encoded fixture: %v", err)
	}

	service := NewService()
	preview, err := service.PreviewDocument(docPath)
	if err != nil {
		t.Fatalf("PreviewDocument returned error: %v", err)
	}

	if preview.Encoding != "gb18030" {
		t.Fatalf("expected gb18030 encoding, got %q", preview.Encoding)
	}
	if preview.Title != "预览文档" {
		t.Fatalf("unexpected title: %q", preview.Title)
	}
	if preview.NormalizedBody != "第一行内容\n第二行内容" {
		t.Fatalf("unexpected normalized body: %q", preview.NormalizedBody)
	}
	if preview.SuspectedGarbled {
		t.Fatalf("expected normalized preview to be clean")
	}
}

func TestPreviewDocumentFlagsLikelyGarbledText(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	docPath := filepath.Join(root, "garbled.md")
	content := "# Go 鐨勭嚎绋嬫ā鍨嬫槸浠€涔堬紵\n\n杩欐槸涓€娈电枒浼间贡鐮佺殑鍐呭銆"
	if err := os.WriteFile(docPath, []byte(content), 0o600); err != nil {
		t.Fatalf("failed to write garbled fixture: %v", err)
	}

	service := NewService()
	preview, err := service.PreviewDocument(docPath)
	if err != nil {
		t.Fatalf("PreviewDocument returned error: %v", err)
	}

	if !preview.SuspectedGarbled {
		t.Fatalf("expected garbled preview to be flagged")
	}
	if preview.Warning == "" {
		t.Fatalf("expected warning message to be populated")
	}
}

func TestImportPathReusesPersistentCache(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	cachePath := filepath.Join(root, "cache.json")
	docPath := filepath.Join(root, "gmp.md")
	content := "# Go 的 GMP 模型是什么？\n\nGMP 是 Go 的调度模型。它把大量 goroutine 映射到更少的线程上执行。"
	if err := os.WriteFile(docPath, []byte(content), 0o600); err != nil {
		t.Fatalf("failed to write fixture: %v", err)
	}

	firstService, err := NewServiceWithCache(cachePath)
	if err != nil {
		t.Fatalf("NewServiceWithCache returned error: %v", err)
	}
	firstResult, err := firstService.ImportPath(docPath)
	if err != nil {
		t.Fatalf("ImportPath returned error: %v", err)
	}
	if len(firstResult.Documents) != 1 {
		t.Fatalf("expected 1 document, got %d", len(firstResult.Documents))
	}
	if firstResult.Documents[0].FromCache {
		t.Fatalf("expected first import to be freshly processed")
	}

	secondService, err := NewServiceWithCache(cachePath)
	if err != nil {
		t.Fatalf("NewServiceWithCache returned error: %v", err)
	}
	secondResult, err := secondService.ImportPath(docPath)
	if err != nil {
		t.Fatalf("ImportPath returned error: %v", err)
	}
	if !secondResult.Documents[0].FromCache {
		t.Fatalf("expected second import to hit persistent cache")
	}
}

func TestListImportedDocumentsRestoresCachedLibrary(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	cachePath := filepath.Join(root, "cache.json")

	firstPath := filepath.Join(root, "gmp.md")
	secondPath := filepath.Join(root, "channel.md")
	if err := os.WriteFile(firstPath, []byte("# GMP\n\nGMP 是 Go 的调度模型。"), 0o600); err != nil {
		t.Fatalf("failed to write first fixture: %v", err)
	}
	if err := os.WriteFile(secondPath, []byte("# Channel\n\nChannel 用于协程通信。"), 0o600); err != nil {
		t.Fatalf("failed to write second fixture: %v", err)
	}

	firstService, err := NewServiceWithCache(cachePath)
	if err != nil {
		t.Fatalf("NewServiceWithCache returned error: %v", err)
	}
	if _, err := firstService.ProcessDocument(firstPath, "gmp.md", ""); err != nil {
		t.Fatalf("ProcessDocument firstPath returned error: %v", err)
	}
	if _, err := firstService.ProcessDocument(secondPath, "channel.md", ""); err != nil {
		t.Fatalf("ProcessDocument secondPath returned error: %v", err)
	}

	secondService, err := NewServiceWithCache(cachePath)
	if err != nil {
		t.Fatalf("NewServiceWithCache returned error: %v", err)
	}
	result, err := secondService.ListImportedDocuments()
	if err != nil {
		t.Fatalf("ListImportedDocuments returned error: %v", err)
	}

	if result.Total != 2 || result.Ready != 2 || result.Failed != 0 {
		t.Fatalf("unexpected counts: total=%d ready=%d failed=%d", result.Total, result.Ready, result.Failed)
	}
	if len(result.Documents) != 2 {
		t.Fatalf("expected 2 restored documents, got %d", len(result.Documents))
	}
	if !result.Documents[0].FromCache || !result.Documents[1].FromCache {
		t.Fatalf("expected restored documents to be marked from cache")
	}
}

func TestClearImportedDocumentsRemovesPersistentLibrary(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	cachePath := filepath.Join(root, "cache.json")
	docPath := filepath.Join(root, "gmp.md")
	if err := os.WriteFile(docPath, []byte("# GMP\n\nGMP 是 Go 的调度模型。"), 0o600); err != nil {
		t.Fatalf("failed to write fixture: %v", err)
	}

	firstService, err := NewServiceWithCache(cachePath)
	if err != nil {
		t.Fatalf("NewServiceWithCache returned error: %v", err)
	}
	if _, err := firstService.ProcessDocument(docPath, "gmp.md", ""); err != nil {
		t.Fatalf("ProcessDocument returned error: %v", err)
	}

	if err := firstService.ClearImportedDocuments(); err != nil {
		t.Fatalf("ClearImportedDocuments returned error: %v", err)
	}

	secondService, err := NewServiceWithCache(cachePath)
	if err != nil {
		t.Fatalf("NewServiceWithCache returned error: %v", err)
	}
	result, err := secondService.ListImportedDocuments()
	if err != nil {
		t.Fatalf("ListImportedDocuments returned error: %v", err)
	}
	if result.Total != 0 || len(result.Documents) != 0 {
		t.Fatalf("expected cleared library to stay empty, got total=%d docs=%d", result.Total, len(result.Documents))
	}
}

func TestProcessDocumentStoresAgentFieldsAndProvider(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	docPath := filepath.Join(root, "gmp.md")
	if err := os.WriteFile(docPath, []byte("# GMP\n\nGMP 是 Go 的调度模型。"), 0o600); err != nil {
		t.Fatalf("failed to write fixture: %v", err)
	}

	service := NewService()
	service.defaultProvider = "deepseek"
	service.summarizers["deepseek"] = stubSummarizer{
		model:   "deepseek-chat",
		version: agent.PromptVersion,
		response: agent.SummarizeResponse{
			Provider:       "deepseek",
			Model:          "deepseek-chat",
			StandardAnswer: "GMP 是 Go 的调度模型，核心是在用户态高效调度 goroutine 到更少的线程上执行。",
			MemoryOutline:  []string{"定义", "调度对象", "价值"},
			SourceQuotes:   []string{"GMP 是 Go 的调度模型。"},
			Notes:          "严格基于原文整理",
		},
	}

	got, err := service.ProcessDocument(docPath, "gmp.md", "deepseek")
	if err != nil {
		t.Fatalf("ProcessDocument returned error: %v", err)
	}
	if got.Provider != "deepseek" {
		t.Fatalf("expected provider to be persisted, got %q", got.Provider)
	}
	if got.Model != "deepseek-chat" {
		t.Fatalf("expected model to be persisted, got %q", got.Model)
	}
	if len(got.MemoryOutline) == 0 || len(got.SourceTexts) == 0 {
		t.Fatalf("expected agent fields to be populated")
	}
	if got.PromptVersion != agent.PromptVersion {
		t.Fatalf("unexpected prompt version: %q", got.PromptVersion)
	}
}

func TestProcessDocumentFallsBackToRuleSummaryWhenAgentFails(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	docPath := filepath.Join(root, "gmp.md")
	if err := os.WriteFile(docPath, []byte("# GMP\n\nGMP 是 Go 的调度模型，用来高效调度 goroutine。"), 0o600); err != nil {
		t.Fatalf("failed to write fixture: %v", err)
	}

	service := NewService()
	service.defaultProvider = "deepseek"
	service.summarizers["deepseek"] = stubSummarizer{
		model:   "deepseek-chat",
		version: agent.PromptVersion,
		err:     os.ErrDeadlineExceeded,
	}

	got, err := service.ProcessDocument(docPath, "gmp.md", "deepseek")
	if err != nil {
		t.Fatalf("ProcessDocument returned error: %v", err)
	}
	if got.Status != StatusReady {
		t.Fatalf("expected fallback result to stay ready, got %s", got.Status)
	}
	if got.Error != "" {
		t.Fatalf("expected fallback result without error, got %q", got.Error)
	}
	if got.CardAnswer == "" {
		t.Fatalf("expected fallback rule answer to be populated")
	}
	if got.Provider != "deepseek" {
		t.Fatalf("expected requested provider to be kept, got %q", got.Provider)
	}
	if got.Model != "deepseek-chat" {
		t.Fatalf("expected provider model to be kept, got %q", got.Model)
	}
	if got.PromptVersion != agent.PromptVersion {
		t.Fatalf("expected prompt version to be kept, got %q", got.PromptVersion)
	}
}

func TestExportDocumentMarkdownCreatesSortedDocument(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	outputPath := filepath.Join(root, "document.md")
	service := NewService()

	if err := service.ExportDocumentMarkdown("第3题 GMP 是什么？", "第三题答案", outputPath); err != nil {
		t.Fatalf("first export returned error: %v", err)
	}
	if err := service.ExportDocumentMarkdown("第1题 Channel 是什么？", "第一题答案", outputPath); err != nil {
		t.Fatalf("second export returned error: %v", err)
	}

	raw, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("failed to read export file: %v", err)
	}

	content := string(raw)
	firstIndex := strings.Index(content, "## 1. 第1题 Channel 是什么？")
	secondIndex := strings.Index(content, "## 3. 第3题 GMP 是什么？")
	if firstIndex == -1 || secondIndex == -1 {
		t.Fatalf("expected exported entries to exist, got %q", content)
	}
	if firstIndex > secondIndex {
		t.Fatalf("expected entries to be sorted by title number, got %q", content)
	}
	if !strings.Contains(content, "**答案**\n\n第一题答案") {
		t.Fatalf("expected first answer to be rendered, got %q", content)
	}
	if !strings.Contains(content, "**答案**\n\n第三题答案") {
		t.Fatalf("expected second answer to be rendered, got %q", content)
	}
}

func TestExportDocumentMarkdownReplacesExistingEntryWithSameIndex(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	outputPath := filepath.Join(root, "document.md")
	service := NewService()

	if err := service.ExportDocumentMarkdown("第2题 原始标题", "旧答案", outputPath); err != nil {
		t.Fatalf("first export returned error: %v", err)
	}
	if err := service.ExportDocumentMarkdown("第2题 更新标题", "新答案", outputPath); err != nil {
		t.Fatalf("second export returned error: %v", err)
	}

	raw, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("failed to read export file: %v", err)
	}

	content := string(raw)
	if strings.Contains(content, "旧答案") {
		t.Fatalf("expected old answer to be replaced, got %q", content)
	}
	if !strings.Contains(content, "## 2. 第2题 更新标题") {
		t.Fatalf("expected updated title to be kept, got %q", content)
	}
	if strings.Count(content, "## 2.") != 1 {
		t.Fatalf("expected only one entry for the same ordering number, got %q", content)
	}
}

func TestExportDocumentMarkdownRejectsTitleWithoutNumber(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	outputPath := filepath.Join(root, "document.md")
	service := NewService()

	err := service.ExportDocumentMarkdown("没有序号的标题", "答案", outputPath)
	if err == nil {
		t.Fatalf("expected export to fail when title has no number")
	}
}

func TestExportMarkdownDocumentPathUsesDocumentFileName(t *testing.T) {
	t.Parallel()

	got, err := ExportMarkdownDocumentPath(`E:\Project\exports`)
	if err != nil {
		t.Fatalf("ExportMarkdownDocumentPath returned error: %v", err)
	}

	want := filepath.Join(`E:\Project\exports`, "document.md")
	if got != want {
		t.Fatalf("expected %q, got %q", want, got)
	}
}

func TestExportDocumentsMarkdownMergesAndSortsEntries(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	outputPath := filepath.Join(root, "document.md")
	service := NewService()

	if err := service.ExportDocumentsMarkdown([]MarkdownExportDocument{
		{Title: "第3题 GMP 是什么？", Answer: "第三题答案"},
	}, outputPath); err != nil {
		t.Fatalf("first batch export returned error: %v", err)
	}

	if err := service.ExportDocumentsMarkdown([]MarkdownExportDocument{
		{Title: "第1题 Channel 是什么？", Answer: "第一题答案"},
		{Title: "第3题 GMP 是什么？", Answer: "第三题新答案"},
	}, outputPath); err != nil {
		t.Fatalf("second batch export returned error: %v", err)
	}

	raw, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("failed to read export file: %v", err)
	}

	content := string(raw)
	firstIndex := strings.Index(content, "## 1. 第1题 Channel 是什么？")
	secondIndex := strings.Index(content, "## 3. 第3题 GMP 是什么？")
	if firstIndex == -1 || secondIndex == -1 {
		t.Fatalf("expected merged entries to exist, got %q", content)
	}
	if firstIndex > secondIndex {
		t.Fatalf("expected merged entries to stay sorted, got %q", content)
	}
	if !strings.Contains(content, "第三题新答案") {
		t.Fatalf("expected latest answer to overwrite existing entry, got %q", content)
	}
	if strings.Contains(content, "第三题答案") {
		t.Fatalf("expected stale answer to be replaced, got %q", content)
	}
}

func TestExportDocumentsMarkdownFailsWhenAnyTitleHasNoNumber(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	outputPath := filepath.Join(root, "document.md")
	service := NewService()

	err := service.ExportDocumentsMarkdown([]MarkdownExportDocument{
		{Title: "第2题 原始标题", Answer: "答案"},
		{Title: "没有序号的标题", Answer: "答案"},
	}, outputPath)
	if err == nil {
		t.Fatalf("expected batch export to fail when a title has no number")
	}
}

func TestExportDocumentsMarkdownKeepsDistinctCompoundNumbers(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	outputPath := filepath.Join(root, "document.md")
	service := NewService()

	err := service.ExportDocumentsMarkdown([]MarkdownExportDocument{
		{Title: "4-25. interface 可以比较吗？", Answer: "答案 A"},
		{Title: "4-17. 哪些类型可以使用 len 和 cap？", Answer: "答案 B"},
		{Title: "2-9. 面试前如何准备？", Answer: "答案 C"},
	}, outputPath)
	if err != nil {
		t.Fatalf("batch export returned error: %v", err)
	}

	raw, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("failed to read export file: %v", err)
	}

	content := string(raw)
	if strings.Count(content, "## ") != 3 {
		t.Fatalf("expected 3 exported entries, got %q", content)
	}
	if !strings.Contains(content, "4-25. interface 可以比较吗？") {
		t.Fatalf("expected first compound number title to be kept, got %q", content)
	}
	if !strings.Contains(content, "4-17. 哪些类型可以使用 len 和 cap？") {
		t.Fatalf("expected second compound number title to be kept, got %q", content)
	}
	if !strings.Contains(content, "2-9. 面试前如何准备？") {
		t.Fatalf("expected third compound number title to be kept, got %q", content)
	}
}

func TestExportDocumentsMarkdownOmitsHiddenMetadataComments(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	outputPath := filepath.Join(root, "document.md")
	service := NewService()

	err := service.ExportDocumentsMarkdown([]MarkdownExportDocument{
		{Title: "4-25. interface 可以比较吗？", Answer: "答案 A"},
	}, outputPath)
	if err != nil {
		t.Fatalf("batch export returned error: %v", err)
	}

	raw, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("failed to read export file: %v", err)
	}

	content := string(raw)
	if strings.Contains(content, "TENQ_EXPORT_ENTRY") {
		t.Fatalf("expected export to omit hidden metadata comments, got %q", content)
	}
}

func TestGenerateInterviewAudioFromCacheReturnsGeneratorResult(t *testing.T) {
	previousFactory := newInterviewAudioGenerator
	t.Cleanup(func() {
		newInterviewAudioGenerator = previousFactory
	})

	newInterviewAudioGenerator = func(config audio.GeneratorConfig) interviewAudioGenerator {
		if !strings.HasSuffix(config.CachePath, filepath.Join(".cache", "tenq-interview", "index.json")) {
			t.Fatalf("unexpected cache path: %q", config.CachePath)
		}
		return stubInterviewAudioGenerator{
			result: audio.Result{
				OutputPath:       filepath.Join("E:\\Project\\Agent\\TenQ-Interview", ".cache", "tenq-interview", "audio", "session.wav"),
				TotalEntries:     3,
				GeneratedEntries: 2,
				SkippedEntries:   1,
				GeneratedAt:      "2026-04-20T16:00:00+08:00",
				Backend:          "onnx",
			},
		}
	}

	service, err := NewServiceWithCache(filepath.Join("E:\\Project\\Agent\\TenQ-Interview", ".cache", "tenq-interview", "index.json"))
	if err != nil {
		t.Fatalf("NewServiceWithCache returned error: %v", err)
	}

	result, err := service.GenerateInterviewAudioFromCache()
	if err != nil {
		t.Fatalf("GenerateInterviewAudioFromCache returned error: %v", err)
	}

	if result.OutputPath == "" || result.GeneratedEntries != 2 || result.Backend != "onnx" {
		t.Fatalf("unexpected audio generation result: %+v", result)
	}
}

func TestCompareDocumentTitlesSortsNumerically(t *testing.T) {
	tests := []struct {
		a    string
		b    string
		want bool
	}{
		{"1-1 第一个问题", "1-2 第二个问题", true},
		{"1-2 第二个问题", "1-10 第十个问题", true},
		{"1-10 第十个问题", "2-1 新问题", true},
		{"2-1 新问题", "11-3 MySQL 的 redo log", true},
		{"11-3 MySQL 的 redo log", "11-10 另一个问题", true},
		{"1-1", "1-2", true},
		{"1-2", "1-10", true},
		{"2-4", "10-1", true},
		{"相同标题", "相同标题", false},
	}

	for _, tt := range tests {
		got := compareDocumentTitles(tt.a, tt.b)
		if got != tt.want {
			t.Errorf("compareDocumentTitles(%q, %q) = %v, want %v", tt.a, tt.b, got, tt.want)
		}
	}
}

func TestParseTitleNumberExtractsDigits(t *testing.T) {
	tests := []struct {
		title string
		want  []int
	}{
		{"1-1 第一个问题", []int{1, 1}},
		{"11-3 MySQL 的 redo log", []int{11, 3}},
		{"2-7.遇到回答不上来", []int{2, 7}},
		{"第 100 题", []int{100}},
		{"没有数字", []int{}},
	}

	for _, tt := range tests {
		got := parseTitleNumber(tt.title)
		if len(got) != len(tt.want) {
			t.Errorf("parseTitleNumber(%q) = %v, want %v", tt.title, got, tt.want)
			continue
		}
		for i := range got {
			if got[i] != tt.want[i] {
				t.Errorf("parseTitleNumber(%q) = %v, want %v", tt.title, got, tt.want)
				break
			}
		}
	}
}
