// Package postgres persists papers and their bibliographic data in PostgreSQL.
package postgres

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/paperstacks.io/paperstacks/internal/paper/domain"
)

// Repository persists papers in PostgreSQL through native pgx.
type Repository struct {
	db *pgxpool.Pool
}

var _ domain.Repository = (*Repository)(nil)

// NewRepository shares the application's connection pool. Its owner closes it.
func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

func (r *Repository) GetByUUID(ctx context.Context, uuid string) (domain.Paper, error) {
	paper, err := scanPaper(r.db.QueryRow(ctx, `SELECT `+paperColumns+` FROM public.paper WHERE uuid = $1 AND deleted_at IS NULL`, uuid))
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Paper{}, domain.ErrPaperNotFound
	}
	if err != nil {
		return domain.Paper{}, fmt.Errorf("get paper by UUID: %w", err)
	}

	papers := []domain.Paper{paper}
	if err := r.hydrate(ctx, papers); err != nil {
		return domain.Paper{}, fmt.Errorf("load paper by UUID: %w", err)
	}

	return papers[0], nil
}

func (r *Repository) GetByDOI(ctx context.Context, doi string) (domain.Paper, error) {
	paper, err := scanPaper(r.db.QueryRow(ctx, `SELECT `+paperColumns+` FROM public.paper WHERE doi = $1 AND deleted_at IS NULL`, doi))
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Paper{}, domain.ErrPaperNotFound
	}
	if err != nil {
		return domain.Paper{}, fmt.Errorf("get paper by DOI: %w", err)
	}

	papers := []domain.Paper{paper}
	if err := r.hydrate(ctx, papers); err != nil {
		return domain.Paper{}, fmt.Errorf("load paper by DOI: %w", err)
	}

	return papers[0], nil
}

func (r *Repository) List(ctx context.Context) ([]domain.Paper, error) {
	rows, err := r.db.Query(ctx, `SELECT `+paperColumns+` FROM public.paper WHERE deleted_at IS NULL ORDER BY uuid ASC`)
	if err != nil {
		return nil, fmt.Errorf("list papers: %w", err)
	}
	defer rows.Close()

	papers, err := scanPapers(rows)
	if err != nil {
		return nil, fmt.Errorf("list papers: %w", err)
	}
	if err := r.hydrate(ctx, papers); err != nil {
		return nil, fmt.Errorf("load listed papers: %w", err)
	}

	return papers, nil
}

func (r *Repository) Search(ctx context.Context, opts domain.SearchOptions) (domain.SearchResult, error) {
	if opts.SortBy != "" && opts.SortBy != "title" && opts.SortBy != "year" {
		return domain.SearchResult{}, domain.ErrInvalidSearch
	}
	query := strings.ToLower(opts.Query)

	const match = `
		deleted_at IS NULL AND ($1 = ''
			OR strpos(lower(title), $1) > 0
			OR EXISTS (
				SELECT 1
				FROM unnest(COALESCE(keywords, ARRAY[]::text[])) AS keyword(value)
				WHERE strpos(lower(keyword.value), $1) > 0
			)
		)`

	var total int
	if err := r.db.QueryRow(ctx, `SELECT count(*) FROM public.paper WHERE `+match, query).Scan(&total); err != nil {
		return domain.SearchResult{}, fmt.Errorf("count papers: %w", err)
	}

	page := max(1, opts.Page)
	pageSize := opts.PageSize
	if pageSize < 1 {
		pageSize = max(1, total)
	}
	// Clamp pages before multiplication to avoid overflowing on untrusted input.
	offset := total
	if page-1 <= total/pageSize {
		offset = (page - 1) * pageSize
	}

	orderBy := "uuid ASC"
	switch opts.SortBy {
	case "title":
		if opts.Desc {
			orderBy = "title DESC, doi ASC, uuid ASC"
		} else {
			orderBy = "title ASC, doi ASC, uuid ASC"
		}
	case "year":
		if opts.Desc {
			orderBy = "publication_year DESC NULLS LAST, publication_month DESC NULLS LAST, publication_day DESC NULLS LAST, doi ASC, uuid ASC"
		} else {
			orderBy = "publication_year ASC NULLS FIRST, publication_month ASC NULLS FIRST, publication_day ASC NULLS FIRST, doi ASC, uuid ASC"
		}
	}

	rows, err := r.db.Query(ctx, `SELECT `+paperColumns+` FROM public.paper WHERE `+match+` ORDER BY `+orderBy+` LIMIT $2 OFFSET $3`, query, pageSize, offset)
	if err != nil {
		return domain.SearchResult{}, fmt.Errorf("search papers: %w", err)
	}
	defer rows.Close()

	papers, err := scanPapers(rows)
	if err != nil {
		return domain.SearchResult{}, fmt.Errorf("search papers: %w", err)
	}
	if err := r.hydrate(ctx, papers); err != nil {
		return domain.SearchResult{}, fmt.Errorf("load searched papers: %w", err)
	}

	return domain.SearchResult{
		Items:    papers,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
		HasNext:  len(papers) < total-offset,
	}, nil
}

