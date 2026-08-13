//nolint:errcheck // to ignore not checking err when defer resp.Body.Close()
package integration

import (
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"testing"

	paperHttp "github.com/paperstacks.io/paperstacks/internal/paper/http"
)

func TestIntegrationSearchPapersQueryParams(t *testing.T) {
	setupIntegrationTest(t)

	type testCase struct {
		query             string
		sortBy            string
		page              int
		pageSize          int
		HTTPStatusCode    int
		expectedLen       int
		expectedFirstUUID string
	}

	tests := []testCase{
		// sort
		{query: "", sortBy: "-title", page: 0, pageSize: 100, HTTPStatusCode: http.StatusOK, expectedLen: 100, expectedFirstUUID: "a9d7c335"},
		{query: "", sortBy: "+title", page: 0, pageSize: 100, HTTPStatusCode: http.StatusOK, expectedLen: 100, expectedFirstUUID: "df34d6d8"},
		{query: "", sortBy: "-year", page: 0, pageSize: 100, HTTPStatusCode: http.StatusOK, expectedLen: 100, expectedFirstUUID: "bc884ec1"},
		{query: "", sortBy: "+year", page: 0, pageSize: 100, HTTPStatusCode: http.StatusOK, expectedLen: 100, expectedFirstUUID: "5966a651"},
		// title
		{query: "we tried", sortBy: "", page: 1, pageSize: 0, HTTPStatusCode: http.StatusOK, expectedLen: 1, expectedFirstUUID: "3df8adca"},
		{query: "regression testing", sortBy: "", page: 1, pageSize: 0, HTTPStatusCode: http.StatusOK, expectedLen: 1, expectedFirstUUID: "55c2240f"},
		// keyword
		{query: "diversity sample", sortBy: "", page: 1, pageSize: 0, HTTPStatusCode: http.StatusOK, expectedLen: 1, expectedFirstUUID: "1fa1a590"},
		{query: "qualitative", sortBy: "", page: 1, pageSize: 100, HTTPStatusCode: http.StatusOK, expectedLen: 15, expectedFirstUUID: "734c9f00"},
		// pagignation
		{query: "gui", sortBy: "", page: 1, pageSize: 0, HTTPStatusCode: http.StatusOK, expectedLen: 10, expectedFirstUUID: "fee26da6"},
		{query: "gui", sortBy: "", page: 1, pageSize: 2, HTTPStatusCode: http.StatusOK, expectedLen: 2, expectedFirstUUID: "fee26da6"},
		{query: "gui", sortBy: "", page: 2, pageSize: 2, HTTPStatusCode: http.StatusOK, expectedLen: 2, expectedFirstUUID: "ce127349"},
		{query: "gui", sortBy: "", page: 1, pageSize: 3, HTTPStatusCode: http.StatusOK, expectedLen: 3, expectedFirstUUID: "fee26da6"},
		{query: "", sortBy: "", page: 5, pageSize: 100, HTTPStatusCode: http.StatusOK, expectedLen: 0, expectedFirstUUID: ""},
		// bad request
		{query: "gui", sortBy: "bad-sort", page: 1, pageSize: 0, HTTPStatusCode: http.StatusBadRequest, expectedLen: 1, expectedFirstUUID: ""},
	}

	for _, tc := range tests {
		t.Run(tc.query+"_"+tc.sortBy, func(t *testing.T) {
			u, err := url.Parse(testAPIPath + "/api/papers")
			if err != nil {
				t.Fatalf("failed to parse url: %v", err)
			}

			q := u.Query()
			if tc.query != "" {
				q.Set("q", tc.query)
			}
			if tc.sortBy != "" {
				q.Set("sortBy", tc.sortBy)
			}
			if tc.page > 0 {
				q.Set("page", strconv.Itoa(tc.page))
			}
			if tc.pageSize > 0 {
				q.Set("pageSize", strconv.Itoa(tc.pageSize))
			}
			u.RawQuery = q.Encode()

			resp := doGetRequest(t, u.String())
			defer resp.Body.Close()

			assertStatusCode(t, resp, tc.HTTPStatusCode)

			if tc.HTTPStatusCode != http.StatusOK {
				return
			}

			var papers []paperHttp.PaperResponse
			decodeJSON(t, resp, &papers)

			if tc.expectedLen >= 0 && len(papers) != tc.expectedLen {
				t.Fatalf("expected %d papers, got %d", tc.expectedLen, len(papers))
			}

			if tc.expectedFirstUUID == "" {
				return
			}

			if len(papers) == 0 {
				t.Fatalf("expected first UUID prefix %q but got empty result", tc.expectedFirstUUID)
			}

			if !strings.HasPrefix(papers[0].UUID, tc.expectedFirstUUID) {
				t.Fatalf("expected first UUID to start with %q, got %q", tc.expectedFirstUUID, papers[0].UUID)
			}
		})
	}
}

