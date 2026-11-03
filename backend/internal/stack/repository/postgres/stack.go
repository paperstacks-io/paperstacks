// Package postgres persists stacks in PostgreSQL.
package postgres

import (
	"fmt"
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

func scanStacks(rows pgx.Rows) ([]domain.Stack, error) {
	defer rows.Close()
	items := make([]domain.Stack, 0)
	for rows.Next() {
		s, err := scanStack(rows)
		if err != nil {
			return nil, fmt.Errorf("scan stack: %w", err)
		}
		items = append(items, s)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("read stacks: %w", err)
	}
	return items, nil
}

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
