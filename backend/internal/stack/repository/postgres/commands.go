package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	paperDomain "github.com/paperstacks.io/paperstacks/internal/paper/domain"
	"github.com/paperstacks.io/paperstacks/internal/stack/domain"
)

func replacePapers(ctx context.Context, tx pgx.Tx, s domain.Stack) error {
	if _, err := tx.Exec(ctx, `DELETE FROM public.stack_paper sp USING public.paper p
		WHERE sp.uuid_stack=$1 AND sp.uuid_paper=p.uuid AND p.deleted_at IS NULL`, s.UUID); err != nil {
		return fmt.Errorf("remove stack paper links: %w", err)
	}
	for _, paper := range s.Papers {
		if _, err := linkPaper(ctx, tx, s.UUID, paper.UUID); err != nil {
			return err
		}
	}
	return nil
}

// Keep the paper active until this transaction finishes linking it.
func linkPaper(ctx context.Context, tx pgx.Tx, stackID, paperID string) (bool, error) {
	var id string
	err := tx.QueryRow(ctx, `SELECT uuid FROM public.paper WHERE uuid=$1 AND deleted_at IS NULL FOR SHARE`, paperID).Scan(&id)
	if errors.Is(err, pgx.ErrNoRows) {
		return false, paperDomain.ErrPaperNotFound
	}
	if err != nil {
		return false, fmt.Errorf("check stack paper: %w", err)
	}
	result, err := tx.Exec(ctx, `INSERT INTO public.stack_paper (uuid_stack, uuid_paper) VALUES ($1,$2) ON CONFLICT (uuid_stack, uuid_paper) DO NOTHING`, stackID, paperID)
	if err != nil {
		return false, fmt.Errorf("link stack paper: %w", err)
	}
	return result.RowsAffected() > 0, nil
}

func (r *Repository) Create(ctx context.Context, s domain.Stack) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin create stack: %w", err)
	}
	defer tx.Rollback(context.Background())
	_, err = tx.Exec(ctx, `INSERT INTO public.stack (uuid, name, owner_external_id, is_public) VALUES ($1,$2,$3,$4)`, s.UUID, s.Name, s.Owner.ExternalID, s.IsPublic)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.Code == "23505" && (pgErr.ConstraintName == "stack_pk" || pgErr.ConstraintName == "stack_owner_name_uq") {
				return fmt.Errorf("create stack: %w: %w", domain.ErrStackAlreadyExists, err)
			}
			if pgErr.Code == "23503" && pgErr.ConstraintName == "stack_owner_fk" {
				return fmt.Errorf("create stack: %w: %w", domain.ErrInvalidStack, err)
			}
		}
		return fmt.Errorf("create stack: %w", err)
	}
	if err := replacePapers(ctx, tx, s); err != nil {
		return err
	}
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit created stack: %w", err)
	}
	return nil
}

func (r *Repository) Update(ctx context.Context, s domain.Stack) (domain.Stack, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return domain.Stack{}, fmt.Errorf("begin update stack: %w", err)
	}
	defer tx.Rollback(context.Background())
	err = tx.QueryRow(ctx, `UPDATE public.stack SET name=$2, is_public=$3
		WHERE uuid=$1 AND owner_external_id=$4 AND deleted_at IS NULL
		RETURNING created_at, updated_at`, s.UUID, s.Name, s.IsPublic, s.Owner.ExternalID).Scan(&s.CreatedAt, &s.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Stack{}, domain.ErrStackNotFound
	}
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" && pgErr.ConstraintName == "stack_owner_name_uq" {
			return domain.Stack{}, fmt.Errorf("update stack: %w: %w", domain.ErrStackAlreadyExists, err)
		}
		return domain.Stack{}, fmt.Errorf("update stack: %w", err)
	}
	if err := replacePapers(ctx, tx, s); err != nil {
		return domain.Stack{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return domain.Stack{}, fmt.Errorf("commit updated stack: %w", err)
	}
	return r.GetByUUID(ctx, s.UUID)
}

func (r *Repository) Delete(ctx context.Context, id string) error {
	result, err := r.db.Exec(ctx, `UPDATE public.stack SET deleted_at=statement_timestamp() WHERE uuid=$1 AND deleted_at IS NULL`, id)
	if err != nil {
		return fmt.Errorf("soft delete stack: %w", err)
	}
	if result.RowsAffected() == 0 {
		return domain.ErrStackNotFound
	}
	return nil
}

func (r *Repository) changePaper(ctx context.Context, stackID, paperID string, add bool) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin change stack paper: %w", err)
	}
	defer tx.Rollback(context.Background())
	var id string
	err = tx.QueryRow(ctx, `SELECT uuid FROM public.stack WHERE uuid=$1 AND deleted_at IS NULL FOR UPDATE`, stackID).Scan(&id)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.ErrStackNotFound
	}
	if err != nil {
		return fmt.Errorf("lock stack: %w", err)
	}
	var changed bool
	if add {
		changed, err = linkPaper(ctx, tx, stackID, paperID)
		if err != nil {
			return err
		}
	} else {
		result, err := tx.Exec(ctx, `DELETE FROM public.stack_paper WHERE uuid_stack=$1 AND uuid_paper=$2`, stackID, paperID)
		if err != nil {
			return fmt.Errorf("unlink stack paper: %w", err)
		}
		changed = result.RowsAffected() > 0
	}
	if changed {
		if _, err := tx.Exec(ctx, `UPDATE public.stack SET updated_at=statement_timestamp() WHERE uuid=$1`, stackID); err != nil {
			return fmt.Errorf("touch stack: %w", err)
		}
	}
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit stack paper change: %w", err)
	}
	return nil
}

func (r *Repository) AddPaper(ctx context.Context, id string, p paperDomain.Paper) error {
	return r.changePaper(ctx, id, p.UUID, true)
}

func (r *Repository) RemovePaper(ctx context.Context, id, paperID string) error {
	return r.changePaper(ctx, id, paperID, false)
}
