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

func BuildCacheKey(path string, fingerprint string, versions RuleVersions) string {
	payload := strings.Join([]string{
		path,
		fingerprint,
		versions.ParserVersion,
		versions.SegmentVersion,
		versions.GeneratorVersion,
	}, "::")

	sum := sha256.Sum256([]byte(payload))
	return hex.EncodeToString(sum[:])
}
