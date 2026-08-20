package citation

type CSLAuthor struct {
	Given  string `json:"given"`
	Family string `json:"family"`
}

type CSLItem struct {
	Title  string      `json:"title"`
	Author []CSLAuthor `json:"author"`
}
