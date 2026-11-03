package postgres

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	paperDomain "github.com/paperstacks.io/paperstacks/internal/paper/domain"
	paperPostgres "github.com/paperstacks.io/paperstacks/internal/paper/repository/postgres"
	"github.com/paperstacks.io/paperstacks/internal/stack/domain"
)

// Repository persists stacks and their paper associations through native pgx.
type Repository struct {
	db     *pgxpool.Pool
	papers *paperPostgres.Repository
}

var _ domain.Repository = (*Repository)(nil)

// NewRepository shares the application's connection pool. Its owner closes it.
func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db: db, papers: paperPostgres.NewRepository(db)}
}

func (r *Repository) GetByUUID(ctx context.Context, id string) (domain.Stack, error) {
	const query = `
		SELECT s.uuid, s.name, s.owner_external_id, s.is_public,
		       s.created_at, s.updated_at, u.email, u.orcid, u.created_at, u.updated_at
		FROM public.stack s
		JOIN public.app_user u ON u.external_id = s.owner_external_id
		WHERE s.uuid = $1 AND s.deleted_at IS NULL`
	s, err := scanStack(r.db.QueryRow(ctx, query, id))
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Stack{}, domain.ErrStackNotFound
	}
	if err != nil {
		return domain.Stack{}, fmt.Errorf("get stack: %w", err)
	}
	items := []domain.Stack{s}
	if err := r.loadPapers(ctx, items); err != nil {
		return domain.Stack{}, err
	}
	return items[0], nil
}

func (r *Repository) StatsByOwner(ctx context.Context, owner string) (domain.Stats, error) {
	const query = `
		SELECT count(*), count(*) FILTER (WHERE is_public),
		       (SELECT count(*)
		        FROM public.stack_paper sp
		        JOIN public.stack s ON s.uuid = sp.uuid_stack
		        JOIN public.paper p ON p.uuid = sp.uuid_paper
		        WHERE s.owner_external_id = $1
		          AND s.deleted_at IS NULL AND p.deleted_at IS NULL)
		FROM public.stack
		WHERE owner_external_id = $1 AND deleted_at IS NULL`
	var stats domain.Stats
	err := r.db.QueryRow(ctx, query, owner).Scan(&stats.TotalStacks, &stats.PublicStacks, &stats.TotalPapers)
	if err != nil {
		return domain.Stats{}, fmt.Errorf("stack statistics: %w", err)
	}
	return stats, nil
}

func (r *Repository) SearchPublic(ctx context.Context, opts domain.SearchOptions) (domain.SearchResult, error) {
	return r.search(ctx, `s.is_public`, nil, opts)
}

func (r *Repository) SearchByOwner(ctx context.Context, owner string, opts domain.SearchOptions) (domain.SearchResult, error) {
	return r.search(ctx, `s.owner_external_id = $1`, []any{owner}, opts)
}

func (r *Repository) List(ctx context.Context, owner string) ([]domain.Stack, error) {
	const query = `
		SELECT s.uuid, s.name, s.owner_external_id, s.is_public,
		       s.created_at, s.updated_at, u.email, u.orcid, u.created_at, u.updated_at
		FROM public.stack s
		JOIN public.app_user u ON u.external_id = s.owner_external_id
		WHERE s.deleted_at IS NULL AND s.owner_external_id = $1
		ORDER BY s.uuid`
	rows, err := r.db.Query(ctx, query, owner)
	if err != nil {
		return nil, fmt.Errorf("list stacks: %w", err)
	}
	items, err := scanStacks(rows)
	if err != nil {
		return nil, fmt.Errorf("list stacks: %w", err)
	}
	if err := r.loadPapers(ctx, items); err != nil {
		return nil, err
	}
	return items, nil
}

