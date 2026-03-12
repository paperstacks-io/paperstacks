//go:build integration
// +build integration

package testintegration

import (
	"net/http"
	"testing"

	"github.com/paperstacks.io/paperstacks/internal/server_"
)

func TestIntegrationPapersGetAll(t *testing.T) {
	endpoint := testAPIPath() + "/papers/"
	resp := doGetRequest(t, endpoint)
	defer resp.Body.Close()

	assertStatusCode(t, resp, http.StatusOK)

	var papers []server_.PaperResponse
	decodeJSON(t, resp, &papers)

	if len(papers) < 3 {
		t.Errorf("expected at least %d papers, got %d", 3, len(papers))
	}
}

func TestIntegrationPapersGetSingle(t *testing.T) {
	endpoint := testAPIPath() + "/papers/1"
	resp := doGetRequest(t, endpoint)
	defer resp.Body.Close()

	assertStatusCode(t, resp, http.StatusOK)

	var paper server_.PaperResponse
	decodeJSON(t, resp, &paper)

	if paper.DOI != "1" {
		t.Fatalf("expected paper_ to have DOI %s, got %s", "1", paper.DOI)
	}
}

func TestIntegrationPapersGetSingleUnknown(t *testing.T) {
	endpoint := testAPIPath() + "/papers/doesntexist"
	resp := doGetRequest(t, endpoint)
	defer resp.Body.Close()

	assertStatusCode(t, resp, http.StatusNotFound)
	assertBody(t, resp, "paper_ not found\n")
}

func TestIntegrationPapersCreate(t *testing.T) {
	paperBody := server_.CreatePaperRequest{
		DOI:   "4",
		Title: "Created Paper",
	}

	endpoint := testAPIPath() + "/papers/"
	resp := doPostRequest(t, endpoint, paperBody)
	defer resp.Body.Close()

	assertStatusCode(t, resp, http.StatusCreated)

	// Verify it was created
	verifyEndpoint := testAPIPath() + "/papers/4"
	verifyResp := doGetRequest(t, verifyEndpoint)
	defer verifyResp.Body.Close()

	assertStatusCode(t, verifyResp, http.StatusOK)
}

func TestIntegrationPapersUpdate(t *testing.T) {
	paperBody := server_.UpdatePaperRequest{
		Title: "Updated Paper Two",
	}

	endpoint := testAPIPath() + "/papers/2"
	resp := doPutRequest(t, endpoint, paperBody)
	defer resp.Body.Close()

	assertStatusCode(t, resp, http.StatusOK)

	// Verify it was updated
	verifyResp := doGetRequest(t, endpoint)
	defer verifyResp.Body.Close()

	var paper server_.PaperResponse
	decodeJSON(t, verifyResp, &paper)

	if paper.Title != "Updated Paper Two" {
		t.Errorf("expected title %s, got %s", "Updated Paper Two", paper.Title)
	}
}

func TestIntegrationPapersDelete(t *testing.T) {
	doiToDelete := "to-be-deleted"
	paperBody := server_.CreatePaperRequest{
		DOI:   doiToDelete,
		Title: "To Be Deleted",
	}

	// Create the paper_ to delete
	createEndpoint := testAPIPath() + "/papers/"
	createResp := doPostRequest(t, createEndpoint, paperBody)
	createResp.Body.Close()

	// Delete the paper_
	deleteEndpoint := testAPIPath() + "/papers/" + doiToDelete
	deleteResp := doDeleteRequest(t, deleteEndpoint)
	defer deleteResp.Body.Close()

	assertStatusCode(t, deleteResp, http.StatusNoContent)

	// Verify it was deleted
	verifyResp := doGetRequest(t, deleteEndpoint)
	defer verifyResp.Body.Close()

	assertStatusCode(t, verifyResp, http.StatusNotFound)
}
