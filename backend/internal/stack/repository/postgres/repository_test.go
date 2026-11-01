package postgres

import (
	"reflect"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	paperDomain "github.com/paperstacks.io/paperstacks/internal/paper/domain"
	"github.com/paperstacks.io/paperstacks/internal/stack/domain"
	userDomain "github.com/paperstacks.io/paperstacks/internal/user/domain"
)

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
