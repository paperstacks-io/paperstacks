package domain

type Document struct {
	Key         string
	PaperUUID   string
	UserID      string
	FileName    string
	ContentType string
	Size        int64
}
