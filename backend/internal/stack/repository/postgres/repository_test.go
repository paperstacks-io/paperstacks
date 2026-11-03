package postgres

import (
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	paperDomain "github.com/paperstacks.io/paperstacks/internal/paper/domain"
	"github.com/paperstacks.io/paperstacks/internal/stack/domain"
	userDomain "github.com/paperstacks.io/paperstacks/internal/user/domain"
)

func TestScanStacksClosesRows(t *testing.T) {
	scanErr := errors.New("scan failed")
	readErr := errors.New("read failed")
	for _, tc := range []struct {
		name string
		rows stackTestRows
		want error
	}{
		{name: "empty"},
		{name: "scan error", rows: stackTestRows{hasRow: true, scanErr: scanErr}, want: scanErr},
		{name: "read error", rows: stackTestRows{readErr: readErr}, want: readErr},
	} {
		t.Run(tc.name, func(t *testing.T) {
			items, err := scanStacks(&tc.rows)
			if !errors.Is(err, tc.want) {
				t.Fatalf("got error %v; want %v", err, tc.want)
			}
			if !tc.rows.closed {
				t.Fatal("rows must be closed before loading papers")
			}
			if tc.want == nil && (items == nil || len(items) != 0) {
				t.Fatalf("expected non-nil empty slice, got %v", items)
			}
		})
	}
}

type stackTestRows struct {
	pgx.Rows
	hasRow  bool
	scanErr error
	readErr error
	closed  bool
}

func (r *stackTestRows) Next() bool {
	hasRow := r.hasRow
	r.hasRow = false
	return hasRow
}

func (r *stackTestRows) Scan(...any) error { return r.scanErr }
func (r *stackTestRows) Err() error        { return r.readErr }
func (r *stackTestRows) Close()            { r.closed = true }

func TestStackDAOMapping(t *testing.T) {
	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	for _, orcid := range []pgtype.Text{{}, {String: "0000-0001", Valid: true}} {
		d := stackDAO{UUID: "stack-id", Name: "Research", OwnerExternalID: "owner", OwnerEmail: "owner@example.invalid",
			OwnerORCID: orcid, IsPublic: true, CreatedAt: now, UpdatedAt: now.Add(time.Hour), OwnerCreatedAt: now, OwnerUpdatedAt: now}
		want := domain.Stack{UUID: d.UUID, Name: d.Name, IsPublic: true, CreatedAt: now, UpdatedAt: now.Add(time.Hour),
			Papers: []paperDomain.Paper{}, Owner: userDomain.User{ExternalID: d.OwnerExternalID, Email: d.OwnerEmail,
				ORCID: orcid.String, CreatedAt: now, UpdatedAt: now}}
		if got := d.domain(); !reflect.DeepEqual(got, want) {
			t.Fatalf("got %+v; want %+v", got, want)
		}
	}
}
