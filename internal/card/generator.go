package card

import (
	"errors"
	"strings"

	"tenq-interview/internal/segment"
)

type Card struct {
	Question string
	Answer   string
	Sources  []string
}

const maxAnswerRunes = 250

func GenerateCard(question string, segments []segment.CandidateSegment) (Card, error) {
	if strings.TrimSpace(question) == "" {
		return Card{}, errors.New("question is required")
	}
	if len(segments) == 0 {
		return Card{}, errors.New("at least one source segment is required")
	}

	sources := make([]string, 0, len(segments))
	answerParts := make([]string, 0, len(segments))
	currentLen := 0

	for _, item := range segments {
		text := compactWhitespace(item.Text)
		if text == "" {
			continue
		}
		sources = append(sources, text)
		nextLen := currentLen + len([]rune(text))
		if len(answerParts) == 0 || nextLen <= maxAnswerRunes {
			answerParts = append(answerParts, text)
			currentLen = len([]rune(strings.Join(answerParts, "")))
		}
	}

	if len(answerParts) == 0 {
		answerParts = append(answerParts, compactWhitespace(segments[0].Text))
	}

	answer := compactWhitespace(strings.Join(answerParts, ""))
	answer = trimToRunes(answer, maxAnswerRunes)
	if answer == "" {
		return Card{}, errors.New("generated answer is empty")
	}

	return Card{
		Question: question,
		Answer:   answer,
		Sources:  sources,
	}, nil
}

func compactWhitespace(input string) string {
	return strings.Join(strings.Fields(input), " ")
}

func trimToRunes(input string, limit int) string {
	runes := []rune(input)
	if len(runes) <= limit {
		return input
	}
	return strings.TrimSpace(string(runes[:limit]))
}
