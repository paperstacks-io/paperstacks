// Package postgres persists stacks in PostgreSQL.
package postgres

import (
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	paperDomain "github.com/paperstacks.io/paperstacks/internal/paper/domain"
	"github.com/paperstacks.io/paperstacks/internal/stack/domain"
	userDomain "github.com/paperstacks.io/paperstacks/internal/user/domain"
)

// stackDAO represents a stack row and its joined owner. SQL nullability and
// foreign-key fields belong here, not in the domain model.
type stackDAO struct {
	UUID            string
	Name            string
	OwnerExternalID string
	OwnerEmail      string
	OwnerORCID      pgtype.Text
	IsPublic        bool
	CreatedAt       time.Time
	UpdatedAt       time.Time
	OwnerCreatedAt  time.Time
	OwnerUpdatedAt  time.Time
}

type stackPaperDAO struct {
	UUID      string
	Title     pgtype.Text
	IsDeleted bool
}

const stackColumns = `s.uuid::text, s.name, s.owner_external_id, s.is_public,
 s.created_at, s.updated_at, u.email, u.orcid, u.created_at, u.updated_at`
const stackFrom = ` FROM public.stack s JOIN public.app_user u ON u.external_id = s.owner_external_id`

func scanStack(row pgx.Row) (domain.Stack, error) {
	var d stackDAO
	if err := row.Scan(&d.UUID, &d.Name, &d.OwnerExternalID, &d.IsPublic,
		&d.CreatedAt, &d.UpdatedAt, &d.OwnerEmail, &d.OwnerORCID, &d.OwnerCreatedAt, &d.OwnerUpdatedAt); err != nil {
		return domain.Stack{}, err
	}
	return d.domain(), nil
}

func (d stackDAO) domain() domain.Stack {
	return domain.Stack{
		UUID:      d.UUID,
		Name:      d.Name,
		IsPublic:  d.IsPublic,
		CreatedAt: d.CreatedAt,
		UpdatedAt: d.UpdatedAt,
		Papers:    []paperDomain.Paper{},
		Owner: userDomain.User{
			ExternalID: d.OwnerExternalID,
			Email:      d.OwnerEmail,
			ORCID:      d.OwnerORCID.String,
			CreatedAt:  d.OwnerCreatedAt,
			UpdatedAt:  d.OwnerUpdatedAt,
		},
	}
}
