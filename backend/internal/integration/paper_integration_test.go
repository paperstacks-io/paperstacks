//nolint:errcheck // to ignore not checking err when defer resp.Body.Close()
package integration

import (
	"net/http"
	"net/url"
	"testing"

	paperHttp "github.com/paperstacks.io/paperstacks/internal/paper/http"
)

func TestIntegrationListPapers(t *testing.T) {
	endpoint := testAPIPath + "/papers/"
	resp := doGetRequest(t, endpoint)
	defer resp.Body.Close()

	assertStatusCode(t, resp, http.StatusOK)

	var papers []paperHttp.PaperResponse
	decodeJSON(t, resp, &papers)

	if len(papers) < 4 {
		t.Errorf("expected at least %d papers, got %d", 4, len(papers))
	}
}

func TestIntegrationGetPaperByDOI(t *testing.T) {
	doi := "10.1109/isese.2005.1541817"
	endpoint := testAPIPath + "/papers/doi/" + doi
	resp := doGetRequest(t, endpoint)
	defer resp.Body.Close()

	assertStatusCode(t, resp, http.StatusOK)

	var paper paperHttp.PaperResponse
	decodeJSON(t, resp, &paper)

	if paper.DOI != doi {
		t.Fatalf("expected paper to have DOI %s, got %s", doi, paper.DOI)
	}
}

func TestIntegrationPapersGetPaperByDOIUnknown(t *testing.T) {
	endpoint := testAPIPath + "/papers/doi/doesntexist"
	resp := doGetRequest(t, endpoint)
	defer resp.Body.Close()

	assertStatusCode(t, resp, http.StatusNotFound)
	assertBody(t, resp, "paper not found\n")
}

func TestIntegrationSavePaper(t *testing.T) {
	paperReq := paperHttp.PaperRequest{
		DOI:   "10.1000/new",
		Title: "Created Paper",
	}

	endpoint := testAPIPath + "/papers/"
	resp := doPostRequest(t, endpoint, paperReq)
	defer resp.Body.Close()

	assertStatusCode(t, resp, http.StatusCreated)

	// Verify it was created
	verifyEndpoint := testAPIPath + "/papers/doi/" + paperReq.DOI
	verifyResp := doGetRequest(t, verifyEndpoint)
	defer verifyResp.Body.Close()

	assertStatusCode(t, verifyResp, http.StatusOK)
}

func TestIntegrationUpdatePaper(t *testing.T) {
	doi := "10.1109/isese.2005.1541817"
	paperBody := paperHttp.PaperRequest{
		Title: "Updated Paper",
	}

	endpoint := testAPIPath + "/papers/doi/" + doi
	resp := doPutRequest(t, endpoint, paperBody)
	defer resp.Body.Close()

	assertStatusCode(t, resp, http.StatusNoContent)

	// Verify it was updated
	verifyResp := doGetRequest(t, endpoint)
	defer verifyResp.Body.Close()

	var paper paperHttp.PaperResponse
	decodeJSON(t, verifyResp, &paper)

	if paper.Title != "Updated Paper" {
		t.Errorf("expected title %s, got %s", "Updated Paper", paper.Title)
	}
}

func TestIntegrationDeletePaper(t *testing.T) {
	doiToDelete := "to-be-deleted"
	paperBody := paperHttp.PaperRequest{
		DOI:   doiToDelete,
		Title: "To Be Deleted",
	}

	// Create the paper to delete
	createEndpoint := testAPIPath + "/papers/doi/"
	doPostRequest(t, createEndpoint, paperBody)

	// Delete the paper
	deleteEndpoint := testAPIPath + "/papers/doi/" + doiToDelete
	deleteResp := doDeleteRequest(t, deleteEndpoint)
	defer deleteResp.Body.Close()

	assertStatusCode(t, deleteResp, http.StatusNoContent)

	// Verify it was deleted
	verifyResp := doGetRequest(t, deleteEndpoint)
	defer verifyResp.Body.Close()

	assertStatusCode(t, verifyResp, http.StatusNotFound)
}

func TestIntegrationGetPaperByTitle(t *testing.T) {
	titles := []string{
		"Code review guidelines for GUI-based testing artifacts",
		"We Tried and Failed: An Experience Report on a Collaborative Workflow for GUI-based Testing",
		"Augmented testing to support manual GUI-based regression testing: An empirical study",
	}

	for _, title := range titles {
		t.Run(title, func(t *testing.T) {
			u, err := url.Parse(testAPIPath + "/papers")
			if err != nil {
				t.Fatalf("failed to parse url: %v", err)
			}

			q := u.Query()
			q.Set("title", title)
			u.RawQuery = q.Encode()

			resp := doGetRequest(t, u.String())

			defer resp.Body.Close()

			assertStatusCode(t, resp, http.StatusOK)

			var papers []paperHttp.PaperResponse
			decodeJSON(t, resp, &papers)

			if len(papers) == 0 {
				t.Fatal("expected at least one paper, got none")
			}

			for i, paper := range papers {
				if paper.Title != title {
					t.Fatalf("expected paper at index %d to have title %q, got %q", i, title, paper.Title)
				}
			}
		})
	}
}

func TestIntegrationPapersGetPaperByTitleUnknown(t *testing.T) {
	endpoint := testAPIPath + "/papers?title=" + url.QueryEscape("doesntexist")
	resp := doGetRequest(t, endpoint)
	defer resp.Body.Close()

	assertStatusCode(t, resp, http.StatusOK)
	assertBody(t, resp, "[]\n")
}

func TestIntegrationPapersGetByKeyword(t *testing.T) {
	keyword := "Code review"
	endpoint := testAPIPath + "/papers?keyword=" + url.QueryEscape(keyword)

	resp := doGetRequest(t, endpoint)
	defer resp.Body.Close()

	assertStatusCode(t, resp, http.StatusOK)

	var papers []paperHttp.PaperResponse
	decodeJSON(t, resp, &papers)

	if len(papers) != 2 {
		t.Fatalf("expected 2 papers, got %d", len(papers))
	}
}

func TestIntegrationPapersGetByKeywordWithSpaces(t *testing.T) {
	keyword := "   Code review   "
	endpoint := testAPIPath + "/papers?keyword=" + url.QueryEscape(keyword)

	resp := doGetRequest(t, endpoint)
	defer resp.Body.Close()

	assertStatusCode(t, resp, http.StatusOK)

	var papers []paperHttp.PaperResponse
	decodeJSON(t, resp, &papers)

	if len(papers) != 2 {
		t.Fatalf("expected 2 papers, got %d", len(papers))
	}
}

func TestIntegrationPapersSearchByTitleAndKeyword(t *testing.T) {
	title := "Code review guidelines for GUI-based testing artifacts"
	keyword := "Code review"

	endpoint := testAPIPath + "/papers?title=" + url.QueryEscape(title) +
		"&keyword=" + url.QueryEscape(keyword)

	resp := doGetRequest(t, endpoint)
	defer resp.Body.Close()

	assertStatusCode(t, resp, http.StatusOK)

	var papers []paperHttp.PaperResponse
	decodeJSON(t, resp, &papers)

	if len(papers) != 1 {
		t.Fatalf("expected 1 paper, got %d", len(papers))
	}

	if papers[0].Title != title {
		t.Fatalf("expected title %q, got %q", title, papers[0].Title)
	}
}
