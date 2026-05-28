package domain

type Document struct {
	UUID        string
	PaperUUID   string
	FileName    string
	ContentType string
	Size        int64
	StorageURI  string
}
