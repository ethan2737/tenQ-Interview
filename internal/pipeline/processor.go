package pipeline

import (
	"os"

	"tenq-interview/internal/card"
	"tenq-interview/internal/importer"
	"tenq-interview/internal/parser"
	"tenq-interview/internal/segment"
)

type Result struct {
	Path     string
	Encoding string
	Title    string
	Body     string
	Card     card.Card
	Segments []segment.CandidateSegment
}

type Processor struct{}

func NewProcessor() *Processor {
	return &Processor{}
}

func (p *Processor) ProcessFile(path string) (Result, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return Result{}, err
	}

	normalized, encoding, err := importer.NormalizeMarkdownBytes(raw)
	if err != nil {
		return Result{}, err
	}

	parsed, err := parser.ParseMarkdown(path, normalized)
	if err != nil {
		return Result{}, err
	}

	segments := segment.SelectCandidateSegments(parsed.Title, parsed.Body, 3)
	generated, err := card.GenerateCard(parsed.Title, segments)
	if err != nil {
		return Result{}, err
	}

	return Result{
		Path:     path,
		Encoding: encoding,
		Title:    parsed.Title,
		Body:     parsed.Body,
		Card:     generated,
		Segments: segments,
	}, nil
}
