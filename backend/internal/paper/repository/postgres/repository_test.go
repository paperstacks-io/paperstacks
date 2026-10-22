package postgres

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"math"
	"reflect"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/paperstacks.io/paperstacks/internal/paper/domain"
)

const testUUID = "67000000-0000-0000-0000-000000000001"

var paperFields = []string{"uuid", "doi", "title", "short", "year", "month", "day", "type", "status", "timestamp", "abstract", "keywords", "pdf"}
var metadataFields = []string{"uuid", "publisher", "journal", "abbrev", "book", "series", "event", "place", "institution", "pages", "volume", "issue", "isbn", "issn", "references", "license", "copyright", "funding", "source", "timestamp"}

func newTestRepository(t *testing.T) (*Repository, sqlmock.Sqlmock) {
	t.Helper()
	db, mock, err := sqlmock.New(sqlmock.ValueConverterOption(arrayConverter{}))
	if err != nil {
		t.Fatal(err)
	}
	db.SetMaxOpenConns(1)
	t.Cleanup(func() {
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Error(err)
		}
		db.Close()
	})
	return NewRepository(db), mock
}

// sqlmock uses database/sql's default converter, which cannot encode pgx arrays.
// Use pgx's own codec for these arguments rather than changing production SQL.
type arrayConverter struct{}

