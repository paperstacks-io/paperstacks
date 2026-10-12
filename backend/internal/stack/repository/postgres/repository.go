package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5/pgconn"
	paperDomain "github.com/paperstacks.io/paperstacks/internal/paper/domain"
	paperPostgres "github.com/paperstacks.io/paperstacks/internal/paper/repository/postgres"
	"github.com/paperstacks.io/paperstacks/internal/stack/domain"
)

type paperGetter interface {
	GetByUUID(context.Context, string) (paperDomain.Paper, error)
}

// Repository persists stacks and their paper associations through database/sql.
type Repository struct {
	db     *sql.DB
	papers paperGetter
}

var _ domain.Repository = (*Repository)(nil)

// NewRepository shares the application's connection pool. Its owner closes it.
func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db, papers: paperPostgres.NewRepository(db)}
}

func wrapError(op string, err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, paperDomain.ErrPaperNotFound) {
		return fmt.Errorf("%s: %w", op, err)
	}
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case "23505":
			return fmt.Errorf("%s: %w: %w", op, domain.ErrStackAlreadyExists, err)
		case "23503", "23502", "23514", "22P02":
			return fmt.Errorf("%s: %w: %w", op, domain.ErrInvalidStack, err)
		}
	}
	return fmt.Errorf("%s: %w", op, err)
}

func (r *Repository) GetByUUID(ctx context.Context, id string) (domain.Stack, error) {
	s, err := scanStack(r.db.QueryRowContext(ctx, `SELECT `+stackColumns+stackFrom+` WHERE s.uuid = $1`, id))
	if errors.Is(err, sql.ErrNoRows) {
		return domain.Stack{}, domain.ErrStackNotFound
	}
	if err != nil {
		return domain.Stack{}, wrapError("get stack", err)
	}
	if err := r.loadPapers(ctx, &s); err != nil {
		return domain.Stack{}, err
	}
	return s, nil
}

func (r *Repository) loadPapers(ctx context.Context, s *domain.Stack) error {
	rows, err := r.db.QueryContext(ctx, `SELECT uuid_paper::text FROM public.stack_paper WHERE uuid_stack = $1 ORDER BY uuid_paper`, s.UUID)
	if err != nil {
		return wrapError("list stack papers", err)
	}
	var ids []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			rows.Close()
			return wrapError("scan stack paper", err)
		}
		ids = append(ids, id)
	}
	err = rows.Err()
	rows.Close() // Release the connection before loading papers from the shared pool.
	if err != nil {
		return wrapError("read stack papers", err)
	}
	for _, id := range ids {
		p, err := r.papers.GetByUUID(ctx, id)
		if err != nil {
			return wrapError("load stack paper", err)
		}
		s.Papers = append(s.Papers, p)
	}
	return nil
}

func (r *Repository) list(ctx context.Context, suffix string, args ...any) ([]domain.Stack, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT `+stackColumns+stackFrom+suffix, args...)
	if err != nil {
		return nil, wrapError("list stacks", err)
	}
	items := make([]domain.Stack, 0)
	for rows.Next() {
		s, err := scanStack(rows)
		if err != nil {
			rows.Close()
			return nil, wrapError("scan stack", err)
		}
		items = append(items, s)
	}
	err = rows.Err()
	rows.Close()
	if err != nil {
		return nil, wrapError("read stacks", err)
	}
	for i := range items {
		if err := r.loadPapers(ctx, &items[i]); err != nil {
			return nil, err
		}
	}
	return items, nil
}

func (r *Repository) List(ctx context.Context, owner string) ([]domain.Stack, error) {
	return r.list(ctx, ` WHERE s.owner_external_id = $1 ORDER BY s.uuid`, owner)
}

func (r *Repository) ListPublic(ctx context.Context, owner string) ([]domain.Stack, error) {
	return r.list(ctx, ` WHERE s.owner_external_id = $1 AND s.is_public ORDER BY s.uuid`, owner)
}

func (r *Repository) CountPublic(ctx context.Context) (int, error) {
	var count int
	err := r.db.QueryRowContext(ctx, `SELECT count(*) FROM public.stack WHERE is_public`).Scan(&count)
	return count, wrapError("count public stacks", err)
}

func (r *Repository) StatsByOwner(ctx context.Context, owner string) (domain.Stats, error) {
	var stats domain.Stats
	err := r.db.QueryRowContext(ctx, `SELECT count(*), count(*) FILTER (WHERE is_public),
 (SELECT count(*) FROM public.stack_paper sp JOIN public.stack s ON s.uuid = sp.uuid_stack WHERE s.owner_external_id = $1)
 FROM public.stack WHERE owner_external_id = $1`, owner).Scan(&stats.TotalStacks, &stats.PublicStacks, &stats.TotalPapers)
	return stats, wrapError("stack statistics", err)
}

func (r *Repository) SearchPublic(ctx context.Context, opts domain.SearchOptions) (domain.SearchResult, error) {
	return r.search(ctx, `s.is_public`, nil, opts)
}

func (r *Repository) SearchByOwner(ctx context.Context, owner string, opts domain.SearchOptions) (domain.SearchResult, error) {
	return r.search(ctx, `s.owner_external_id = $1`, []any{owner}, opts)
}

func (r *Repository) search(ctx context.Context, filter string, args []any, opts domain.SearchOptions) (domain.SearchResult, error) {
	order := "s.uuid"
	switch opts.SortBy {
	case "":
	case "name":
		order = "lower(s.name)"
	case "updated_at":
		order = "s.updated_at"
	default:
		return domain.SearchResult{}, domain.ErrInvalidSearch
	}
	if opts.Desc && opts.SortBy != "" {
		order += " DESC"
	}
	order += ", s.uuid ASC"
	args = append(args, strings.ToLower(strings.TrimSpace(opts.Query)))
	where := fmt.Sprintf(` WHERE %s AND strpos(lower(s.name), $%d) > 0`, filter, len(args))
	var total int
	if err := r.db.QueryRowContext(ctx, `SELECT count(*) FROM public.stack s`+where, args...).Scan(&total); err != nil {
		return domain.SearchResult{}, wrapError("count search results", err)
	}
	page, size := max(1, opts.Page), opts.PageSize
	if size <= 0 {
		size = max(1, total)
	}
	// Clamp out-of-range pages before multiplication to avoid integer overflow.
	offset := total
	if page-1 <= total/size {
		offset = (page - 1) * size
	}
	args = append(args, size, offset)
	suffix := fmt.Sprintf(`%s ORDER BY %s LIMIT $%d OFFSET $%d`, where, order, len(args)-1, len(args))
	items, err := r.list(ctx, suffix, args...)
	if err != nil {
		return domain.SearchResult{}, err
	}
	return domain.SearchResult{
		Items:    items,
		Total:    total,
		Page:     page,
		PageSize: size,
		HasNext:  len(items) < total-offset,
	}, nil
}
