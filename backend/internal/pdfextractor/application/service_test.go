package application

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/paperstacks.io/paperstacks/internal/pdfextractor/domain"
)

func TestParseMetadataHeuristics(t *testing.T) {
	tests := []struct {
		name             string
		rawText          string
		expectedDOI      string
		expectedAbstract string
		expectedKeywords []string
	}{
		{
			name: "academic paper with DOI, abstract and keywords",
			rawText: `ACM Transactions on Software Engineering
DOI: 10.1145/1234567.1234568

Abstract: This paper explores automated software testing using LLM agents.
1. Introduction
Modern software engineering...

Keywords: Software Engineering, LLM, Agents, Testing
`,
			expectedDOI:      "10.1145/1234567.1234568",
			expectedAbstract: "This paper explores automated software testing using LLM agents.",
			expectedKeywords: []string{"Software Engineering", "LLM", "Agents", "Testing"},
		},
		{
			name: "scanned paper or empty text",
			rawText: `Scan page 1
Some random noisy text`,
			expectedDOI:      "",
			expectedAbstract: "",
			expectedKeywords: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res := parseMetadataHeuristics(tt.rawText)
			res.DetermineMissingFields()

			if res.DOI != tt.expectedDOI {
				t.Errorf("expected DOI %q, got %q", tt.expectedDOI, res.DOI)
			}
			if res.Abstract != tt.expectedAbstract {
				t.Errorf("expected Abstract %q, got %q", tt.expectedAbstract, res.Abstract)
			}
			if len(res.Keywords) != len(tt.expectedKeywords) {
				t.Errorf("expected %d keywords, got %d", len(tt.expectedKeywords), len(res.Keywords))
			} else {
				for i := range res.Keywords {
					if res.Keywords[i] != tt.expectedKeywords[i] {
						t.Errorf("keyword[%d] expected %q, got %q", i, tt.expectedKeywords[i], res.Keywords[i])
					}
				}
			}
		})
	}
}

func TestDetermineMissingFields(t *testing.T) {
	meta := domain.ExtractedMetadata{
		Title: "Paper Title",
		DOI:   "10.1234/5678",
	}
	meta.DetermineMissingFields()

	expectedMissing := map[string]bool{
		"authors":          true,
		"abstract":         true,
		"publication_date": true,
		"keywords":         true,
	}

	for _, field := range meta.MissingFields {
		delete(expectedMissing, field)
	}

	if len(expectedMissing) > 0 {
		var missingNames []string
		for k := range expectedMissing {
			missingNames = append(missingNames, k)
		}
		t.Errorf("expected fields were not marked as missing: %s", strings.Join(missingNames, ", "))
	}
}

func TestService_EmptyReader(t *testing.T) {
	s := NewService()
	_, err := s.ExtractMetadata(context.Background(), strings.NewReader(""), domain.Options{})
	if err == nil || !errors.Is(err, domain.ErrEmptyPDF) {
		t.Fatalf("expected ErrEmptyPDF, got: %v", err)
	}
}

