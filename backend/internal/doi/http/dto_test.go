package http

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/paperstacks.io/paperstacks/internal/doi/domain"
)

func TestNewMetadataResponseDoesNotExposeLookupOnlyKeywords(t *testing.T) {
	t.Parallel()

	response := NewMetadataResponse(&domain.Metadata{
		DOI:      "10.1000/182",
		Title:    "Example title",
		Keywords: []string{"quality assurance", "gui testing"},
	})

	body, err := json.Marshal(response)
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}

	if strings.Contains(string(body), `"keywords"`) {
		t.Fatalf("response JSON unexpectedly exposed keywords: %s", string(body))
	}
}
