package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5/pgconn"
	paperDomain "github.com/paperstacks.io/paperstacks/internal/paper/domain"
	"github.com/paperstacks.io/paperstacks/internal/stack/domain"
)

func (r *Repository) transaction(ctx context.Context, op string, fn func(*sql.Tx) error) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return wrapError(op, err)
	}
	defer tx.Rollback()
	if err = fn(tx); err != nil {
		return wrapError(op, err)
	}
	return wrapError(op, tx.Commit())
}

func replacePapers(ctx context.Context, tx *sql.Tx, s domain.Stack) error {
	if _, err := tx.ExecContext(ctx, `DELETE FROM public.stack_paper WHERE uuid_stack = $1`, s.UUID); err != nil {
		return err
	}
	for _, p := range s.Papers {
		if _, err := tx.ExecContext(ctx, `INSERT INTO public.stack_paper (uuid_stack, uuid_paper) VALUES ($1, $2) ON CONFLICT DO NOTHING`, s.UUID, p.UUID); err != nil {
			return err
		}
	}
	return nil
}

func (r *Repository) Create(ctx context.Context, s domain.Stack) error {
	return r.transaction(ctx, "create stack", func(tx *sql.Tx) error {
		_, err := tx.ExecContext(ctx, `INSERT INTO public.stack (uuid, name, owner_external_id, is_public, created_at, updated_at) VALUES ($1,$2,$3,$4,$5,$6)`, s.UUID, s.Name, s.Owner.ExternalID, s.IsPublic, s.CreatedAt, s.UpdatedAt)
		if err != nil {
			return err
		}
		return replacePapers(ctx, tx, s)
	})
}

func requireChanged(result sql.Result, err error) error {
	if err != nil {
		return err
	}
	n, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return domain.ErrStackNotFound
	}
	return nil
}

func (r *Repository) Update(ctx context.Context, s domain.Stack) (domain.Stack, error) {
	err := r.transaction(ctx, "update stack", func(tx *sql.Tx) error {
		err := tx.QueryRowContext(ctx, `UPDATE public.stack SET name=$2, is_public=$3, updated_at=$4 WHERE uuid=$1 AND owner_external_id=$5 RETURNING created_at, updated_at`, s.UUID, s.Name, s.IsPublic, s.UpdatedAt, s.Owner.ExternalID).Scan(&s.CreatedAt, &s.UpdatedAt)
		if errors.Is(err, sql.ErrNoRows) {
			return domain.ErrStackNotFound
		}
		if err != nil {
			return err
		}
		return replacePapers(ctx, tx, s)
	})
	if err != nil {
		return domain.Stack{}, err
	}
	return s, nil
}

func (r *Repository) Delete(ctx context.Context, id string) error {
	return wrapError("delete stack", requireChanged(r.db.ExecContext(ctx, `DELETE FROM public.stack WHERE uuid=$1`, id)))
}

func (r *Repository) changePaper(ctx context.Context, stackID, paperID string, add bool) error {
	return r.transaction(ctx, "change stack paper", func(tx *sql.Tx) error {
		var id string
		err := tx.QueryRowContext(ctx, `SELECT uuid::text FROM public.stack WHERE uuid=$1 FOR UPDATE`, stackID).Scan(&id)
		if errors.Is(err, sql.ErrNoRows) {
			return domain.ErrStackNotFound
		}
		if err != nil {
			return err
		}
		query := `DELETE FROM public.stack_paper WHERE uuid_stack=$1 AND uuid_paper=$2`
		if add {
			query = `INSERT INTO public.stack_paper (uuid_stack, uuid_paper) VALUES ($1,$2) ON CONFLICT DO NOTHING`
		}
		result, err := tx.ExecContext(ctx, query, stackID, paperID)
		if err != nil {
			var pgErr *pgconn.PgError
			if errors.As(err, &pgErr) && pgErr.Code == "23503" {
				return fmt.Errorf("%w: %w", paperDomain.ErrPaperNotFound, err)
			}
			return err
		}
		n, err := result.RowsAffected()
		if err != nil {
			return err
		}
		if n > 0 {
			_, err = tx.ExecContext(ctx, `UPDATE public.stack SET updated_at=statement_timestamp() WHERE uuid=$1`, stackID)
		}
		return err
	})
}

func (r *Repository) AddPaper(ctx context.Context, id string, p paperDomain.Paper) error {
	return r.changePaper(ctx, id, p.UUID, true)
}

func (r *Repository) RemovePaper(ctx context.Context, id, paperID string) error {
	return r.changePaper(ctx, id, paperID, false)
}