func TestIntegrationSearchPapersPaginationHeaders(t *testing.T) {
	setupIntegrationTest(t)

	type testCase struct {
		name             string
		query            string
		sortBy           string
		page             int
		pageSize         int
		expectedPage     string
		expectedPageSize string
		expectedTotal    string
		expectedHasNext  string
	}

	tests := []testCase{
		{
			name:             "first page has next",
			query:            "gui",
			page:             1,
			pageSize:         50,
			expectedPage:     "1",
			expectedPageSize: "50",
			expectedTotal:    "62",
			expectedHasNext:  "true",
		},
		{
			name:             "second page has no next",
			query:            "gui",
			page:             2,
			pageSize:         50,
			expectedPage:     "2",
			expectedPageSize: "50",
			expectedTotal:    "62",
			expectedHasNext:  "false",
		},
		{
			name:             "defaults from service",
			query:            "gui",
			expectedPage:     "1",
			expectedPageSize: "10",
			expectedTotal:    "62",
			expectedHasNext:  "true",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			u, err := url.Parse(testAPIPath + "/api/papers")
			if err != nil {
				t.Fatalf("failed to parse url: %v", err)
			}

			q := u.Query()
			if tc.query != "" {
				q.Set("q", tc.query)
			}
			if tc.sortBy != "" {
				q.Set("sortBy", tc.sortBy)
			}
			if tc.page > 0 {
				q.Set("page", strconv.Itoa(tc.page))
			}
			if tc.pageSize > 0 {
				q.Set("pageSize", strconv.Itoa(tc.pageSize))
			}
			u.RawQuery = q.Encode()

			resp := doGetRequest(t, u.String())
			defer resp.Body.Close()

			assertStatusCode(t, resp, http.StatusOK)

			if got := resp.Header.Get("X-Page"); got != tc.expectedPage {
				t.Fatalf("expected X-Page %q, got %q", tc.expectedPage, got)
			}
			if got := resp.Header.Get("X-Page-Size"); got != tc.expectedPageSize {
				t.Fatalf("expected X-Page-Size %q, got %q", tc.expectedPageSize, got)
			}
			if got := resp.Header.Get("X-Total-Count"); got != tc.expectedTotal {
				t.Fatalf("expected X-Total-Count %q, got %q", tc.expectedTotal, got)
			}
			if got := resp.Header.Get("X-Has-Next"); got != tc.expectedHasNext {
				t.Fatalf("expected X-Has-Next %q, got %q", tc.expectedHasNext, got)
			}

			exposed := resp.Header.Get("Access-Control-Expose-Headers")
			for _, key := range []string{"X-Page", "X-Page-Size", "X-Total-Count", "X-Has-Next"} {
				if !strings.Contains(exposed, key) {
					t.Fatalf("expected Access-Control-Expose-Headers to contain %q, got %q", key, exposed)
				}
			}
		})
	}
}

func TestIntegrationListPapers(t *testing.T) {
	setupIntegrationTest(t)

	endpoint := testAPIPath + "/api/papers"
	resp := doGetRequest(t, endpoint)
	defer resp.Body.Close()

	assertStatusCode(t, resp, http.StatusOK)

	var papers []paperHttp.PaperResponse
	decodeJSON(t, resp, &papers)

	if len(papers) < 4 {
		t.Errorf("expected at least %d papers, got %d", 4, len(papers))
	}
}

func TestIntegrationGetPaperByUUID(t *testing.T) {
	setupIntegrationTest(t)

	uuid := "36583bb4-8cdc-554e-bcf5-f67b60d0b290"
	endpoint := testAPIPath + "/api/papers/" + uuid
	resp := doGetRequest(t, endpoint)
	defer resp.Body.Close()

	assertStatusCode(t, resp, http.StatusOK)

	var paper paperHttp.PaperResponse
	decodeJSON(t, resp, &paper)

	if paper.UUID != uuid {
		t.Fatalf("expected paper to have UUID %s, got %s", uuid, paper.UUID)
	}
}

func TestIntegrationDeletePaper(t *testing.T) {
	setupIntegrationTest(t)

	uuid := "0f324174-926b-585d-b121-3a1e3f7fee0b"
	endpoint := testAPIPath + "/api/papers/" + uuid
	resp := doDeleteRequest(t, endpoint)
	assertStatusCode(t, resp, http.StatusNoContent)

	resp = doGetRequest(t, endpoint)
	assertStatusCode(t, resp, http.StatusNotFound)
}

func TestIntegrationSavePaper(t *testing.T) {
	setupIntegrationTest(t)

	paper := paperHttp.PaperRequest{
		DOI:             "10.1109/some_DOI",
		Title:           "Test article",
		PublicationDate: paperHttp.PublicationDate{Year: 2026, Month: 8, Day: 13},
	}
	endpoint := testAPIPath + "/api/papers"
	resp := doPostRequest(t, endpoint, paper)
	defer resp.Body.Close()
	assertStatusCode(t, resp, http.StatusCreated)

	loc := resp.Header.Get("location")
	if loc == "" {
		t.Fatal("expected location header to be not empty")
	}

	var created paperHttp.PaperResponse
	decodeJSON(t, resp, &created)
	if created.UUID == "" {
		t.Fatal("expected UUID to be not empty")
	}

	if created.PublicationDate != paper.PublicationDate {
		t.Fatalf("created publication date = %#v, want %#v", created.PublicationDate, paper.PublicationDate)
	}
}
