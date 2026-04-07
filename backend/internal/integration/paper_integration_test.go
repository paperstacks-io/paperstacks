//nolint:errcheck // to ignore not checking err when defer resp.Body.Close()
package integration

import (
	"math"
	"net/http"
	"net/url"
	"strconv"
	"testing"

	paperHttp "github.com/paperstacks.io/paperstacks/internal/paper/http"
)

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
	endpoint := testAPIPath + "/api/papers/uuid/" + uuid
	resp := doGetRequest(t, endpoint)
	defer resp.Body.Close()

	assertStatusCode(t, resp, http.StatusOK)

	var paper paperHttp.PaperResponse
	decodeJSON(t, resp, &paper)

	if paper.UUID != uuid {
		t.Fatalf("expected paper to have UUID %s, got %s", uuid, paper.UUID)
	}
}

func TestIntegrationGetPaperByDOI(t *testing.T) {
	doi := "10.1109/isese.2005.1541817"
	endpoint := testAPIPath + "/api/papers/doi/" + doi
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
	endpoint := testAPIPath + "/api/papers/doi/doesntexist"
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

	endpoint := testAPIPath + "/api/papers"
	resp := doPostRequest(t, endpoint, paperReq)
	defer resp.Body.Close()

	assertStatusCode(t, resp, http.StatusCreated)

	// Verify it was created
	verifyEndpoint := testAPIPath + "/api/papers/doi/" + paperReq.DOI
	verifyResp := doGetRequest(t, verifyEndpoint)
	defer verifyResp.Body.Close()

	assertStatusCode(t, verifyResp, http.StatusOK)
}

func TestIntegrationUpdatePaper(t *testing.T) {
	doi := "10.1109/isese.2005.1541817"
	endpoint := testAPIPath + "/api/papers/doi/" + doi

	resp := doGetRequest(t, endpoint)
	defer resp.Body.Close()
	var before paperHttp.PaperResponse
	decodeJSON(t, resp, &before)

	update := paperHttp.PaperRequest{
		UUID:  before.UUID,
		DOI:   before.DOI,
		Title: "Updated Paper",
	}

	resp = doPutRequest(t, endpoint, update)
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
	createEndpoint := testAPIPath + "/api/papers"
	doPostRequest(t, createEndpoint, paperBody)

	// Delete the paper
	deleteEndpoint := testAPIPath + "/api/papers/doi/" + doiToDelete
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
			u, err := url.Parse(testAPIPath + "/api/papers")
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
	endpoint := testAPIPath + "/api/papers?title=" + url.QueryEscape("doesntexist")
	resp := doGetRequest(t, endpoint)
	defer resp.Body.Close()

	assertStatusCode(t, resp, http.StatusOK)
	assertBody(t, resp, "[]\n")
}

func TestIntegrationPapersGetByKeyword(t *testing.T) {
	keyword := "Code review"
	endpoint := testAPIPath + "/api/papers?keyword=" + url.QueryEscape(keyword)

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
	endpoint := testAPIPath + "/api/papers?keyword=" + url.QueryEscape(keyword)

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

	endpoint := testAPIPath + "/api/papers?title=" + url.QueryEscape(title) +
		"&keyword=" + url.QueryEscape(keyword)

	resp := doGetRequest(t, endpoint)
	defer resp.Body.Close()

	assertStatusCode(t, resp, http.StatusOK)

	var papers []paperHttp.PaperResponse
	decodeJSON(t, resp, &papers)

	if len(papers) != 2 {
		t.Fatalf("expected 2 paper, got %d", len(papers))
	}

	if papers[0].Title != title {
		t.Fatalf("expected title %q, got %q", title, papers[0].Title)
	}
}

func TestIntegrationPapersSortByTitleDesc(t *testing.T) {
	endpoint := testAPIPath + "/api/papers?title=gui&sortBy=-title"
	resp := doGetRequest(t, endpoint)
	defer resp.Body.Close()

	assertStatusCode(t, resp, http.StatusOK)

	var papers []paperHttp.PaperResponse
	decodeJSON(t, resp, &papers)

	if len(papers) != 3 {
		t.Fatalf("expected 3 papers, got %d", len(papers))
	}

	titles := []string{
		"We Tried and Failed: An Experience Report on a Collaborative Workflow for GUI-based Testing",
		"Code review guidelines for GUI-based testing artifacts",
		"Augmented testing to support manual GUI-based regression testing: An empirical study",
	}

	for i, paper := range papers {
		if paper.Title != titles[i] {
			t.Fatalf("expected paper at index %d to have title %q, got %q", i, titles[i], paper.Title)
		}
	}
}

func TestIntegrationPapersSortByTitleAsc(t *testing.T) {
	endpoint := testAPIPath + "/api/papers?title=gui&sortBy=+title"
	resp := doGetRequest(t, endpoint)
	defer resp.Body.Close()

	assertStatusCode(t, resp, http.StatusOK)

	var papers []paperHttp.PaperResponse
	decodeJSON(t, resp, &papers)

	if len(papers) != 3 {
		t.Fatalf("expected 3 papers, got %d", len(papers))
	}

	titles := []string{
		"Augmented testing to support manual GUI-based regression testing: An empirical study",
		"Code review guidelines for GUI-based testing artifacts",
		"We Tried and Failed: An Experience Report on a Collaborative Workflow for GUI-based Testing",
	}

	for i, paper := range papers {
		if paper.Title != titles[i] {
			t.Fatalf("expected paper at index %d to have title %q, got %q", i, titles[i], paper.Title)
		}
	}
}

func TestIntegrationPapersSortByYearDesc(t *testing.T) {
	endpoint := testAPIPath + "/api/papers?title=gui&sortBy=-year"
	resp := doGetRequest(t, endpoint)
	defer resp.Body.Close()

	assertStatusCode(t, resp, http.StatusOK)

	var papers []paperHttp.PaperResponse
	decodeJSON(t, resp, &papers)

	if len(papers) != 3 {
		t.Fatalf("expected 3 papers, got %d", len(papers))
	}

	year := math.MaxInt
	for i, paper := range papers {
		paperYear, err := strconv.Atoi(paper.PublicationYear)
		if err != nil {
			t.Fatalf("expected paper at index %d to have a valid publication year, got %s", i, paper.PublicationYear)
		}

		if paperYear <= year {
			year = paperYear
		} else {
			t.Fatalf("exptected paper at index %d to have publication year less than or equal to %d, got %d", i, year, paperYear)
		}
	}
}

func TestIntegrationPapersSortByYearAsc(t *testing.T) {
	endpoint := testAPIPath + "/api/papers?title=gui&sortBy=+year"
	resp := doGetRequest(t, endpoint)
	defer resp.Body.Close()

	assertStatusCode(t, resp, http.StatusOK)

	var papers []paperHttp.PaperResponse
	decodeJSON(t, resp, &papers)

	if len(papers) != 3 {
		t.Fatalf("expected 3 papers, got %d", len(papers))
	}

	year := math.MinInt
	for i, paper := range papers {

		paperYear, err := strconv.Atoi(paper.PublicationYear)
		if err != nil {
			t.Fatalf("expected paper at index %d to have a valid publication year, got %s", i, paper.PublicationYear)
		}

		if paperYear >= year {
			year = paperYear
		} else {
			t.Fatalf("expected paper at index %d to have publication year greater then or equal to %d, got %d", i, year, paperYear)
		}
	}
}
