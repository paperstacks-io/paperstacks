package integration

import (
	"net/http"
	"testing"
)

func TestIntegrationHealthz(t *testing.T) {
	app := startApplication(t)
	endpoint := app.baseURL + "/healthz"
	resp := app.doGetRequest(t, endpoint)
	defer resp.Body.Close()

	assertStatusCode(t, resp, http.StatusNoContent)
	assertBody(t, resp, "")
}
