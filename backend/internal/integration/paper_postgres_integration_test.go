package integration

import (
	"context"
	"errors"
	"os"
	"reflect"
	"sync"
	"testing"
	"time"
	"uuid"

	"github.com/jackc/pgx/v5/pgxpool"
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
	poolConfig, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		t.Fatal(err)
	}
	poolConfig.MaxConns = 1
	poolConfig.MinConns = 0
	poolConfig.MinIdleConns = 0
	db, err := pgxpool.NewWithConfig(context.Background(), poolConfig)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := db.Ping(ctx); err != nil {
		t.Fatal(err)
	}
	repo := postgres.NewRepository(db)
	id, otherID, duplicateID := uuid.New().String(), uuid.New().String(), uuid.New().String()
	t.Cleanup(func() {
		cleanup, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		// Hard-delete only this test's fixtures; repository Delete intentionally retains them.
		tx, err := db.Begin(cleanup)
		if err != nil {
			t.Error(err)
			return
		}
		defer tx.Rollback(cleanup)
		ids := []string{id, otherID, duplicateID}
		var authorKeys []int64
		if err := tx.QueryRow(cleanup, `SELECT array_agg(key_author) FROM public.paper_author WHERE uuid_paper = ANY($1::uuid[])`, ids).Scan(&authorKeys); err != nil {
			t.Error(err)
			return
		}
		if _, err := tx.Exec(cleanup, `DELETE FROM public.paper_author WHERE uuid_paper = ANY($1::uuid[])`, ids); err != nil {
			t.Error(err)
			return
		}
		if _, err := tx.Exec(cleanup, `DELETE FROM public.paper WHERE uuid = ANY($1::uuid[])`, ids); err != nil {
			t.Error(err)
			return
		}
		var affiliationKeys []int64
		if err := tx.QueryRow(cleanup, `
			WITH removed AS (
				DELETE FROM public.author a WHERE key = ANY($1::bigint[])
				AND NOT EXISTS (SELECT 1 FROM public.paper_author WHERE key_author = a.key)
				RETURNING key_affiliation
			) SELECT array_agg(key_affiliation) FILTER (WHERE key_affiliation IS NOT NULL) FROM removed`, authorKeys).Scan(&affiliationKeys); err != nil {
			t.Error(err)
			return
		}
		if _, err := tx.Exec(cleanup, `DELETE FROM public.affiliation a WHERE key = ANY($1::bigint[])
			AND NOT EXISTS (SELECT 1 FROM public.author WHERE key_affiliation = a.key)`, affiliationKeys); err != nil {
			t.Error(err)
			return
		}
		if err := tx.Commit(cleanup); err != nil {
			t.Error(err)
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
	t.Run("concurrent reads share the pool", func(t *testing.T) {
		var workers sync.WaitGroup
		for range 8 {
			workers.Add(1)
			go func() {
				defer workers.Done()
				got, err := repo.GetByUUID(ctx, id)
				if err != nil {
					t.Error(err)
					return
				}
				if !reflect.DeepEqual(saved, got) {
					t.Errorf("concurrent read: got %+v; want %+v", got, saved)
				}
			}()
		}
		workers.Wait()
	})
	duplicateUUID := original
	duplicateUUID.DOI = "10.6700/duplicate-uuid-" + id
	if _, err := repo.Save(ctx, duplicateUUID); !errors.Is(err, domain.ErrPaperAlreadyExists) {
		t.Fatalf("duplicate UUID: %v", err)
	}
	duplicate := original
	duplicate.UUID = duplicateID
	if _, err := repo.Save(ctx, duplicate); !errors.Is(err, domain.ErrPaperAlreadyExists) {
		t.Fatalf("duplicate DOI: %v", err)
	}
	if _, err := repo.Save(ctx, domain.Paper{UUID: otherID, Title: "Other " + id}); err != nil {
		t.Fatal(err)
	}
	conflicting := original
	conflicting.UUID = otherID
	if err := repo.Update(ctx, otherID, conflicting); !errors.Is(err, domain.ErrPaperAlreadyExists) {
		t.Fatalf("update with duplicate DOI: %v", err)
	}
	unchanged, err := repo.GetByUUID(ctx, otherID)
	if err != nil || unchanged.DOI != "" || unchanged.Title != "Other "+id {
		t.Fatalf("conflicting update changed paper: %+v, %v", unchanged, err)
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
	if err := db.QueryRow(ctx, `SELECT key_author FROM public.paper_author WHERE uuid_paper=$1 ORDER BY position LIMIT 1`, id).Scan(&authorKey); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(ctx, `INSERT INTO public.paper_author (uuid_paper,key_author,position) VALUES ($1,$2,0)`, otherID, authorKey); err != nil {
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
	var deletedAt, repeatedDeletedAt time.Time
	if err := db.QueryRow(ctx, `SELECT deleted_at FROM public.paper WHERE uuid=$1`, id).Scan(&deletedAt); err != nil {
		t.Fatal(err)
	}
	if err := repo.Delete(ctx, id); !errors.Is(err, domain.ErrPaperNotFound) {
		t.Fatalf("repeated delete: %v", err)
	}
	if err := db.QueryRow(ctx, `SELECT deleted_at FROM public.paper WHERE uuid=$1`, id).Scan(&repeatedDeletedAt); err != nil || !deletedAt.Equal(repeatedDeletedAt) {
		t.Fatalf("repeated delete changed timestamp: %v", err)
	}
	if _, err := repo.GetByDOI(ctx, original.DOI); !errors.Is(err, domain.ErrPaperNotFound) {
		t.Fatalf("deleted DOI lookup: %v", err)
	}
	if err := repo.Update(ctx, id, original); !errors.Is(err, domain.ErrPaperNotFound) {
		t.Fatalf("update deleted paper: %v", err)
	}
	items, err = repo.List(ctx)
	if err != nil {
		t.Fatal(err)
	}
	for _, p := range items {
		if p.UUID == id || p.UUID == otherID {
			t.Fatal("list includes deleted paper")
		}
	}
	for _, query := range []string{id, ""} {
		result, err := repo.Search(ctx, domain.SearchOptions{Query: query})
		if err != nil {
			t.Fatal(err)
		}
		if query == id && (result.Total != 0 || len(result.Items) != 0 || result.HasNext) {
			t.Fatalf("search counts deleted papers: %+v", result)
		}
		for _, p := range result.Items {
			if p.UUID == id || p.UUID == otherID {
				t.Fatal("search includes deleted paper")
			}
		}
	}
	if _, err := repo.Save(ctx, duplicate); !errors.Is(err, domain.ErrPaperAlreadyExists) {
		t.Fatalf("deleted paper DOI should remain reserved: %v", err)
	}
	var metadataCount, authorCount, affiliationCount int
	if err := db.QueryRow(ctx, `SELECT count(*) FROM public.metadata WHERE uuid_paper=$1`, id).Scan(&metadataCount); err != nil {
		t.Fatal(err)
	}
	if err := db.QueryRow(ctx, `SELECT count(*) FROM public.paper_author WHERE uuid_paper=$1`, id).Scan(&authorCount); err != nil {
		t.Fatal(err)
	}
	if err := db.QueryRow(ctx, `SELECT count(*) FROM public.affiliation WHERE name=$1`, "Lab "+id).Scan(&affiliationCount); err != nil {
		t.Fatal(err)
	}
	if metadataCount != 1 || authorCount != 2 || affiliationCount != 1 {
		t.Fatalf("soft delete lost related data: metadata=%d authors=%d affiliations=%d", metadataCount, authorCount, affiliationCount)
	}
}