func (r *Repository) ListPublic(ctx context.Context, owner string) ([]domain.Stack, error) {
	const query = `
		SELECT s.uuid, s.name, s.owner_external_id, s.is_public,
		       s.created_at, s.updated_at, u.email, u.orcid, u.created_at, u.updated_at
		FROM public.stack s
		JOIN public.app_user u ON u.external_id = s.owner_external_id
		WHERE s.deleted_at IS NULL AND s.owner_external_id = $1 AND s.is_public
		ORDER BY s.uuid`
	rows, err := r.db.Query(ctx, query, owner)
	if err != nil {
		return nil, fmt.Errorf("list public stacks: %w", err)
	}
	items, err := scanStacks(rows)
	if err != nil {
		return nil, fmt.Errorf("list public stacks: %w", err)
	}
	if err := r.loadPapers(ctx, items); err != nil {
		return nil, err
	}
	return items, nil
}

func (r *Repository) CountPublic(ctx context.Context) (int, error) {
	var count int
	err := r.db.QueryRow(ctx, `SELECT count(*) FROM public.stack WHERE deleted_at IS NULL AND is_public`).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("count public stacks: %w", err)
	}
	return count, nil
}

func (r *Repository) loadPapers(ctx context.Context, stacks []domain.Stack) error {
	if len(stacks) == 0 {
		return nil
	}
	stackIDs := make([]string, len(stacks))
	for i := range stacks {
		stackIDs[i] = stacks[i].UUID
	}
	rows, err := r.db.Query(ctx, `SELECT sp.uuid_stack, sp.uuid_paper
		FROM public.stack_paper sp JOIN public.paper p ON p.uuid = sp.uuid_paper
		WHERE sp.uuid_stack = ANY($1::uuid[]) AND p.deleted_at IS NULL ORDER BY sp.uuid_paper`, stackIDs)
	if err != nil {
		return fmt.Errorf("list stack papers: %w", err)
	}
	var paperIDs []string
	links := make(map[string][]string, len(stacks))
	for rows.Next() {
		var stackID, paperID string
		if err := rows.Scan(&stackID, &paperID); err != nil {
			rows.Close()
			return fmt.Errorf("scan stack paper: %w", err)
		}
		paperIDs = append(paperIDs, paperID)
		links[stackID] = append(links[stackID], paperID)
	}
	err = rows.Err()
	rows.Close() // Release the connection before loading papers from the shared pool.
	if err != nil {
		return fmt.Errorf("read stack papers: %w", err)
	}
	papers, err := r.papers.GetByUUIDs(ctx, paperIDs)
	if err != nil {
		return fmt.Errorf("load stack papers: %w", err)
	}
	byID := make(map[string]paperDomain.Paper, len(papers))
	for _, p := range papers {
		byID[p.UUID] = p
	}
	for i := range stacks {
		for _, id := range links[stacks[i].UUID] {
			if p, ok := byID[id]; ok {
				stacks[i].Papers = append(stacks[i].Papers, p)
			}
		}
	}
	return nil
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
	where := fmt.Sprintf(` WHERE s.deleted_at IS NULL AND %s AND strpos(lower(s.name), $%d) > 0`, filter, len(args))
	var total int
	if err := r.db.QueryRow(ctx, `SELECT count(*) FROM public.stack s`+where, args...).Scan(&total); err != nil {
		return domain.SearchResult{}, fmt.Errorf("count search results: %w", err)
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
	query := fmt.Sprintf(`
		SELECT s.uuid, s.name, s.owner_external_id, s.is_public,
		       s.created_at, s.updated_at, u.email, u.orcid, u.created_at, u.updated_at
		FROM public.stack s
		JOIN public.app_user u ON u.external_id = s.owner_external_id
		%s ORDER BY %s LIMIT $%d OFFSET $%d`, where, order, len(args)-1, len(args))
	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return domain.SearchResult{}, fmt.Errorf("search stacks: %w", err)
	}
	items, err := scanStacks(rows)
	if err != nil {
		return domain.SearchResult{}, fmt.Errorf("search stacks: %w", err)
	}
	if err := r.loadPapers(ctx, items); err != nil {
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
