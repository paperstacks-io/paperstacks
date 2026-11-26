package pdfextractor

import (
	"bytes"
	"context"
	"errors"
	"io"
	"sort"
	"strings"
	"testing"

	"github.com/paperstacks.io/paperstacks/internal/common/objectstorage"
)

type mockSyncStore struct {
	objects map[string][]byte
	err     error
}

func (m *mockSyncStore) Get(_ context.Context, key string) (*objectstorage.Object, error) {
	if m.err != nil {
		return nil, m.err
	}
	data, ok := m.objects[key]
	if !ok {
		return nil, objectstorage.ErrObjectNotFound
	}
	return &objectstorage.Object{
		Body: io.NopCloser(bytes.NewReader(data)),
		Info: objectstorage.ObjectInfo{Key: key, Size: int64(len(data))},
	}, nil
}

func (m *mockSyncStore) Put(_ context.Context, input objectstorage.PutObjectInput) (objectstorage.ObjectInfo, error) {
	if m.err != nil {
		return objectstorage.ObjectInfo{}, m.err
	}
	data, err := io.ReadAll(input.Body)
	if err != nil {
		return objectstorage.ObjectInfo{}, err
	}
	m.objects[input.Key] = data
	return objectstorage.ObjectInfo{
		Key:         input.Key,
		Size:        input.Size,
		ContentType: input.ContentType,
	}, nil
}

func (m *mockSyncStore) Walk(_ context.Context, prefix string, fn func(key string) error) error {
	if m.err != nil {
		return m.err
	}
	var keys []string
	for k := range m.objects {
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

type mockExtractor struct {
	xml []byte
	err error
}

func (m *mockExtractor) ExtractFulltext(_ context.Context, pdf io.Reader) ([]byte, error) {
	if _, err := io.ReadAll(pdf); err != nil {
		return nil, err
	}
	if m.err != nil {
		return nil, m.err
	}
	return m.xml, nil
}

type conditionalExtractor struct{}

func (conditionalExtractor) ExtractFulltext(_ context.Context, pdf io.Reader) ([]byte, error) {
	data, err := io.ReadAll(pdf)
	if err != nil {
		return nil, err
	}
	if !bytes.HasPrefix(data, []byte("%PDF-")) {
		return nil, ErrInvalidPDF
	}
	return []byte("<TEI>ok</TEI>"), nil
}

func TestSync(t *testing.T) {
	t.Parallel()

	store := &mockSyncStore{
		objects: map[string][]byte{
			"pdf/01-existing.pdf":       []byte("%PDF-1.4\n1"),
			"pdf/01-existing.xml":       []byte("<TEI>existing</TEI>"),
			"pdf/02-missing.pdf":        []byte("%PDF-1.4\n2"),
			"pdf/03-dotted.pdf":         []byte("%PDF-1.4\n3"),
			"pdf/03-dotted.rev1.pdf":    []byte("%PDF-1.4\n3r"),
			"pdf/03-dotted.rev1.xml":    []byte("<TEI>existing</TEI>"),
			"pdf/04-orphan.xml":         []byte("<TEI>orphan</TEI>"),
			"pdf/05-last.pdf":           []byte("%PDF-1.4\n5"),
			"other/ignored-by-pref.pdf": []byte("%PDF-1.4\nother"),
		},
	}
	extractor := &mockExtractor{xml: []byte("<TEI>synced</TEI>")}

	if err := Sync(context.Background(), store, extractor, "pdf/", nil); err != nil {
		t.Fatalf("Sync() error = %v", err)
	}

	for _, wantXML := range []string{
		"pdf/02-missing.xml",
		"pdf/03-dotted.xml",
		"pdf/05-last.xml",
	} {
		got, ok := store.objects[wantXML]
		if !ok {
			t.Fatalf("expected %q to be created", wantXML)
		}
		if string(got) != "<TEI>synced</TEI>" {
			t.Fatalf("%s body = %q, want %q", wantXML, string(got), "<TEI>synced</TEI>")
		}
	}

	if string(store.objects["pdf/01-existing.xml"]) != "<TEI>existing</TEI>" {
		t.Fatalf("existing xml was overwritten: %q", string(store.objects["pdf/01-existing.xml"]))
	}
	if _, ok := store.objects["other/ignored-by-pref.xml"]; ok {
		t.Fatal("unexpected extraction outside prefix")
	}
}

func TestSyncContinuesOnSingleError(t *testing.T) {
	t.Parallel()

	store := &mockSyncStore{
		objects: map[string][]byte{
			"pdf/1-bad.pdf":  []byte("corrupt"),
			"pdf/2-good.pdf": []byte("%PDF-1.4\ngood"),
		},
	}

	if err := Sync(context.Background(), store, conditionalExtractor{}, "pdf/", nil); err != nil {
		t.Fatalf("Sync() error = %v, want nil", err)
	}

	if _, ok := store.objects["pdf/1-bad.xml"]; ok {
		t.Fatal("did not expect XML for failing PDF")
	}
	if _, ok := store.objects["pdf/2-good.xml"]; !ok {
		t.Fatal("expected pdf/2-good.xml to be created after recovering from error")
	}
}

func TestSyncStopsOnContextCancel(t *testing.T) {
	t.Parallel()

	store := &mockSyncStore{
		objects: map[string][]byte{
			"pdf/1.pdf": []byte("%PDF-1.4"),
		},
	}
	extractor := &mockExtractor{xml: []byte("<TEI/>")}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	if err := Sync(ctx, store, extractor, "pdf/", nil); !errors.Is(err, context.Canceled) {
		t.Fatalf("Sync() error = %v, want %v", err, context.Canceled)
	}
}
