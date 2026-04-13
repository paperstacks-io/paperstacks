//nolint:errcheck // to ignore not checking err when defer resp.Body.Close()
package integration

import (
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"testing"

	"github.com/paperstacks.io/paperstacks/internal/paper/domain"
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
		{query: "", sortBy: "-title", page: 0, pageSize: 0, HTTPStatusCode: http.StatusOK, expectedLen: 4, expectedFirstUUID: "a4b06"},
		{query: "", sortBy: "+title", page: 0, pageSize: 0, HTTPStatusCode: http.StatusOK, expectedLen: 4, expectedFirstUUID: "6752d"},
		{query: "", sortBy: "-year", page: 0, pageSize: 0, HTTPStatusCode: http.StatusOK, expectedLen: 4, expectedFirstUUID: "6752d"},
		{query: "", sortBy: "+year", page: 0, pageSize: 0, HTTPStatusCode: http.StatusOK, expectedLen: 4, expectedFirstUUID: "bc627"},
		// title
		{query: "we tried", sortBy: "", page: 1, pageSize: 0, HTTPStatusCode: http.StatusOK, expectedLen: 1, expectedFirstUUID: "a4b06"},
		{query: "regression testing", sortBy: "", page: 1, pageSize: 0, HTTPStatusCode: http.StatusOK, expectedLen: 1, expectedFirstUUID: "6752d"},
		// keyword
		{query: "linux", sortBy: "", page: 1, pageSize: 0, HTTPStatusCode: http.StatusOK, expectedLen: 1, expectedFirstUUID: "a4b06"},
		{query: "usability prob", sortBy: "", page: 1, pageSize: 0, HTTPStatusCode: http.StatusOK, expectedLen: 1, expectedFirstUUID: "bc627"},
		// pagignation
		{query: "gui", sortBy: "", page: 1, pageSize: 0, HTTPStatusCode: http.StatusOK, expectedLen: 3, expectedFirstUUID: "202a0"},
		{query: "gui", sortBy: "", page: 1, pageSize: 2, HTTPStatusCode: http.StatusOK, expectedLen: 2, expectedFirstUUID: "202a0"},
		{query: "gui", sortBy: "", page: 2, pageSize: 2, HTTPStatusCode: http.StatusOK, expectedLen: 1, expectedFirstUUID: "6752de"},
		{query: "gui", sortBy: "", page: 1, pageSize: 3, HTTPStatusCode: http.StatusOK, expectedLen: 3, expectedFirstUUID: "202a0"},
		{query: "", sortBy: "", page: 5, pageSize: 2, HTTPStatusCode: http.StatusOK, expectedLen: 0, expectedFirstUUID: ""},
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
			pageSize:         2,
			expectedPage:     "1",
			expectedPageSize: "2",
			expectedTotal:    "3",
			expectedHasNext:  "true",
		},
		{
			name:             "second page has no next",
			query:            "gui",
			page:             2,
			pageSize:         2,
			expectedPage:     "2",
			expectedPageSize: "2",
			expectedTotal:    "3",
			expectedHasNext:  "false",
		},
		{
			name:             "defaults from service",
			query:            "gui",
			expectedPage:     "1",
			expectedPageSize: "10",
			expectedTotal:    "3",
			expectedHasNext:  "false",
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

	uuid := "a4b065f1-1b88-4f50-a7fe-1177f3489fcf"
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

	uuid := "a4b065f1-1b88-4f50-a7fe-1177f3489fcf"
	endpoint := testAPIPath + "/api/papers/" + uuid
	resp := doDeleteRequest(t, endpoint)
	assertStatusCode(t, resp, http.StatusNoContent)

	resp = doGetRequest(t, endpoint)
	assertStatusCode(t, resp, http.StatusNotFound)
}

func TestIntegrationSavePaper(t *testing.T) {
	setupIntegrationTest(t)

	paper := domain.Paper{
		DOI:   "10.1109/some_DOI",
		Title: "Test article",
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
}
