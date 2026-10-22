package postgres

import (
	"context"
	"database/sql"
	"errors"
	"math"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jackc/pgx/v5/pgconn"
	paperDomain "github.com/paperstacks.io/paperstacks/internal/paper/domain"
	"github.com/paperstacks.io/paperstacks/internal/stack/domain"
	userDomain "github.com/paperstacks.io/paperstacks/internal/user/domain"
)

const stackID = "68000000-0000-0000-0000-000000000001"
const paperID = "68000000-0000-0000-0000-000000000002"

var testTime = time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
var stackFields = []string{"uuid", "name", "owner_external_id", "is_public", "created_at", "updated_at", "email", "orcid", "owner_created_at", "owner_updated_at"}

func newTestRepository(t *testing.T) (*Repository, sqlmock.Sqlmock) {
	t.Helper()
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	// Also guards against loading child rows while a parent result holds the pool.
	db.SetMaxOpenConns(1)
	t.Cleanup(func() {
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Error(err)
		}
		db.Close()
	})
	return NewRepository(db), mock
}
func stackRows() *sqlmock.Rows {
	return sqlmock.NewRows(stackFields).AddRow(stackID, "Research", "owner", true, testTime, testTime, "owner@example.test", nil, testTime, testTime)
}
func testStack() domain.Stack {
	return domain.Stack{UUID: stackID, Name: "Research", Owner: userDomain.User{ExternalID: "owner"}, CreatedAt: testTime, UpdatedAt: testTime}
}

type stubPapers struct {
	paper paperDomain.Paper
	err   error
}

func (s stubPapers) GetByUUID(context.Context, string) (paperDomain.Paper, error) {
	return s.paper, s.err
}

