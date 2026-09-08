// Package postgres persists users in PostgreSQL.
package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/paperstacks.io/paperstacks/internal/user/domain"
)

const userColumns = `external_id, email, orcid, created_at, updated_at`

// Repository persists users in PostgreSQL through database/sql.
type Repository struct {
	db *sql.DB
}

var _ domain.Repository = (*Repository)(nil)

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) GetByExternalID(ctx context.Context, externalID string) (domain.User, error) {
	user, err := scanUser(r.db.QueryRowContext(ctx, `SELECT `+userColumns+` FROM public.app_user WHERE external_id = $1`, externalID))
	if errors.Is(err, sql.ErrNoRows) {
		return domain.User{}, domain.ErrUserNotFound
	}
	if err != nil {
		return domain.User{}, fmt.Errorf("get user by external ID: %w", err)
	}
	return user, nil
}

func (r *Repository) GetByEmail(ctx context.Context, email string) (domain.User, error) {
	user, err := scanUser(r.db.QueryRowContext(ctx, `SELECT `+userColumns+` FROM public.app_user WHERE email = $1`, email))
	if errors.Is(err, sql.ErrNoRows) {
		return domain.User{}, domain.ErrUserNotFound
	}
	if err != nil {
		return domain.User{}, fmt.Errorf("get user by email: %w", err)
	}
	return user, nil
}

func (r *Repository) List(ctx context.Context) ([]domain.User, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT `+userColumns+` FROM public.app_user ORDER BY external_id ASC`)
	if err != nil {
		return nil, fmt.Errorf("list users: %w", err)
	}
	defer rows.Close()

	users := make([]domain.User, 0)
	for rows.Next() {
		user, err := scanUser(rows)
		if err != nil {
			return nil, fmt.Errorf("list users: %w", err)
		}
		users = append(users, user)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("list users: %w", err)
	}
	return users, nil
}

func (r *Repository) SaveIfNotExist(ctx context.Context, user domain.User) (domain.User, error) {
	orcid := any(user.ORCID)
	if user.ORCID == "" {
		orcid = nil
	}

	stored, err := scanUser(r.db.QueryRowContext(ctx, `
		INSERT INTO public.app_user (external_id, email, orcid, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (external_id) DO UPDATE SET external_id = app_user.external_id
		RETURNING `+userColumns,
		user.ExternalID, user.Email, orcid, user.CreatedAt, user.UpdatedAt,
	))
	if err != nil {
		return domain.User{}, fmt.Errorf("save user: %w", err)
	}
	return stored, nil
}

func (r *Repository) Update(ctx context.Context, externalID string, user domain.User) error {
	orcid := any(user.ORCID)
	if user.ORCID == "" {
		orcid = nil
	}

	result, err := r.db.ExecContext(ctx, `
		UPDATE public.app_user
		SET email = $1, orcid = $2, updated_at = $3
		WHERE external_id = $4`, user.Email, orcid, user.UpdatedAt, externalID)
	if err != nil {
		return fmt.Errorf("update user: %w", err)
	}
	changed, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("check updated user: %w", err)
	}
	if changed == 0 {
		return domain.ErrUserNotFound
	}
	return nil
}

func (r *Repository) Delete(ctx context.Context, externalID string) error {
	result, err := r.db.ExecContext(ctx, `DELETE FROM public.app_user WHERE external_id = $1`, externalID)
	if err != nil {
		return fmt.Errorf("delete user: %w", err)
	}
	changed, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("check deleted user: %w", err)
	}
	if changed == 0 {
		return domain.ErrUserNotFound
	}
	return nil
}

type rowScanner interface {
	Scan(dest ...any) error
}

func scanUser(row rowScanner) (domain.User, error) {
	var user domain.User
	var orcid sql.NullString
	if err := row.Scan(&user.ExternalID, &user.Email, &orcid, &user.CreatedAt, &user.UpdatedAt); err != nil {
		return domain.User{}, err
	}
	user.ORCID = orcid.String
	return user, nil
}
