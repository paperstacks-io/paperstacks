//go:build integration
// +build integration

package testintegration

import (
	"net/http"
	"testing"
)

func TestIntegrationHealthz(t *testing.T) {
	endpoint := testAPIPath() + "/healthz"
	resp := doGetRequest(t, endpoint)
	defer resp.Body.Close()

	assertStatusCode(t, resp, http.StatusNoContent)
	assertBody(t, resp, "")
}