func (r *Repository) Save(ctx context.Context, paper domain.Paper) (domain.Paper, error) {
	paper = paper.Normalize()

	tx, err := r.db.Begin(ctx)
	if err != nil {
		return domain.Paper{}, fmt.Errorf("begin save paper: %w", err)
	}
	defer tx.Rollback(context.Background())

	if err := insertPaper(ctx, tx, paper); err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			switch pgErr.ConstraintName {
			case "paper_pk", "paper_doi_uq":
				return domain.Paper{}, fmt.Errorf("save paper: %w: %w", domain.ErrPaperAlreadyExists, err)
			}
		}
		return domain.Paper{}, fmt.Errorf("save paper: %w", err)
	}
	if err := insertChildren(ctx, tx, paper); err != nil {
		return domain.Paper{}, fmt.Errorf("save paper: %w", err)
	}
	if err := tx.Commit(ctx); err != nil {
		return domain.Paper{}, fmt.Errorf("commit saved paper: %w", err)
	}

	return paper, nil
}

func (r *Repository) Update(ctx context.Context, uuid string, paper domain.Paper) error {
	if paper.UUID != "" && paper.UUID != uuid {
		return domain.ErrUUIDMismatch
	}
	paper.UUID = uuid
	paper = paper.Normalize()

	tx, err := r.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin update paper: %w", err)
	}
	defer tx.Rollback(context.Background())

	if err := updatePaper(ctx, tx, uuid, paper); err != nil {
		return err
	}

	if err := removePaperAuthors(ctx, tx, uuid); err != nil {
		return fmt.Errorf("replace paper authors: %w", err)
	}
	if _, err := tx.Exec(ctx, `DELETE FROM public.metadata WHERE uuid_paper = $1`, uuid); err != nil {
		return fmt.Errorf("replace paper metadata: %w", err)
	}
	if err := insertChildren(ctx, tx, paper); err != nil {
		return fmt.Errorf("replace paper children: %w", err)
	}
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit updated paper: %w", err)
	}

	return nil
}

// Delete marks an active paper as deleted; the database retention job purges it later.
func (r *Repository) Delete(ctx context.Context, uuid string) error {
	result, err := r.db.Exec(ctx, `
		UPDATE public.paper
		SET deleted_at = statement_timestamp()
		WHERE uuid = $1 AND deleted_at IS NULL`, uuid)
	if err != nil {
		return fmt.Errorf("soft delete paper: %w", err)
	}
	if result.RowsAffected() == 0 {
		return domain.ErrPaperNotFound
	}
	return nil
}

func (r *Repository) hydrate(ctx context.Context, papers []domain.Paper) error {
	if len(papers) == 0 {
		return nil
	}

	byUUID := make(map[string]*domain.Paper, len(papers))
	uuids := make([]string, len(papers))
	for i := range papers {
		byUUID[papers[i].UUID] = &papers[i]
		uuids[i] = papers[i].UUID
	}

	if err := loadMetadata(ctx, r.db, byUUID, uuids); err != nil {
		return fmt.Errorf("load metadata: %w", err)
	}
	if err := loadAuthors(ctx, r.db, byUUID, uuids); err != nil {
		return fmt.Errorf("load authors: %w", err)
	}
	return nil
}

func insertChildren(ctx context.Context, tx pgx.Tx, paper domain.Paper) error {
	if err := insertMetadata(ctx, tx, paper.UUID, paper.Metadata); err != nil {
		return fmt.Errorf("insert metadata: %w", err)
	}
	if err := insertAuthors(ctx, tx, paper.UUID, paper.Authors); err != nil {
		return fmt.Errorf("insert authors: %w", err)
	}
	return nil
}
