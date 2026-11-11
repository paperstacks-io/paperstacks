package application

import (
	"bytes"
	"context"
	"errors"
	"io"
	"sort"
	"strings"
	"testing"

	"github.com/paperstacks.io/paperstacks/internal/common/objectstorage"
	"github.com/paperstacks.io/paperstacks/internal/paper/domain"
	paperMemory "github.com/paperstacks.io/paperstacks/internal/paper/repository/memory"
)

type walkableMockStorage struct {
	mockObjectStorage
}

func (w walkableMockStorage) Walk(_ context.Context, prefix string, fn func(key string) error) error {
	if w.err != nil {
		return w.err
	}
	var keys []string
	for k := range w.objects {
		if strings.HasPrefix(k, prefix) {
			keys = append(keys, k)
		}
	}
	sort.Strings(keys)
	for _, k := range keys {
		if err := fn(k); err != nil {
			return err
		}
	}
	return nil
}

type mockXMLExtractor struct {
	gotPDF []byte
	xml    []byte
	err    error
}

func (m *mockXMLExtractor) ExtractFulltext(_ context.Context, pdf io.Reader) ([]byte, error) {
	m.gotPDF, _ = io.ReadAll(pdf)
	return m.xml, m.err
}

func TestExtractPDFXML(t *testing.T) {
	ctx := context.Background()
	objects := map[string]storedPDF{
		"pdf/test.pdf": {body: []byte("%PDF-1.4\ncontent")},
	}
	service := NewPaperService(paperMemory.NewRepository(), mockObjectStorage{objects: objects})

	// Without extractor configured, returns empty string and nil error.
	if xmlKey, err := service.ExtractPDFXML(ctx, "pdf/test.pdf", strings.NewReader("%PDF-1.4")); err != nil || xmlKey != "" {
		t.Fatalf("ExtractPDFXML() without extractor = %q, %v; want empty, nil", xmlKey, err)
	}

	extractor := &mockXMLExtractor{xml: []byte("<TEI>extracted</TEI>")}
	service.SetXMLExtractor(extractor)

	// Missing PDF in storage returns ErrObjectNotFound.
	if _, err := service.ExtractPDFXML(ctx, "pdf/missing.pdf", strings.NewReader("%PDF-1.4")); !errors.Is(err, objectstorage.ErrObjectNotFound) {
		t.Fatalf("ExtractPDFXML(missing) error = %v, want %v", err, objectstorage.ErrObjectNotFound)
	}

	// Direct reader extraction (no S3 download).
	xmlKey, err := service.ExtractPDFXML(ctx, " pdf/test.pdf ", strings.NewReader("%PDF-1.4\ndirect"))
	if err != nil {
		t.Fatalf("ExtractPDFXML() error = %v", err)
	}
	if xmlKey != "pdf/test.xml" {
		t.Fatalf("xmlKey = %q, want %q", xmlKey, "pdf/test.xml")
	}
	if string(extractor.gotPDF) != "%PDF-1.4\ndirect" {
		t.Fatalf("extractor got PDF = %q", string(extractor.gotPDF))
	}

	// Stored PDF extraction (used by CLI).
	delete(objects, "pdf/test.xml")
	xmlKey, err = service.ExtractStoredPDFXML(ctx, " pdf/test.pdf ")
	if err != nil {
		t.Fatalf("ExtractStoredPDFXML() error = %v", err)
	}
	if xmlKey != "pdf/test.xml" {
		t.Fatalf("xmlKey = %q, want %q", xmlKey, "pdf/test.xml")
	}
	if string(extractor.gotPDF) != "%PDF-1.4\ncontent" {
		t.Fatalf("extractor got PDF = %q", string(extractor.gotPDF))
	}
	storedXML, ok := objects["pdf/test.xml"]
	if !ok {
		t.Fatal("expected pdf/test.xml to be stored in object storage")
	}
	if string(storedXML.body) != "<TEI>extracted</TEI>" {
		t.Fatalf("stored XML = %q, want %q", string(storedXML.body), "<TEI>extracted</TEI>")
	}
}

func TestDeletePDFAlsoDeletesXML(t *testing.T) {
	ctx := context.Background()
	const uuid = "test-delete-xml"
	repo := paperMemory.NewRepository()
	if _, err := repo.Save(ctx, domain.Paper{UUID: uuid, PDFStorageKey: "pdf/" + uuid + ".pdf"}); err != nil {
		t.Fatal(err)
	}
	objects := map[string]storedPDF{
		"pdf/" + uuid + ".pdf": {},
		"pdf/" + uuid + ".xml": {},
	}
	service := NewPaperService(repo, mockObjectStorage{objects: objects})
	if err := service.DeletePDF(ctx, uuid); err != nil {
		t.Fatalf("DeletePDF error = %v", err)
	}
	if _, ok := objects["pdf/"+uuid+".pdf"]; ok {
		t.Fatal("expected PDF to be deleted")
	}
	if _, ok := objects["pdf/"+uuid+".xml"]; ok {
		t.Fatal("expected XML to be deleted")
	}
}

