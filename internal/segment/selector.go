package segment

import (
	"regexp"
	"sort"
	"strings"
	"unicode/utf8"
)

type CandidateSegment struct {
	Text  string
	Score int
}

var tokenPattern = regexp.MustCompile(`[\p{Han}\p{L}\p{N}]+`)

func SelectCandidateSegments(question string, body string, limit int) []CandidateSegment {
	if limit <= 0 {
		limit = 1
	}

	paragraphs := splitParagraphs(body)
	if len(paragraphs) == 0 {
		return nil
	}

	tokens := questionTokens(question)
	candidates := make([]CandidateSegment, 0, len(paragraphs))
	for index, paragraph := range paragraphs {
		score := scoreParagraph(tokens, paragraph)
		if score == 0 {
			score = max(1, len(paragraphs)-index)
		}
		candidates = append(candidates, CandidateSegment{
			Text:  paragraph,
			Score: score,
		})
	}

	sort.SliceStable(candidates, func(i, j int) bool {
		return candidates[i].Score > candidates[j].Score
	})

	if len(candidates) > limit {
		candidates = candidates[:limit]
	}

	return candidates
}

func splitParagraphs(body string) []string {
	chunks := strings.Split(body, "\n\n")
	paragraphs := make([]string, 0, len(chunks))
	for _, chunk := range chunks {
		trimmed := strings.TrimSpace(strings.ReplaceAll(chunk, "\n", " "))
		if trimmed != "" {
			paragraphs = append(paragraphs, trimmed)
		}
	}
	return paragraphs
}

func questionTokens(question string) []string {
	rawTokens := tokenPattern.FindAllString(question, -1)
	seen := map[string]struct{}{}
	stopWords := map[string]struct{}{
		"什么": {}, "怎么": {}, "如何": {}, "为何": {}, "为什么": {}, "多少": {},
		"区别": {}, "作用": {}, "实现": {}, "底层": {}, "原理": {}, "可以": {},
		"语言": {}, "中的": {}, "的是": {}, "的": {}, "是": {},
	}

	var tokens []string
	for _, token := range rawTokens {
		token = strings.TrimSpace(token)
		if token == "" {
			continue
		}
		if _, blocked := stopWords[token]; blocked {
			continue
		}
		if _, exists := seen[token]; !exists {
			seen[token] = struct{}{}
			tokens = append(tokens, token)
		}
		if utf8.RuneCountInString(token) >= 4 {
			runes := []rune(token)
			for i := 0; i < len(runes)-1; i++ {
				chunk := string(runes[i : i+2])
				if _, blocked := stopWords[chunk]; blocked {
					continue
				}
				if _, exists := seen[chunk]; !exists {
					seen[chunk] = struct{}{}
					tokens = append(tokens, chunk)
				}
			}
		}
	}
	return tokens
}

func scoreParagraph(tokens []string, paragraph string) int {
	score := 0
	for _, token := range tokens {
		if strings.Contains(paragraph, token) {
			score += 3
		}
	}
	return score
}

func max(a int, b int) int {
	if a > b {
		return a
	}
	return b
}
