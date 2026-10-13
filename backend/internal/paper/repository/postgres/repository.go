package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5/pgconn"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/paperstacks.io/paperstacks/internal/paper/domain"
)

// Repository persists papers in PostgreSQL through database/sql.
type Repository struct {
	db *sql.DB
}

var _ domain.Repository = (*Repository)(nil)

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) GetByUUID(ctx context.Context, uuid string) (domain.Paper, error) {
	paper, err := scanPaper(r.db.QueryRowContext(ctx, `SELECT `+paperColumns+` FROM public.paper WHERE uuid = $1`, uuid))
	if errors.Is(err, sql.ErrNoRows) {
		return domain.Paper{}, domain.ErrPaperNotFound
	}
	if err != nil {
		return domain.Paper{}, wrapError("get paper by UUID", err)
	}

	papers := []domain.Paper{paper}
	if err := r.hydrate(ctx, papers); err != nil {
		return domain.Paper{}, wrapError("load paper by UUID", err)
	}

	return papers[0], nil
}

func (r *Repository) GetByDOI(ctx context.Context, doi string) (domain.Paper, error) {
	paper, err := scanPaper(r.db.QueryRowContext(ctx, `SELECT `+paperColumns+` FROM public.paper WHERE doi = $1`, doi))
	if errors.Is(err, sql.ErrNoRows) {
		return domain.Paper{}, domain.ErrPaperNotFound
	}
	if err != nil {
		return domain.Paper{}, wrapError("get paper by DOI", err)
	}

	papers := []domain.Paper{paper}
	if err := r.hydrate(ctx, papers); err != nil {
		return domain.Paper{}, wrapError("load paper by DOI", err)
	}

	return papers[0], nil
}

func (r *Repository) List(ctx context.Context) ([]domain.Paper, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT `+paperColumns+` FROM public.paper ORDER BY uuid ASC`)
	if err != nil {
		return nil, wrapError("list papers", err)
	}
	defer rows.Close()

	papers, err := scanPapers(rows)
	if err != nil {
		return nil, wrapError("list papers", err)
	}
	if err := r.hydrate(ctx, papers); err != nil {
		return nil, wrapError("load listed papers", err)
	}

	return papers, nil
}

func (r *Repository) Search(ctx context.Context, opts domain.SearchOptions) (domain.SearchResult, error) {
	query := strings.ToLower(opts.Query)

	const match = `
		($1 = ''
			OR strpos(lower(title), $1) > 0
			OR EXISTS (
				SELECT 1
				FROM unnest(COALESCE(keywords, ARRAY[]::text[])) AS keyword(value)
				WHERE strpos(lower(keyword.value), $1) > 0
			)
		)`

	var total int
	if err := r.db.QueryRowContext(ctx, `SELECT count(*) FROM public.paper WHERE `+match, query).Scan(&total); err != nil {
		return domain.SearchResult{}, wrapError("count papers", err)
	}

	page := max(1, opts.Page)
	pageSize := opts.PageSize
	if pageSize < 1 {
		pageSize = max(1, total)
	}
	offset := (page - 1) * pageSize

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

	rows, err := r.db.QueryContext(ctx, `SELECT `+paperColumns+` FROM public.paper WHERE `+match+` ORDER BY `+orderBy+` LIMIT $2 OFFSET $3`, query, pageSize, offset)
	if err != nil {
		return domain.SearchResult{}, wrapError("search papers", err)
	}
	defer rows.Close()

	papers, err := scanPapers(rows)
	if err != nil {
		return domain.SearchResult{}, wrapError("search papers", err)
	}
	if err := r.hydrate(ctx, papers); err != nil {
		return domain.SearchResult{}, wrapError("load searched papers", err)
	}

	return domain.SearchResult{
		Items:    papers,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
		HasNext:  offset+len(papers) < total,
	}, nil
}

func (r *Repository) Save(ctx context.Context, paper domain.Paper) (domain.Paper, error) {
	paper = paper.Normalize()

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return domain.Paper{}, wrapError("begin save paper", err)
	}
	defer tx.Rollback()

	if err := insertPaper(ctx, tx, paper); err != nil {
		return domain.Paper{}, wrapError("save paper", err)
	}
	if err := insertChildren(ctx, tx, paper); err != nil {
		return domain.Paper{}, wrapError("save paper", err)
	}
	if err := tx.Commit(); err != nil {
		return domain.Paper{}, wrapError("commit saved paper", err)
	}

	return paper, nil
}

func (r *Repository) Update(ctx context.Context, uuid string, paper domain.Paper) error {
	paper = paper.Normalize()

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return wrapError("begin update paper", err)
	}
	defer tx.Rollback()

	if err := updatePaper(ctx, tx, uuid, paper); err != nil {
		return err
	}

	if err := removePaperAuthors(ctx, tx, uuid); err != nil {
		return wrapError("replace paper authors", err)
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM public.metadata WHERE uuid_paper = $1`, uuid); err != nil {
		return wrapError("replace paper metadata", err)
	}
	if err := insertChildren(ctx, tx, paper); err != nil {
		return wrapError("replace paper children", err)
	}
	if err := tx.Commit(); err != nil {
		return wrapError("commit updated paper", err)
	}

	return nil
}

func (r *Repository) Delete(ctx context.Context, uuid string) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return wrapError("begin delete paper", err)
	}
	defer tx.Rollback()

	if err := removePaperAuthors(ctx, tx, uuid); err != nil {
		return wrapError("remove paper authors", err)
	}
	result, err := tx.ExecContext(ctx, `DELETE FROM public.paper WHERE uuid = $1`, uuid)
	if err != nil {
		return wrapError("delete paper", err)
	}
	changed, err := result.RowsAffected()
	if err != nil {
		return wrapError("check deleted paper", err)
	}
	if changed == 0 {
		return domain.ErrPaperNotFound
	}
	if err := tx.Commit(); err != nil {
		return wrapError("commit deleted paper", err)
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

func insertChildren(ctx context.Context, tx *sql.Tx, paper domain.Paper) error {
	if err := insertMetadata(ctx, tx, paper.UUID, paper.Metadata); err != nil {
		return fmt.Errorf("insert metadata: %w", err)
	}
	if err := insertAuthors(ctx, tx, paper.UUID, paper.Authors); err != nil {
		return fmt.Errorf("insert authors: %w", err)
	}
	return nil
}

func wrapError(operation string, err error) error {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23505" {
		return fmt.Errorf("%s: %w: %w", operation, domain.ErrPaperAlreadyExists, err)
	}
	return fmt.Errorf("%s: %w", operation, err)
}