func (arrayConverter) ConvertValue(value any) (driver.Value, error) {
	var oid uint32
	switch value.(type) {
	case pgtype.FlatArray[string]:
		oid = pgtype.TextArrayOID
	case pgtype.FlatArray[int64]:
		oid = pgtype.Int8ArrayOID
	default:
		return driver.DefaultParameterConverter.ConvertValue(value)
	}
	encoded, err := pgtype.NewMap().Encode(oid, pgtype.TextFormatCode, value, nil)
	if err != nil {
		return nil, err
	}
	if encoded == nil {
		return nil, nil
	}
	return string(encoded), nil
}
func paperRows() *sqlmock.Rows {
	return sqlmock.NewRows(paperFields).AddRow(testUUID, "10.123/test", "Paper", nil, 2026, 10, 18, "journal-article", nil, nil, nil, `{"science","data"}`, nil)
}
func expectChildren(mock sqlmock.Sqlmock) {
	mock.ExpectQuery(`SELECT .* FROM public.metadata`).WillReturnRows(sqlmock.NewRows(metadataFields).AddRow(testUUID, "Publisher", nil, nil, nil, nil, nil, nil, nil, "1-9", "S1", nil, `{"isbn"}`, nil, `{"reference"}`, nil, nil, nil, nil, nil))
	mock.ExpectQuery(`SELECT pa.uuid_paper`).WillReturnRows(sqlmock.NewRows([]string{"uuid", "first", "middle", "last", "affiliation", "orcid"}).AddRow(testUUID, "Ada", nil, "Lovelace", "Lab", nil).AddRow(testUUID, "Grace", nil, "Hopper", nil, "0000"))
}
func TestReadOperations(t *testing.T) {
	for _, method := range []string{"uuid", "doi", "list", "search"} {
		t.Run(method, func(t *testing.T) {
			repo, mock := newTestRepository(t)
			if method == "search" {
				mock.ExpectQuery(`SELECT count`).WithArgs("paper").WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(2))
			}
			query := mock.ExpectQuery(`SELECT .* FROM public.paper`)
			switch method {
			case "uuid":
				query.WithArgs(testUUID)
			case "doi":
				query.WithArgs("10.123/test")
			case "search":
				query.WithArgs("paper", 1, 0)
			}
			query.WillReturnRows(paperRows())
			expectChildren(mock)
			var got domain.Paper
			var err error
			switch method {
			case "uuid":
				got, err = repo.GetByUUID(context.Background(), testUUID)
			case "doi":
				got, err = repo.GetByDOI(context.Background(), "10.123/test")
			case "list":
				var items []domain.Paper
				items, err = repo.List(context.Background())
				if len(items) == 1 {
					got = items[0]
				}
			case "search":
				var result domain.SearchResult
				result, err = repo.Search(context.Background(), domain.SearchOptions{Query: "Paper", SortBy: "year", Desc: true, PageSize: 1})
				if len(result.Items) == 1 {
					got = result.Items[0]
				}
				if result.Total != 2 || !result.HasNext {
					t.Fatalf("bad pagination: %+v", result)
				}
			}
			if err != nil {
				t.Fatal(err)
			}
			if got.UUID != testUUID || got.PublicationDate != (domain.Date{Year: 2026, Month: 10, Day: 18}) || len(got.Keywords) != 2 || got.Metadata.Publisher != "Publisher" || got.Metadata.Volume != "S1" || len(got.Authors) != 2 || got.Authors[0].NameFirst != "Ada" || got.Authors[1].NameFirst != "Grace" {
				t.Fatalf("bad mapping: %+v", got)
			}
		})
	}
}
func TestReadErrors(t *testing.T) {
	cause := errors.New("connection lost")
	for _, tc := range []struct {
		name      string
		err, want error
	}{{"missing", sql.ErrNoRows, domain.ErrPaperNotFound}, {"query", cause, cause}, {"cancelled", context.Canceled, context.Canceled}} {
		t.Run(tc.name, func(t *testing.T) {
			repo, mock := newTestRepository(t)
			mock.ExpectQuery(`SELECT`).WillReturnError(tc.err)
			_, err := repo.GetByUUID(context.Background(), testUUID)
			if !errors.Is(err, tc.want) {
				t.Fatalf("got %v", err)
			}
		})
	}
	for _, stage := range []string{"scan", "iteration", "metadata", "authors"} {
		t.Run(stage, func(t *testing.T) {
			repo, mock := newTestRepository(t)
			rows := paperRows()
			if stage == "scan" {
				rows = sqlmock.NewRows([]string{"uuid"}).AddRow(testUUID)
			}
			if stage == "iteration" {
				rows.RowError(0, cause)
			}
			mock.ExpectQuery(`SELECT .* FROM public.paper`).WillReturnRows(rows)
			if stage == "metadata" {
				mock.ExpectQuery(`SELECT .* FROM public.metadata`).WillReturnError(cause)
			}
			if stage == "authors" {
				mock.ExpectQuery(`SELECT .* FROM public.metadata`).WillReturnRows(sqlmock.NewRows(metadataFields))
				mock.ExpectQuery(`SELECT pa.uuid_paper`).WillReturnError(cause)
			}
			if _, err := repo.List(context.Background()); err == nil {
				t.Fatal("expected failure")
			}
		})
	}
}
func TestSearchValidationAndBounds(t *testing.T) {
	repo, mock := newTestRepository(t)
	if _, err := repo.Search(context.Background(), domain.SearchOptions{SortBy: "title; DROP TABLE paper"}); !errors.Is(err, domain.ErrInvalidSearch) {
		t.Fatalf("got %v", err)
	}
	mock.ExpectQuery(`SELECT count`).WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
	mock.ExpectQuery(`ORDER BY title ASC, doi ASC, uuid ASC LIMIT \$2 OFFSET \$3`).WithArgs("", math.MaxInt, 1).WillReturnRows(sqlmock.NewRows(paperFields))
	got, err := repo.Search(context.Background(), domain.SearchOptions{SortBy: "title", Page: math.MaxInt, PageSize: math.MaxInt})
	if err != nil || got.HasNext || len(got.Items) != 0 || got.Items == nil {
		t.Fatalf("got %+v, %v", got, err)
	}
}
func expectRemoveAuthors(mock sqlmock.Sqlmock) {
	mock.ExpectQuery(`WITH deleted_links`).WithArgs(testUUID).WillReturnRows(sqlmock.NewRows([]string{"keys"}).AddRow(`{1}`))
	mock.ExpectQuery(`WITH deleted_authors`).WillReturnRows(sqlmock.NewRows([]string{"keys"}).AddRow(`{2}`))
	mock.ExpectExec(`DELETE FROM public.affiliation`).WillReturnResult(sqlmock.NewResult(0, 1))
}
func expectInsertChildren(mock sqlmock.Sqlmock) {
	mock.ExpectExec(`INSERT INTO public.metadata`).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectQuery(`INSERT INTO public.affiliation`).WithArgs("Lab").WillReturnRows(sqlmock.NewRows([]string{"key"}).AddRow(2))
	mock.ExpectQuery(`INSERT INTO public.author`).WithArgs("Ada", nil, "Lovelace", nil, int64(2)).WillReturnRows(sqlmock.NewRows([]string{"key"}).AddRow(1))
	mock.ExpectExec(`INSERT INTO public.paper_author`).WithArgs(testUUID, int64(1), 0).WillReturnResult(sqlmock.NewResult(0, 1))
}
func samplePaper() domain.Paper {
	return domain.Paper{UUID: testUUID, Title: " Paper ", Authors: []domain.Author{{NameFirst: "Ada", NameLast: "Lovelace", Affiliation: "Lab"}}}
}
func TestSave(t *testing.T) {
	repo, mock := newTestRepository(t)
	mock.ExpectBegin()
	mock.ExpectExec(`INSERT INTO public.paper \(`).WithArgs(testUUID, nil, "Paper", nil, nil, nil, nil, nil, nil, nil, nil, sqlmock.AnyArg(), nil).WillReturnResult(sqlmock.NewResult(0, 1))
	expectInsertChildren(mock)
	mock.ExpectCommit()
	got, err := repo.Save(context.Background(), samplePaper())
	if err != nil || got.Title != "Paper" {
		t.Fatalf("got %+v, %v", got, err)
	}
}
func TestUpdate(t *testing.T) {
	repo, mock := newTestRepository(t)
	p := samplePaper()
	p.UUID = ""
	mock.ExpectBegin()
	mock.ExpectExec(`UPDATE public.paper`).WithArgs(testUUID, nil, "Paper", nil, nil, nil, nil, nil, nil, nil, nil, sqlmock.AnyArg(), nil).WillReturnResult(sqlmock.NewResult(0, 1))
	expectRemoveAuthors(mock)
	mock.ExpectExec(`DELETE FROM public.metadata`).WithArgs(testUUID).WillReturnResult(sqlmock.NewResult(0, 1))
	expectInsertChildren(mock)
	mock.ExpectCommit()
	if err := repo.Update(context.Background(), testUUID, p); err != nil {
		t.Fatal(err)
	}
}
func TestUpdateUUIDMismatch(t *testing.T) {
	repo, _ := newTestRepository(t)
	if err := repo.Update(context.Background(), "other", samplePaper()); !errors.Is(err, domain.ErrUUIDMismatch) {
		t.Fatalf("got %v", err)
	}
}
func TestUpdateMissing(t *testing.T) {
	repo, mock := newTestRepository(t)
	mock.ExpectBegin()
	mock.ExpectExec(`UPDATE public.paper`).WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectRollback()
	if err := repo.Update(context.Background(), testUUID, samplePaper()); !errors.Is(err, domain.ErrPaperNotFound) {
		t.Fatalf("got %v", err)
	}
}
func TestDelete(t *testing.T) {
	repo, mock := newTestRepository(t)
	mock.ExpectBegin()
	mock.ExpectQuery(`SELECT uuid::text FROM public.paper WHERE uuid = \$1 FOR UPDATE`).WithArgs(testUUID).WillReturnRows(sqlmock.NewRows([]string{"uuid"}).AddRow(testUUID))
	expectRemoveAuthors(mock)
	mock.ExpectExec(`DELETE FROM public.paper`).WithArgs(testUUID).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()
	if err := repo.Delete(context.Background(), testUUID); err != nil {
		t.Fatal(err)
	}
}
func TestDeleteMissing(t *testing.T) {
	repo, mock := newTestRepository(t)
	mock.ExpectBegin()
	mock.ExpectQuery(`SELECT uuid::text`).WillReturnError(sql.ErrNoRows)
	mock.ExpectRollback()
	if err := repo.Delete(context.Background(), testUUID); !errors.Is(err, domain.ErrPaperNotFound) {
		t.Fatalf("got %v", err)
	}
}
func TestSaveFailures(t *testing.T) {
	for _, stage := range []string{"begin", "paper", "metadata", "author", "commit"} {
		t.Run(stage, func(t *testing.T) {
			repo, mock := newTestRepository(t)
			cause := errors.New("database unavailable")
			begin := mock.ExpectBegin()
			if stage == "begin" {
				begin.WillReturnError(cause)
			} else {
				insert := mock.ExpectExec(`INSERT INTO public.paper \(`)
				if stage == "paper" {
					insert.WillReturnError(cause)
				} else {
					insert.WillReturnResult(sqlmock.NewResult(0, 1))
					metadata := mock.ExpectExec(`INSERT INTO public.metadata`)
					if stage == "metadata" {
						metadata.WillReturnError(cause)
					} else {
						metadata.WillReturnResult(sqlmock.NewResult(0, 1))
						affiliation := mock.ExpectQuery(`INSERT INTO public.affiliation`)
						if stage == "author" {
							affiliation.WillReturnError(cause)
						} else {
							affiliation.WillReturnRows(sqlmock.NewRows([]string{"key"}).AddRow(2))
							mock.ExpectQuery(`INSERT INTO public.author`).WillReturnRows(sqlmock.NewRows([]string{"key"}).AddRow(1))
							mock.ExpectExec(`INSERT INTO public.paper_author`).WillReturnResult(sqlmock.NewResult(0, 1))
						}
					}
				}
				if stage == "commit" {
					mock.ExpectCommit().WillReturnError(cause)
				} else {
					mock.ExpectRollback()
				}
			}
			_, err := repo.Save(context.Background(), samplePaper())
			if !errors.Is(err, cause) {
				t.Fatalf("got %v", err)
			}
		})
	}
}
func TestConstraintErrors(t *testing.T) {
	for _, tc := range []struct {
		code string
		want error
	}{{"23505", domain.ErrPaperAlreadyExists}, {"23503", domain.ErrInvalidPaper}, {"23514", domain.ErrInvalidPaper}, {"22P02", domain.ErrInvalidPaper}, {"22003", domain.ErrInvalidPaper}} {
		cause := &pgconn.PgError{Code: tc.code}
		err := wrapError("save", cause)
		if !errors.Is(err, cause) || !errors.Is(err, tc.want) {
			t.Fatalf("got %v", err)
		}
	}
}
func TestNullableDAOs(t *testing.T) {
	if got := (paperDAO{}).domain(); !reflect.DeepEqual(got, domain.Paper{}) {
		t.Fatalf("got %+v", got)
	}
	if got := (metadataDAO{}).domain(); !reflect.DeepEqual(got, domain.Metadata{}) {
		t.Fatalf("got %+v", got)
	}
	if got := (authorDAO{}).domain(); !reflect.DeepEqual(got, domain.Author{}) {
		t.Fatalf("got %+v", got)
	}
	stamp := time.Date(2026, 10, 18, 12, 0, 0, 123000000, time.FixedZone("offset", 7200))
	got := (paperDAO{StatusTimestamp: sql.NullTime{Time: stamp, Valid: true}}).domain()
	if got.PublicationStatusTimestamp != "2026-10-18T10:00:00.123Z" {
		t.Fatal(got.PublicationStatusTimestamp)
	}
}
