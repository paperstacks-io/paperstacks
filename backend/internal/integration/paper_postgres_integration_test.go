package integration

import (
	"context"
	"database/sql"
	"errors"
	"os"
	"reflect"
	"testing"
	"time"
	"uuid"

	"github.com/paperstacks.io/paperstacks/internal/paper/domain"
	"github.com/paperstacks.io/paperstacks/internal/paper/repository/postgres"
)

// Uses an initialized test database, with app_rw credentials. Each run owns
// unique paper IDs and removes only those records when finished.
func TestPaperPostgresRepository(t *testing.T) {
	dsn := os.Getenv("PAPER_TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("set PAPER_TEST_DATABASE_URL to run PostgreSQL repository tests")
	}
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })
	db.SetMaxOpenConns(1)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		t.Fatal(err)
	}
	repo := postgres.NewRepository(db)
	id, otherID, duplicateID := uuid.New().String(), uuid.New().String(), uuid.New().String()
	t.Cleanup(func() {
		cleanup, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		for _, id := range []string{id, otherID, duplicateID} {
			if err := repo.Delete(cleanup, id); err != nil && !errors.Is(err, domain.ErrPaperNotFound) {
				t.Error(err)
			}
		}
	})
	original := domain.Paper{
		UUID: id, DOI: "10.6700/" + id, Title: "Paper " + id, TitleShort: "Short",
		PublicationDate: domain.Date{Year: 2026, Month: 10, Day: 18}, Type: domain.PublicationType("journal-article"),
		PublicationStatus: "published", PublicationStatusTimestamp: "2026-10-18T10:00:00Z",
		Abstract: "Abstract", Keywords: []string{"science", id}, PDFURL: "https://example.invalid/paper.pdf",
		Authors: []domain.Author{{NameFirst: "Ada", NameLast: "Lovelace", Affiliation: "Lab " + id}, {NameFirst: "Grace", NameLast: "Hopper"}},
		Metadata: domain.Metadata{Publisher: "Publisher", JournalTitle: "Journal", JournalAbbrev: "J", BookTitle: "Book", SeriesTitle: "Series",
			EventTitle: "Conference", EventLocation: "Berlin", Institution: "Institute", Pages: "1-9", Volume: "S1", Issue: "A",
			ISBN: []string{"isbn"}, ISSN: []string{"issn"}, References: []string{"ref"}, License: "CC-BY", Copyright: "Copyright",
			Funding: "Grant", DataSource: "crossref", DataSourceTimestamp: "2026-10-18T10:00:00Z"},
	}
	saved, err := repo.Save(ctx, original)
	if err != nil {
		t.Fatal(err)
	}
	loaded, err := repo.GetByUUID(ctx, id)
	if err != nil || !reflect.DeepEqual(saved, loaded) {
		t.Fatalf("round trip: got %+v, %v; want %+v", loaded, err, saved)
	}
	byDOI, err := repo.GetByDOI(ctx, original.DOI)
	if err != nil || byDOI.UUID != id {
		t.Fatalf("DOI lookup: %+v, %v", byDOI, err)
	}
	duplicate := original
	duplicate.UUID = duplicateID
	if _, err := repo.Save(ctx, duplicate); !errors.Is(err, domain.ErrPaperAlreadyExists) {
		t.Fatalf("duplicate DOI: %v", err)
	}
	if _, err := repo.Save(ctx, domain.Paper{UUID: otherID, Title: "Other " + id}); err != nil {
		t.Fatal(err)
	}
	for _, sortBy := range []string{"title", "year"} {
		result, err := repo.Search(ctx, domain.SearchOptions{Query: id, SortBy: sortBy, Page: 1, PageSize: 1})
		if err != nil || result.Total != 2 || len(result.Items) != 1 || !result.HasNext {
			t.Fatalf("search %s: %+v, %v", sortBy, result, err)
		}
	}
	items, err := repo.List(ctx)
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, p := range items {
		if p.UUID == id {
			found = true
		}
	}
	if !found {
		t.Fatal("paper missing from list")
	}
	// Fail after updating the parent and removing children: everything must roll back.
	invalid := original
	invalid.Title = "Must roll back"
	invalid.Metadata.DataSourceTimestamp = "not-a-date"
	if err := repo.Update(ctx, id, invalid); err == nil {
		t.Fatal("expected invalid timestamp failure")
	}
	loaded, err = repo.GetByUUID(ctx, id)
	if err != nil || !reflect.DeepEqual(saved, loaded) {
		t.Fatalf("rollback lost data: %+v, %v", loaded, err)
	}
	changed := original
	changed.UUID = ""
	changed.Title = "Updated " + id
	if err := repo.Update(ctx, id, changed); err != nil {
		t.Fatal(err)
	}
	loaded, err = repo.GetByUUID(ctx, id)
	if err != nil || loaded.Title != changed.Title || len(loaded.Authors) != 2 || loaded.Metadata.Publisher != "Publisher" {
		t.Fatalf("update: %+v, %v", loaded, err)
	}
	// An author shared by another paper must survive deletion of the first paper.
	var authorKey int64
	if err := db.QueryRowContext(ctx, `SELECT key_author FROM public.paper_author WHERE uuid_paper=$1 ORDER BY position LIMIT 1`, id).Scan(&authorKey); err != nil {
		t.Fatal(err)
	}
	if _, err := db.ExecContext(ctx, `INSERT INTO public.paper_author (uuid_paper,key_author,position) VALUES ($1,$2,0)`, otherID, authorKey); err != nil {
		t.Fatal(err)
	}
	if err := repo.Delete(ctx, id); err != nil {
		t.Fatal(err)
	}
	if _, err := repo.GetByUUID(ctx, id); !errors.Is(err, domain.ErrPaperNotFound) {
		t.Fatalf("deleted paper: %v", err)
	}
	other, err := repo.GetByUUID(ctx, otherID)
	if err != nil || len(other.Authors) != 1 || other.Authors[0].NameFirst != "Ada" {
		t.Fatalf("shared author removed: %+v, %v", other, err)
	}
	if err := repo.Delete(ctx, otherID); err != nil {
		t.Fatal(err)
	}
	var remaining int
	if err := db.QueryRowContext(ctx, `SELECT count(*) FROM public.affiliation WHERE name=$1`, "Lab "+id).Scan(&remaining); err != nil {
		t.Fatal(err)
	}
	if remaining != 0 {
		t.Fatalf("orphan affiliations: %d", remaining)
	}
}
