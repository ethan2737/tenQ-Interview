package cache

import "testing"

func TestBuildCacheKeyIncludesRuleVersions(t *testing.T) {
	t.Parallel()

	versions := RuleVersions{
		ParserVersion:    "parser-v1",
		SegmentVersion:   "segment-v1",
		GeneratorVersion: "generator-v1",
	}

	keyA := BuildCacheKey("docs-go/gmp.md", "fingerprint-a", versions)
	keyB := BuildCacheKey("docs-go/gmp.md", "fingerprint-a", RuleVersions{
		ParserVersion:    "parser-v2",
		SegmentVersion:   "segment-v1",
		GeneratorVersion: "generator-v1",
	})

	if keyA == keyB {
		t.Fatalf("expected cache key to change when parser version changes")
	}
}

func TestBuildCacheKeyIncludesProviderAndPromptVersion(t *testing.T) {
	t.Parallel()

	versions := RuleVersions{
		ParserVersion:    "parser-v1",
		SegmentVersion:   "segment-v1",
		GeneratorVersion: "generator-v1",
	}

	keyA := BuildCacheKey("docs-go/gmp.md", "fingerprint-a", versions, "deepseek", "deepseek-chat", "prompt-v1")
	keyB := BuildCacheKey("docs-go/gmp.md", "fingerprint-a", versions, "modelscope", "qwen-plus", "prompt-v1")

	if keyA == keyB {
		t.Fatalf("expected cache key to change when provider changes")
	}
}
