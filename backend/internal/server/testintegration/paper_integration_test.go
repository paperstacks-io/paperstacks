//nolint:errcheck // to ignore not checking err when defer resp.Body.Close()
package testintegration

import (
	"net/http"
	"testing"

	"github.com/paperstacks.io/paperstacks/internal/server"
)

func TestIntegrationPapersGetAll(t *testing.T) {
	endpoint := testAPIPath() + "/papers/"
	resp := doGetRequest(t, endpoint)
	defer resp.Body.Close()

	assertStatusCode(t, resp, http.StatusOK)

	var papers []server.PaperResponse
	decodeJSON(t, resp, &papers)

	if len(papers) < 4 {
		t.Errorf("expected at least %d papers, got %d", 4, len(papers))
	}
}

func TestIntegrationPapersGetSingle(t *testing.T) {
	doi := "10.1109/isese.2005.1541817"
	endpoint := testAPIPath() + "/papers/doi/" + doi
	resp := doGetRequest(t, endpoint)
	defer resp.Body.Close()

	assertStatusCode(t, resp, http.StatusOK)

	var paper server.PaperResponse
	decodeJSON(t, resp, &paper)

	if paper.DOI != doi {
		t.Fatalf("expected paper to have DOI %s, got %s", doi, paper.DOI)
	}
}

func TestIntegrationPapersGetSingleUnknown(t *testing.T) {
	endpoint := testAPIPath() + "/papers/doi/doesntexist"
	resp := doGetRequest(t, endpoint)
	defer resp.Body.Close()

	assertStatusCode(t, resp, http.StatusNotFound)
	assertBody(t, resp, "paper not found\n")
}

func TestIntegrationPapersCreate(t *testing.T) {
	paperBody := server.CreatePaperRequest{
		DOI:   "10.1000/new",
		Title: "Created Paper",
	}

	endpoint := testAPIPath() + "/papers/"
	resp := doPostRequest(t, endpoint, paperBody)
	defer resp.Body.Close()

	assertStatusCode(t, resp, http.StatusCreated)

	// Verify it was created
	verifyEndpoint := testAPIPath() + "/papers/doi/" + paperBody.DOI
	verifyResp := doGetRequest(t, verifyEndpoint)
	defer verifyResp.Body.Close()

	assertStatusCode(t, verifyResp, http.StatusOK)
}

func TestIntegrationPapersUpdate(t *testing.T) {
	doi := "10.1109/isese.2005.1541817"
	paperBody := server.UpdatePaperRequest{
		Title: "Updated Paper",
	}

	endpoint := testAPIPath() + "/papers/doi/" + doi
	resp := doPutRequest(t, endpoint, paperBody)
	defer resp.Body.Close()

	assertStatusCode(t, resp, http.StatusOK)

	// Verify it was updated
	verifyResp := doGetRequest(t, endpoint)
	defer verifyResp.Body.Close()

	var paper server.PaperResponse
	decodeJSON(t, verifyResp, &paper)

	if paper.Title != "Updated Paper" {
		t.Errorf("expected title %s, got %s", "Updated Paper", paper.Title)
	}
}

func TestIntegrationPapersDelete(t *testing.T) {
	doiToDelete := "to-be-deleted"
	paperBody := server.CreatePaperRequest{
		DOI:   doiToDelete,
		Title: "To Be Deleted",
	}

	// Create the paper to delete
	createEndpoint := testAPIPath() + "/papers/doi/"
	doPostRequest(t, createEndpoint, paperBody)

	// Delete the paper
	deleteEndpoint := testAPIPath() + "/papers/doi/" + doiToDelete
	deleteResp := doDeleteRequest(t, deleteEndpoint)
	defer deleteResp.Body.Close()

	assertStatusCode(t, deleteResp, http.StatusNoContent)

	// Verify it was deleted
	verifyResp := doGetRequest(t, deleteEndpoint)
	defer verifyResp.Body.Close()

	assertStatusCode(t, verifyResp, http.StatusNotFound)
}
