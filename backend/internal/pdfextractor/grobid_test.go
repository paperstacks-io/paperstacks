package pdfextractor

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (fn roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return fn(req)
}

func TestGrobidClientCheck(t *testing.T) {
	t.Parallel()

	t.Run("healthy service returns nil", func(t *testing.T) {
		t.Parallel()

		httpClient := &http.Client{
			Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
				if req.Method != http.MethodGet {
					t.Fatalf("method = %q, want %q", req.Method, http.MethodGet)
				}
				if got := req.URL.String(); got != "http://grobid:8070/api/isalive" {
					t.Fatalf("URL = %q, want %q", got, "http://grobid:8070/api/isalive")
				}
				return &http.Response{
					StatusCode: http.StatusOK,
					Body:       io.NopCloser(strings.NewReader("true")),
					Header:     make(http.Header),
				}, nil
			}),
		}

		client := NewGrobidClient("http://grobid:8070/", httpClient)
		if err := client.Check(context.Background()); err != nil {
			t.Fatalf("Check() error = %v", err)
		}
	})

	t.Run("unhealthy status returns error", func(t *testing.T) {
		t.Parallel()

		httpClient := &http.Client{
			Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
				return &http.Response{
					StatusCode: http.StatusServiceUnavailable,
					Body:       io.NopCloser(strings.NewReader("")),
					Header:     make(http.Header),
				}, nil
			}),
		}

		client := NewGrobidClient("http://grobid:8070", httpClient)
		if err := client.Check(context.Background()); err == nil {
			t.Fatal("Check() expected error, got nil")
		}
	})
}

func TestGrobidClientExtractFulltext(t *testing.T) {
	t.Parallel()

	wantPDF := []byte("%PDF-1.7 sample pdf content")
	wantXML := "<TEI><teiHeader><fileDesc><titleStmt><title>Sample Paper</title></titleStmt></fileDesc></teiHeader></TEI>"

	httpClient := &http.Client{
		Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			if req.Method != http.MethodPost {
				t.Fatalf("method = %q, want %q", req.Method, http.MethodPost)
			}
			if got := req.URL.String(); got != "http://grobid:8070/api/processFulltextDocument" {
				t.Fatalf("URL = %q, want %q", got, "http://grobid:8070/api/processFulltextDocument")
			}
			if got := req.Header.Get("Accept"); got != "application/xml" {
				t.Fatalf("Accept = %q, want %q", got, "application/xml")
			}

			file, header, err := req.FormFile(formFieldInput)
			if err != nil {
				t.Fatalf("FormFile(%q) error = %v", formFieldInput, err)
			}
			defer file.Close()

			if header.Filename != defaultFileName {
				t.Fatalf("filename = %q, want %q", header.Filename, defaultFileName)
			}

			gotPDF, err := io.ReadAll(file)
			if err != nil {
				t.Fatalf("read uploaded file: %v", err)
			}
			if !bytes.Equal(gotPDF, wantPDF) {
				t.Fatalf("uploaded PDF = %q, want %q", gotPDF, wantPDF)
			}

			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(strings.NewReader(wantXML)),
				Header:     make(http.Header),
			}, nil
		}),
	}

	client := NewGrobidClient("http://grobid:8070", httpClient)
	gotXML, err := client.ExtractFulltext(context.Background(), bytes.NewReader(wantPDF))
	if err != nil {
		t.Fatalf("ExtractFulltext() error = %v", err)
	}

	if string(gotXML) != wantXML {
		t.Fatalf("ExtractFulltext() = %q, want %q", string(gotXML), wantXML)
	}
}

func TestGrobidClientExtractFulltextErrors(t *testing.T) {
	t.Parallel()

	t.Run("nil reader", func(t *testing.T) {
		t.Parallel()
		client := NewGrobidClient("http://grobid:8070", nil)
		_, err := client.ExtractFulltext(context.Background(), nil)
		if !errors.Is(err, ErrNilPDF) {
			t.Fatalf("error = %v, want %v", err, ErrNilPDF)
		}
	})

	t.Run("invalid pdf header", func(t *testing.T) {
		t.Parallel()
		client := NewGrobidClient("http://grobid:8070", nil)
		_, err := client.ExtractFulltext(context.Background(), strings.NewReader("not a pdf"))
		if !errors.Is(err, ErrInvalidPDF) {
			t.Fatalf("error = %v, want %v", err, ErrInvalidPDF)
		}
	})

	t.Run("retries on 503 and succeeds", func(t *testing.T) {
		t.Parallel()
		attempts := 0
		httpClient := &http.Client{
			Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
				attempts++
				_, _ = io.Copy(io.Discard, req.Body)
				if attempts < 3 {
					return &http.Response{
						StatusCode: http.StatusServiceUnavailable,
						Body:       io.NopCloser(strings.NewReader("busy")),
						Header:     make(http.Header),
					}, nil
				}
				return &http.Response{
					StatusCode: http.StatusOK,
					Body:       io.NopCloser(strings.NewReader("<TEI>recovered</TEI>")),
					Header:     make(http.Header),
				}, nil
			}),
		}

		client := NewGrobidClient("http://grobid:8070", httpClient)
		client.retryDelay = time.Millisecond

		got, err := client.ExtractFulltext(context.Background(), strings.NewReader("%PDF-1.7 test"))
		if err != nil {
			t.Fatalf("ExtractFulltext() error = %v", err)
		}
		if string(got) != "<TEI>recovered</TEI>" {
			t.Fatalf("got = %q, want %q", string(got), "<TEI>recovered</TEI>")
		}
		if attempts != 3 {
			t.Fatalf("attempts = %d, want 3", attempts)
		}
	})

	tests := []struct {
		name       string
		statusCode int
		body       string
		wantErr    error
	}{
		{
			name:       "204 no content",
			statusCode: http.StatusNoContent,
			body:       "",
			wantErr:    ErrEmptyExtraction,
		},
		{
			name:       "200 empty body",
			statusCode: http.StatusOK,
			body:       "   \n\t  ",
			wantErr:    ErrEmptyExtraction,
		},
		{
			name:       "503 service busy",
			statusCode: http.StatusServiceUnavailable,
			body:       "busy",
			wantErr:    ErrServiceBusy,
		},
		{
			name:       "500 internal server error",
			statusCode: http.StatusInternalServerError,
			body:       "segmentation fault",
			wantErr:    nil,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			httpClient := &http.Client{
				Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
					_, _ = io.Copy(io.Discard, req.Body)
					return &http.Response{
						StatusCode: tc.statusCode,
						Body:       io.NopCloser(strings.NewReader(tc.body)),
						Header:     make(http.Header),
					}, nil
				}),
			}

			client := NewGrobidClient("http://grobid:8070", httpClient)
			client.retryDelay = time.Millisecond
			_, err := client.ExtractFulltext(context.Background(), strings.NewReader("%PDF-1.7"))
			if err == nil {
				t.Fatal("expected error, got nil")
			}
			if tc.wantErr != nil && !errors.Is(err, tc.wantErr) {
				t.Fatalf("error = %v, want %v", err, tc.wantErr)
			}
		})
	}
}
