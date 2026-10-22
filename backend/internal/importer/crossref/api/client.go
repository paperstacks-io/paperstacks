// Package api fetches changed works from the Crossref REST API
// (https://api.crossref.org/swagger-ui/index.html) one page at a time.
// Items are returned as raw JSON; they share the public data export's
// schema, so callers decode them with the parser package.
package api

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

// DefaultBaseURL is the public Crossref REST API.
const DefaultBaseURL = "https://api.crossref.org"

// pageSize is the maximum number of rows Crossref returns per request.
const pageSize = 1000

// FirstCursor starts a new deep-paging sequence.
const FirstCursor = "*"

// Filter selects which Crossref date a day window is matched against.
type Filter string

const (
	// FilterUpdateDate matches works whose metadata was (re)deposited by
	// the publisher on that day.
	FilterUpdateDate Filter = "update-date"
	// FilterIndexDate matches works Crossref re-indexed on that day,
	// which also includes changes such as citation counts.
	FilterIndexDate Filter = "index-date"
)

// IsValid reports whether f is a supported filter.
func (f Filter) IsValid() bool {
	return f == FilterUpdateDate || f == FilterIndexDate
}

// Page is one response of a cursor-paged /works query.
type Page struct {
	Items        []json.RawMessage
	NextCursor   string
	TotalResults int
}

// Client queries the Crossref /works endpoint. Requests carrying a mailto
// address are routed to Crossref's polite pool; without one, to the
// public pool.
type Client struct {
	baseURL    string
	mailto     string
	httpClient *http.Client
	maxRetries int
	backoff    time.Duration
}

// NewClient returns a Client for baseURL. mailto may be empty.
func NewClient(baseURL, mailto string, httpClient *http.Client) *Client {
	if httpClient == nil {
		// Full-record pages of 1000 works can be tens of megabytes.
		httpClient = &http.Client{Timeout: 2 * time.Minute}
	}
	return &Client{
		baseURL:    baseURL,
		mailto:     mailto,
		httpClient: httpClient,
		maxRetries: 5,
		backoff:    2 * time.Second,
	}
}

// FetchPage returns the page at cursor of all works matching filter on
// day (UTC, inclusive). Pass FirstCursor to start; continue with
// Page.NextCursor until a page has no items. Rate limiting (429), server
// errors and transport errors are retried with exponential backoff.
func (c *Client) FetchPage(ctx context.Context, filter Filter, day time.Time, cursor string) (Page, error) {
	endpoint := c.pageURL(filter, day, cursor)
	wait := c.backoff
	for attempt := 0; ; attempt++ {
		page, err := c.get(ctx, endpoint)
		if err == nil {
			return page, nil
		}

		var retry *retryableError
		if !errors.As(err, &retry) || attempt == c.maxRetries || ctx.Err() != nil {
			return Page{}, err
		}
		if retry.after > 0 {
			wait = retry.after
		}
		if err := sleep(ctx, wait); err != nil {
			return Page{}, err
		}
		wait *= 2
	}
}

func (c *Client) pageURL(filter Filter, day time.Time, cursor string) string {
	d := day.UTC().Format(time.DateOnly)
	q := url.Values{}
	q.Set("filter", fmt.Sprintf("from-%s:%s,until-%s:%s", filter, d, filter, d))
	q.Set("rows", strconv.Itoa(pageSize))
	q.Set("cursor", cursor)
	if c.mailto != "" {
		q.Set("mailto", c.mailto)
	}
	return c.baseURL + "/works?" + q.Encode()
}

func (c *Client) get(ctx context.Context, endpoint string) (Page, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return Page{}, fmt.Errorf("create crossref request: %w", err)
	}
	req.Header.Set("Accept", "application/json")
	ua := "paperstacks-crossref-sync/1.0"
	if c.mailto != "" {
		ua += " (mailto:" + c.mailto + ")"
	}
	req.Header.Set("User-Agent", ua)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return Page{}, &retryableError{err: fmt.Errorf("request crossref: %w", err)}
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		snippet, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		err := fmt.Errorf("crossref returned status %d: %s", resp.StatusCode, snippet)
		if resp.StatusCode == http.StatusTooManyRequests || resp.StatusCode >= 500 {
			return Page{}, &retryableError{err: err, after: retryAfter(resp.Header)}
		}
		return Page{}, err
	}

	var body struct {
		Message struct {
			TotalResults int               `json:"total-results"`
			NextCursor   string            `json:"next-cursor"`
			Items        []json.RawMessage `json:"items"`
		} `json:"message"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		// Usually a connection dropped mid-body; worth another attempt.
		return Page{}, &retryableError{err: fmt.Errorf("decode crossref response: %w", err)}
	}
	return Page{
		Items:        body.Message.Items,
		NextCursor:   body.Message.NextCursor,
		TotalResults: body.Message.TotalResults,
	}, nil
}

type retryableError struct {
	err   error
	after time.Duration // from Retry-After; 0 means use the backoff.
}

func (e *retryableError) Error() string { return e.err.Error() }
func (e *retryableError) Unwrap() error { return e.err }

func retryAfter(h http.Header) time.Duration {
	secs, err := strconv.Atoi(h.Get("Retry-After"))
	if err != nil || secs <= 0 {
		return 0
	}
	return time.Duration(secs) * time.Second
}

func sleep(ctx context.Context, d time.Duration) error {
	t := time.NewTimer(d)
	defer t.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-t.C:
		return nil
	}
}
