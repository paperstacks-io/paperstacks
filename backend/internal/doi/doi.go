package doi

type Metadata struct {
	DOI       string   `json:"doi"`
	Title     string   `json:"title"`
	Publisher string   `json:"publisher"`
	Type      string   `json:"type"`
	Authors   []string `json:"authors"`
	Published string   `json:"published"`
	URL       string   `json:"url"`
}
