package postgres

import (
	"reflect"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/paperstacks.io/paperstacks/internal/paper/domain"
)

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
	got := (paperDAO{StatusTimestamp: pgtype.Timestamptz{Time: stamp, Valid: true}}).domain()
	if got.PublicationStatusTimestamp != "2026-10-18T10:00:00.123Z" {
		t.Fatal(got.PublicationStatusTimestamp)
	}
}
