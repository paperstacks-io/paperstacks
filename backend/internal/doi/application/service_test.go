package application

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (fn roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return fn(req)
}

func TestDOIServiceResolveMetadata(t *testing.T) {
	t.Parallel()

	client := &http.Client{
		Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			if got := req.URL.String(); got != "https://api.crossref.org/works/10.1000%2F182" {
				t.Fatalf("request URL = %q", got)
			}

			body := `{
				"message": {
					"DOI": "10.1000/182",
					"title": ["Example title"],
					"subject": ["quality assurance", "gui testing"],
					"publisher": "Example publisher",
					"type": "journal-article",
					"URL": "https://doi.org/10.1000/182",
					"author": [
						{"given": "Ada", "family": "Lovelace"},
						{"name": "Anonymous"}
					],
					"issued": {"date-parts": [[2024, 5, 1]]}
				}
			}`

			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(strings.NewReader(body)),
				Header:     make(http.Header),
			}, nil
		}),
	}

	service := NewDOIService(client)

	got, err := service.ResolveMetadata(context.Background(), " 10.1000/182 ")
	if err != nil {
		t.Fatalf("ResolveMetadata() error = %v", err)
	}

	if got.DOI != "10.1000/182" {
		t.Fatalf("DOI = %q, want %q", got.DOI, "10.1000/182")
	}

	if got.Title != "Example title" {
		t.Fatalf("Title = %q, want %q", got.Title, "Example title")
	}

	if len(got.Keywords) != 2 || got.Keywords[0] != "quality assurance" || got.Keywords[1] != "gui testing" {
		t.Fatalf("Keywords = %#v", got.Keywords)
	}

	if got.Published != "2024-05-01" {
		t.Fatalf("Published = %q, want %q", got.Published, "2024-05-01")
	}

	if len(got.Authors) != 2 || got.Authors[0] != "Ada Lovelace" || got.Authors[1] != "Anonymous" {
		t.Fatalf("Authors = %#v", got.Authors)
	}
}

func TestDOIServiceResolveMetadataReturnsNotFound(t *testing.T) {
	t.Parallel()

	client := &http.Client{
		Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: http.StatusNotFound,
				Body:       io.NopCloser(strings.NewReader("")),
				Header:     make(http.Header),
			}, nil
		}),
	}

	service := NewDOIService(client)

	_, err := service.ResolveMetadata(context.Background(), "10.1000/404")
	if err != ErrNotFound {
		t.Fatalf("ResolveMetadata() error = %v, want %v", err, ErrNotFound)
	}
}

func TestDOIServiceResolveMetadataRejectsEmptyDOI(t *testing.T) {
	t.Parallel()

	service := NewDOIService(nil)

	_, err := service.ResolveMetadata(context.Background(), " ")
	if err != ErrEmptyDOI {
		t.Fatalf("ResolveMetadata() error = %v, want %v", err, ErrEmptyDOI)
	}
}
