package workbench

import (
	"os"
	"path/filepath"
	"testing"

	"golang.org/x/text/encoding/simplifiedchinese"
)

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

	first, err := service.ProcessDocument(docPath, "gmp.md")
	if err != nil {
		t.Fatalf("ProcessDocument returned error: %v", err)
	}
	if first.FromCache {
		t.Fatalf("expected first run to be fresh")
	}

	second, err := service.ProcessDocument(docPath, "gmp.md")
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

	first, err := service.ProcessDocument(docPath, "broken.md")
	if err != nil {
		t.Fatalf("ProcessDocument returned error: %v", err)
	}
	if first.Status != StatusFailed {
		t.Fatalf("expected first run to fail, got %s", first.Status)
	}

	if err := os.WriteFile(docPath, []byte("# 修复后文档\n\n现在已经有正文，可以重新生成题卡。"), 0o600); err != nil {
		t.Fatalf("failed to rewrite fixture: %v", err)
	}

	second, err := service.ProcessDocument(docPath, "broken.md")
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
	if _, err := firstService.ProcessDocument(firstPath, "gmp.md"); err != nil {
		t.Fatalf("ProcessDocument firstPath returned error: %v", err)
	}
	if _, err := firstService.ProcessDocument(secondPath, "channel.md"); err != nil {
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
	if _, err := firstService.ProcessDocument(docPath, "gmp.md"); err != nil {
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
