package doi

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const crossrefWorksURL = "https://api.crossref.org/works/"

var (
	ErrEmptyDOI = errors.New("doi is empty")
	ErrNotFound = errors.New("doi not found")
)

type Service struct {
	client *http.Client
}

func NewService(client *http.Client) *Service {
	if client == nil {
		client = &http.Client{Timeout: 10 * time.Second}
	}
	return &Service{client: client}
}

func (s *Service) ResolveMetadata(ctx context.Context, rawDOI string) (*Metadata, error) {
	doi := strings.TrimSpace(rawDOI)
	if doi == "" {
		return nil, ErrEmptyDOI
	}

	endpoint := crossrefWorksURL + url.PathEscape(doi)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("create crossref request: %w", err)
	}
	req.Header.Set("Accept", "service/json")
	req.Header.Set("User-Agent", "paperstacks-backend/1.0")

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request crossref: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, ErrNotFound
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("crossref returned status %d", resp.StatusCode)
	}

	var body struct {
		Message struct {
			DOI       string `json:"DOI"`
			Title     []string
			Publisher string `json:"publisher"`
			Type      string `json:"type"`
			URL       string `json:"URL"`
			Author    []struct {
				Given  string `json:"given"`
				Family string `json:"family"`
				Name   string `json:"name"`
			} `json:"author"`
			PublishedPrint struct {
				DateParts [][]int `json:"date-parts"`
			} `json:"published-print"`
			PublishedOnline struct {
				DateParts [][]int `json:"date-parts"`
			} `json:"published-online"`
			Issued struct {
				DateParts [][]int `json:"date-parts"`
			} `json:"issued"`
		} `json:"message"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return nil, fmt.Errorf("decode crossref response: %w", err)
	}

	return &Metadata{
		DOI:       body.Message.DOI,
		Title:     firstOrEmpty(body.Message.Title),
		Publisher: body.Message.Publisher,
		Type:      body.Message.Type,
		Authors:   authorsFromMessage(body.Message.Author),
		Published: publishedDate(body.Message.PublishedPrint.DateParts, body.Message.PublishedOnline.DateParts, body.Message.Issued.DateParts),
		URL:       body.Message.URL,
	}, nil
}

func firstOrEmpty(values []string) string {
	if len(values) == 0 {
		return ""
	}
	return values[0]
}

func authorsFromMessage(authors []struct {
	Given  string `json:"given"`
	Family string `json:"family"`
	Name   string `json:"name"`
},
) []string {
	if len(authors) == 0 {
		return nil
	}

	result := make([]string, 0, len(authors))
	for _, author := range authors {
		switch {
		case strings.TrimSpace(author.Name) != "":
			result = append(result, author.Name)
		default:
			full := strings.TrimSpace(strings.TrimSpace(author.Given) + " " + strings.TrimSpace(author.Family))
			if full != "" {
				result = append(result, full)
			}
		}
	}

	return result
}

func publishedDate(dateParts ...[][]int) string {
	for _, parts := range dateParts {
		if len(parts) == 0 || len(parts[0]) == 0 {
			continue
		}

		values := parts[0]
		if len(values) >= 3 {
			return fmt.Sprintf("%04d-%02d-%02d", values[0], values[1], values[2])
		}
		if len(values) == 2 {
			return fmt.Sprintf("%04d-%02d", values[0], values[1])
		}
		return fmt.Sprintf("%04d", values[0])
	}
	return ""
}