func TestGetByUUID(t *testing.T) {
	repo, mock := newTestRepository(t)
	repo.papers = stubPapers{paper: paperDomain.Paper{UUID: paperID, Title: "Paper"}}
	mock.ExpectQuery(`SELECT .* FROM public.stack s .* WHERE s.uuid = \$1`).WithArgs(stackID).WillReturnRows(stackRows())
	mock.ExpectQuery(`SELECT uuid_paper::text`).WithArgs(stackID).WillReturnRows(sqlmock.NewRows([]string{"uuid"}).AddRow(paperID))
	got, err := repo.GetByUUID(context.Background(), stackID)
	if err != nil {
		t.Fatal(err)
	}
	if got.UUID != stackID || got.Owner.ExternalID != "owner" || got.Owner.Email != "owner@example.test" || got.Owner.ORCID != "" || !got.CreatedAt.Equal(testTime) || len(got.Papers) != 1 || got.Papers[0].Title != "Paper" {
		t.Fatalf("unexpected stack: %+v", got)
	}
}
func TestGetByUUIDFailures(t *testing.T) {
	failure := errors.New("connection lost")
	for _, tc := range []struct {
		name string
		err  error
		want error
	}{
		{"missing", sql.ErrNoRows, domain.ErrStackNotFound}, {"query failure", failure, failure}, {"cancelled", context.Canceled, context.Canceled},
	} {
		t.Run(tc.name, func(t *testing.T) {
			repo, mock := newTestRepository(t)
			mock.ExpectQuery(`SELECT`).WillReturnError(tc.err)
			_, err := repo.GetByUUID(context.Background(), stackID)
			if !errors.Is(err, tc.want) {
				t.Fatalf("got %v, want %v", err, tc.want)
			}
		})
	}
}
func TestListFilters(t *testing.T) {
	for _, public := range []bool{false, true} {
		name := "private and public"
		if public {
			name = "public only"
		}
		t.Run(name, func(t *testing.T) {
			repo, mock := newTestRepository(t)
			where := ` WHERE s.owner_external_id = $1`
			if public {
				where += ` AND s.is_public`
			}
			mock.ExpectQuery(regexp.QuoteMeta(where + ` ORDER BY s.uuid`)).WithArgs("owner").WillReturnRows(stackRows())
			mock.ExpectQuery(`SELECT uuid_paper::text`).WithArgs(stackID).WillReturnRows(sqlmock.NewRows([]string{"uuid"}))
			var got []domain.Stack
			var err error
			if public {
				got, err = repo.ListPublic(context.Background(), "owner")
			} else {
				got, err = repo.List(context.Background(), "owner")
			}
			if err != nil || len(got) != 1 || got[0].Papers == nil {
				t.Fatalf("got %+v, %v", got, err)
			}
		})
	}
}
func TestListRowFailures(t *testing.T) {
	for _, scan := range []bool{false, true} {
		t.Run(map[bool]string{true: "scan", false: "iteration"}[scan], func(t *testing.T) {
			repo, mock := newTestRepository(t)
			rows := stackRows()
			if scan {
				rows = sqlmock.NewRows([]string{"uuid"}).AddRow(stackID)
			} else {
				rows.RowError(0, errors.New("read failed"))
			}
			mock.ExpectQuery(`SELECT`).WillReturnRows(rows)
			if _, err := repo.List(context.Background(), "owner"); err == nil {
				t.Fatal("expected failure")
			}
		})
	}
}
func TestPaperLoadingFailure(t *testing.T) {
	repo, mock := newTestRepository(t)
	failure := errors.New("paper query failed")
	repo.papers = stubPapers{err: failure}
	mock.ExpectQuery(`SELECT .* FROM public.stack s`).WillReturnRows(stackRows())
	mock.ExpectQuery(`SELECT uuid_paper::text`).WillReturnRows(sqlmock.NewRows([]string{"uuid"}).AddRow(paperID))
	if _, err := repo.GetByUUID(context.Background(), stackID); !errors.Is(err, failure) {
		t.Fatalf("got %v", err)
	}
}
func TestSearch(t *testing.T) {
	repo, mock := newTestRepository(t)
	mock.ExpectQuery(regexp.QuoteMeta(`WHERE s.owner_external_id = $1 AND strpos(lower(s.name), $2) > 0`)).WithArgs("owner", "research").WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(3))
	mock.ExpectQuery(regexp.QuoteMeta(`ORDER BY lower(s.name) DESC, s.uuid ASC LIMIT $3 OFFSET $4`)).WithArgs("owner", "research", 1, 1).WillReturnRows(stackRows())
	mock.ExpectQuery(`SELECT uuid_paper::text`).WillReturnRows(sqlmock.NewRows([]string{"uuid"}))
	got, err := repo.SearchByOwner(context.Background(), "owner", domain.SearchOptions{Query: " Research ", SortBy: "name", Desc: true, Page: 2, PageSize: 1})
	if err != nil {
		t.Fatal(err)
	}
	if got.Total != 3 || got.Page != 2 || got.PageSize != 1 || !got.HasNext || len(got.Items) != 1 {
		t.Fatalf("unexpected result: %+v", got)
	}
}
func TestSearchPublicEmptyAndOutOfRange(t *testing.T) {
	for _, page := range []int{0, math.MaxInt} {
		repo, mock := newTestRepository(t)
		mock.ExpectQuery(regexp.QuoteMeta(`WHERE s.is_public AND strpos(lower(s.name), $1) > 0`)).WithArgs("").WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
		mock.ExpectQuery(`ORDER BY s.updated_at, s.uuid ASC LIMIT \$2 OFFSET \$3`).WithArgs("", 1, 0).WillReturnRows(sqlmock.NewRows(stackFields))
		got, err := repo.SearchPublic(context.Background(), domain.SearchOptions{SortBy: "updated_at", Page: page})
		if err != nil || got.Total != 0 || got.PageSize != 1 || got.HasNext || got.Items == nil {
			t.Fatalf("got %+v, %v", got, err)
		}
	}
}
func TestSearchRejectsUnknownSort(t *testing.T) {
	repo, _ := newTestRepository(t)
	_, err := repo.SearchPublic(context.Background(), domain.SearchOptions{SortBy: "name; DROP TABLE stack"})
	if !errors.Is(err, domain.ErrInvalidSearch) {
		t.Fatalf("got %v", err)
	}
}
func TestCounts(t *testing.T) {
	repo, mock := newTestRepository(t)
	mock.ExpectQuery(`SELECT count\(\*\) FROM public.stack WHERE is_public`).WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(2))
	count, err := repo.CountPublic(context.Background())
	if err != nil || count != 2 {
		t.Fatalf("got %d, %v", count, err)
	}
	mock.ExpectQuery(`SELECT count\(\*\), count\(\*\) FILTER`).WithArgs("owner").WillReturnRows(sqlmock.NewRows([]string{"total", "public", "papers"}).AddRow(3, 2, 5))
	stats, err := repo.StatsByOwner(context.Background(), "owner")
	if err != nil || stats != (domain.Stats{TotalStacks: 3, PublicStacks: 2, TotalPapers: 5}) {
		t.Fatalf("got %+v, %v", stats, err)
	}
}
func TestCreateTransaction(t *testing.T) {
	repo, mock := newTestRepository(t)
	s := testStack()
	s.Papers = []paperDomain.Paper{{UUID: paperID}}
	mock.ExpectBegin()
	mock.ExpectExec(`INSERT INTO public.stack \(`).WithArgs(s.UUID, s.Name, "owner", false, testTime, testTime).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(`DELETE FROM public.stack_paper`).WithArgs(stackID).WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec(`INSERT INTO public.stack_paper`).WithArgs(stackID, paperID).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()
	if err := repo.Create(context.Background(), s); err != nil {
		t.Fatal(err)
	}
}
func TestCreateRollsBackOnChildFailure(t *testing.T) {
	repo, mock := newTestRepository(t)
	s := testStack()
	s.Papers = []paperDomain.Paper{{UUID: paperID}}
	cause := &pgconn.PgError{Code: "23503"}
	mock.ExpectBegin()
	mock.ExpectExec(`INSERT INTO public.stack \(`).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(`DELETE FROM public.stack_paper`).WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec(`INSERT INTO public.stack_paper`).WillReturnError(cause)
	mock.ExpectRollback()
	err := repo.Create(context.Background(), s)
	if !errors.Is(err, domain.ErrInvalidStack) || !errors.Is(err, cause) {
		t.Fatalf("got %v", err)
	}
}
func TestTransactionFailures(t *testing.T) {
	for _, stage := range []string{"begin", "insert", "commit"} {
		t.Run(stage, func(t *testing.T) {
			repo, mock := newTestRepository(t)
			cause := errors.New("connection failure")
			begin := mock.ExpectBegin()
			if stage == "begin" {
				begin.WillReturnError(cause)
			} else {
				insert := mock.ExpectExec(`INSERT INTO public.stack \(`)
				if stage == "insert" {
					insert.WillReturnError(cause)
					mock.ExpectRollback()
				} else {
					insert.WillReturnResult(sqlmock.NewResult(0, 1))
					mock.ExpectExec(`DELETE FROM public.stack_paper`).WillReturnResult(sqlmock.NewResult(0, 0))
					mock.ExpectCommit().WillReturnError(cause)
				}
			}
			if err := repo.Create(context.Background(), testStack()); !errors.Is(err, cause) {
				t.Fatalf("got %v", err)
			}
		})
	}
}
func TestUpdatePreservesCreationTime(t *testing.T) {
	repo, mock := newTestRepository(t)
	s := testStack()
	s.CreatedAt = testTime.Add(time.Hour)
	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(`WHERE uuid=$1 AND owner_external_id=$5 RETURNING created_at, updated_at`)).WithArgs(stackID, "Research", false, testTime, "owner").WillReturnRows(sqlmock.NewRows([]string{"created", "updated"}).AddRow(testTime, testTime))
	mock.ExpectExec(`DELETE FROM public.stack_paper`).WithArgs(stackID).WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectCommit()
	got, err := repo.Update(context.Background(), s)
	if err != nil || !got.CreatedAt.Equal(testTime) {
		t.Fatalf("got %+v, %v", got, err)
	}
}
func TestUpdateMissingOrWrongOwner(t *testing.T) {
	repo, mock := newTestRepository(t)
	mock.ExpectBegin()
	mock.ExpectQuery(`UPDATE public.stack .*owner_external_id=\$5`).WillReturnError(sql.ErrNoRows)
	mock.ExpectRollback()
	if _, err := repo.Update(context.Background(), testStack()); !errors.Is(err, domain.ErrStackNotFound) {
		t.Fatalf("got %v", err)
	}
}
func TestDelete(t *testing.T) {
	for _, n := range []int64{0, 1} {
		repo, mock := newTestRepository(t)
		mock.ExpectExec(`DELETE FROM public.stack WHERE uuid=\$1`).WithArgs(stackID).WillReturnResult(sqlmock.NewResult(0, n))
		err := repo.Delete(context.Background(), stackID)
		if n == 0 && !errors.Is(err, domain.ErrStackNotFound) || n == 1 && err != nil {
			t.Fatalf("got %v", err)
		}
	}
}
func TestPaperMembership(t *testing.T) {
	for _, add := range []bool{true, false} {
		for _, changed := range []int64{0, 1} {
			repo, mock := newTestRepository(t)
			mock.ExpectBegin()
			mock.ExpectQuery(`SELECT uuid::text FROM public.stack WHERE uuid=\$1 FOR UPDATE`).WithArgs(stackID).WillReturnRows(sqlmock.NewRows([]string{"uuid"}).AddRow(stackID))
			query := `DELETE FROM public.stack_paper`
			if add {
				query = `INSERT INTO public.stack_paper`
			}
			mock.ExpectExec(query).WithArgs(stackID, paperID).WillReturnResult(sqlmock.NewResult(0, changed))
			if changed > 0 {
				mock.ExpectExec(`UPDATE public.stack SET updated_at=statement_timestamp\(\)`).WithArgs(stackID).WillReturnResult(sqlmock.NewResult(0, 1))
			}
			mock.ExpectCommit()
			var err error
			if add {
				err = repo.AddPaper(context.Background(), stackID, paperDomain.Paper{UUID: paperID})
			} else {
				err = repo.RemovePaper(context.Background(), stackID, paperID)
			}
			if err != nil {
				t.Fatal(err)
			}
		}
	}
}
func TestAddPaperMissingStackOrPaper(t *testing.T) {
	for _, missingStack := range []bool{true, false} {
		repo, mock := newTestRepository(t)
		mock.ExpectBegin()
		lock := mock.ExpectQuery(`SELECT uuid::text`)
		want := paperDomain.ErrPaperNotFound
		if missingStack {
			lock.WillReturnError(sql.ErrNoRows)
			want = domain.ErrStackNotFound
		} else {
			lock.WillReturnRows(sqlmock.NewRows([]string{"uuid"}).AddRow(stackID))
			mock.ExpectExec(`INSERT INTO public.stack_paper`).WillReturnError(&pgconn.PgError{Code: "23503"})
		}
		mock.ExpectRollback()
		if err := repo.AddPaper(context.Background(), stackID, paperDomain.Paper{UUID: paperID}); !errors.Is(err, want) {
			t.Fatalf("got %v, want %v", err, want)
		}
	}
}
func TestWrapError(t *testing.T) {
	for _, tc := range []struct {
		code string
		want error
	}{
		{"23505", domain.ErrStackAlreadyExists}, {"23503", domain.ErrInvalidStack}, {"22P02", domain.ErrInvalidStack},
	} {
		cause := &pgconn.PgError{Code: tc.code}
		err := wrapError("operation", cause)
		if !errors.Is(err, tc.want) || !errors.Is(err, cause) {
			t.Fatalf("got %v", err)
		}
	}
}
