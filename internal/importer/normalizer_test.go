package importer

import (
	"testing"

	"golang.org/x/text/encoding/simplifiedchinese"
)

func TestNormalizeMarkdownBytesPreservesUTF8Input(t *testing.T) {
	t.Parallel()

	input := []byte("# Go 的 GMP 模型是什么？\n\nG 表示 goroutine，M 表示线程，P 表示处理器上下文。")

	normalized, encoding, err := NormalizeMarkdownBytes(input)
	if err != nil {
		t.Fatalf("NormalizeMarkdownBytes returned error: %v", err)
	}

	if encoding != EncodingUTF8 {
		t.Fatalf("expected encoding %q, got %q", EncodingUTF8, encoding)
	}

	if normalized != string(input) {
		t.Fatalf("expected normalized text to match original utf-8 input")
	}
}

func TestNormalizeMarkdownBytesFallsBackToGB18030(t *testing.T) {
	t.Parallel()

	raw := "# channel 的底层实现是什么？\n\nchannel 是 Go 提供的线程安全通信机制。"
	encoded, err := simplifiedchinese.GB18030.NewEncoder().Bytes([]byte(raw))
	if err != nil {
		t.Fatalf("failed to create gb18030 fixture: %v", err)
	}

	normalized, encoding, err := NormalizeMarkdownBytes(encoded)
	if err != nil {
		t.Fatalf("NormalizeMarkdownBytes returned error: %v", err)
	}

	if encoding != EncodingGB18030 {
		t.Fatalf("expected encoding %q, got %q", EncodingGB18030, encoding)
	}

	if normalized != raw {
		t.Fatalf("expected gb18030 content to be decoded back to original text")
	}
}

func TestDetectLikelyGarbledText(t *testing.T) {
	t.Parallel()

	if suspect, _ := DetectLikelyGarbledText("Go 的 GMP 模型是什么？"); suspect {
		t.Fatalf("expected normal chinese text to be accepted")
	}

	suspect, reason := DetectLikelyGarbledText("Go 鐨勭嚎绋嬫ā鍨嬫槸浠€涔堬紵")
	if !suspect {
		t.Fatalf("expected mojibake-like text to be flagged")
	}
	if reason == "" {
		t.Fatalf("expected suspicion reason to be populated")
	}
}
