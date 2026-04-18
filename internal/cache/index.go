package cache

import (
	"crypto/sha256"
	"encoding/hex"
	"strings"
)

type RuleVersions struct {
	ParserVersion    string
	SegmentVersion   string
	GeneratorVersion string
}

func BuildCacheKey(path string, fingerprint string, versions RuleVersions, extras ...string) string {
	parts := []string{
		path,
		fingerprint,
		versions.ParserVersion,
		versions.SegmentVersion,
		versions.GeneratorVersion,
	}
	parts = append(parts, extras...)

	payload := strings.Join(parts, "::")

	sum := sha256.Sum256([]byte(payload))
	return hex.EncodeToString(sum[:])
}
