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
	t.Parallel()

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
		// bad request
		{query: "gui", sortBy: "bad-sort", page: 1, pageSize: 0, HTTPStatusCode: http.StatusBadRequest, expectedLen: 1, expectedFirstUUID: ""},
	}

	for _, tc := range tests {
		t.Run(tc.query+"_"+tc.sortBy, func(t *testing.T) {
			t.Parallel()

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

func TestIntegrationListPapers(t *testing.T) {
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
