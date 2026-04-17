package parser

import "testing"

func TestParseMarkdownExtractsHeadingAndBody(t *testing.T) {
	t.Parallel()

	markdown := "# Go 的 GMP 模型是什么？\n\nG 表示 goroutine。\n\nM 表示线程。\n\nP 表示处理器上下文。"

	doc, err := ParseMarkdown("docs-go/gmp.md", markdown)
	if err != nil {
		t.Fatalf("ParseMarkdown returned error: %v", err)
	}

	if doc.Title != "Go 的 GMP 模型是什么？" {
		t.Fatalf("unexpected title: %q", doc.Title)
	}

	if doc.Body == "" {
		t.Fatalf("expected body to be populated")
	}

	if doc.Path != "docs-go/gmp.md" {
		t.Fatalf("unexpected path: %q", doc.Path)
	}
}

func TestParseMarkdownStripsUTF8BOMFromHeading(t *testing.T) {
	t.Parallel()

	markdown := "\uFEFF# GMP 是什么？\n\nGMP 是 Go 的调度模型。"

	doc, err := ParseMarkdown("docs-go/bom.md", markdown)
	if err != nil {
		t.Fatalf("ParseMarkdown returned error: %v", err)
	}

	if doc.Title != "GMP 是什么？" {
		t.Fatalf("unexpected title: %q", doc.Title)
	}

	if doc.Body != "GMP 是 Go 的调度模型。" {
		t.Fatalf("unexpected body: %q", doc.Body)
	}
}
