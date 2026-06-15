package domain

type Document struct {
	UUID         string
	PaperUUID    string
	UploaderUUID string
	FileName     string
	ContentType  string
	Size         int64
	StorageURI   string
}