func TestSyncPDFXMLs(t *testing.T) {
	t.Parallel()

	objects := map[string]storedPDF{
		"pdf/01-existing.pdf":       {body: []byte("%PDF-1.4\n1")},
		"pdf/01-existing.xml":       {body: []byte("<TEI>existing</TEI>")},
		"pdf/02-missing.pdf":        {body: []byte("%PDF-1.4\n2")},
		"pdf/03-dotted.pdf":         {body: []byte("%PDF-1.4\n3")},
		"pdf/03-dotted.rev1.pdf":    {body: []byte("%PDF-1.4\n3r")},
		"pdf/03-dotted.rev1.xml":    {body: []byte("<TEI>existing</TEI>")},
		"pdf/04-orphan.xml":         {body: []byte("<TEI>orphan</TEI>")},
		"pdf/05-last.pdf":           {body: []byte("%PDF-1.4\n5")},
		"other/ignored-by-pref.pdf": {body: []byte("%PDF-1.4\nother")},
	}
	service := NewPaperService(paperMemory.NewRepository(), walkableMockStorage{mockObjectStorage{objects: objects}})
	service.SetXMLExtractor(&mockXMLExtractor{xml: []byte("<TEI>synced</TEI>")})

	if err := service.SyncPDFXMLs(context.Background(), "pdf/", nil); err != nil {
		t.Fatalf("SyncPDFXMLs() error = %v", err)
	}

	for _, wantXML := range []string{
		"pdf/02-missing.xml",
		"pdf/03-dotted.xml",
		"pdf/05-last.xml",
	} {
		got, ok := objects[wantXML]
		if !ok {
			t.Fatalf("expected %q to be created", wantXML)
		}
		if string(got.body) != "<TEI>synced</TEI>" {
			t.Fatalf("%s body = %q, want %q", wantXML, string(got.body), "<TEI>synced</TEI>")
		}
	}

	if string(objects["pdf/01-existing.xml"].body) != "<TEI>existing</TEI>" {
		t.Fatalf("existing xml was overwritten: %q", string(objects["pdf/01-existing.xml"].body))
	}
	if _, ok := objects["other/ignored-by-pref.xml"]; ok {
		t.Fatal("unexpected extraction outside prefix")
	}
}

func TestSyncPDFXMLsContinuesOnSingleError(t *testing.T) {
	t.Parallel()

	objects := map[string]storedPDF{
		"pdf/1-bad.pdf":  {body: []byte("corrupt")},
		"pdf/2-good.pdf": {body: []byte("%PDF-1.4\ngood")},
	}
	service := NewPaperService(paperMemory.NewRepository(), walkableMockStorage{mockObjectStorage{objects: objects}})
	service.SetXMLExtractor(&conditionalXMLExtractor{})

	if err := service.SyncPDFXMLs(context.Background(), "pdf/", nil); err != nil {
		t.Fatalf("SyncPDFXMLs() error = %v, want nil", err)
	}

	if _, ok := objects["pdf/1-bad.xml"]; ok {
		t.Fatal("did not expect XML for failing PDF")
	}
	if _, ok := objects["pdf/2-good.xml"]; !ok {
		t.Fatal("expected pdf/2-good.xml to be created after recovering from error")
	}
}

func TestSyncPDFXMLsStopsOnContextCancel(t *testing.T) {
	t.Parallel()

	objects := map[string]storedPDF{
		"pdf/1.pdf": {body: []byte("%PDF-1.4")},
	}
	service := NewPaperService(paperMemory.NewRepository(), walkableMockStorage{mockObjectStorage{objects: objects}})
	service.SetXMLExtractor(&mockXMLExtractor{xml: []byte("<TEI/>")})

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	if err := service.SyncPDFXMLs(ctx, "pdf/", nil); !errors.Is(err, context.Canceled) {
		t.Fatalf("SyncPDFXMLs() error = %v, want %v", err, context.Canceled)
	}
}

type conditionalXMLExtractor struct{}

func (conditionalXMLExtractor) ExtractFulltext(_ context.Context, pdf io.Reader) ([]byte, error) {
	data, _ := io.ReadAll(pdf)
	if !bytes.HasPrefix(data, []byte("%PDF-")) {
		return nil, errors.New("invalid pdf")
	}
	return []byte("<TEI>ok</TEI>"), nil
}
