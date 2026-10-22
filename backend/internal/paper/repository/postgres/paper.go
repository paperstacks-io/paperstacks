package postgres

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/paperstacks.io/paperstacks/internal/paper/domain"
)

const paperColumns = `
	uuid::text,
	doi,
	title,
	title_short,
	publication_year,
	publication_month,
	publication_day,
	paper_type::text,
	publication_status::text,
	publication_status_timestamp,
	abstract,
	keywords,
	pdf_url
`

func scanPapers(rows *sql.Rows) ([]domain.Paper, error) {
	papers := make([]domain.Paper, 0)
	for rows.Next() {
		paper, err := scanPaper(rows)
		if err != nil {
			return nil, err
		}
		papers = append(papers, paper)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return papers, nil
}

type rowScanner interface {
	Scan(dest ...any) error
}

func scanPaper(row rowScanner) (domain.Paper, error) {
	pgTypes := pgtype.NewMap()
	var d paperDAO
	err := row.Scan(
		&d.UUID, &d.DOI, &d.Title, &d.TitleShort,
		&d.Year, &d.Month, &d.Day, &d.Type, &d.PublicationStatus,
		&d.StatusTimestamp, &d.Abstract, pgTypes.SQLScanner(&d.Keywords), &d.PDFURL,
	)
	if err != nil {
		return domain.Paper{}, err
	}
	return d.domain(), nil
}

func updatePaper(ctx context.Context, tx *sql.Tx, uuid string, paper domain.Paper) error {
	args, err := paperArguments(uuid, paper)
	if err != nil {
		return wrapError("prepare paper update", err)
	}
	result, err := tx.ExecContext(ctx, `
		UPDATE public.paper
		SET doi = $2,
			title = $3,
			title_short = $4,
			publication_year = $5,
			publication_month = $6,
			publication_day = $7,
			paper_type = $8,
			publication_status = $9,
			publication_status_timestamp = $10,
			abstract = $11,
			keywords = $12,
			pdf_url = $13
		WHERE uuid = $1`, args...)
	if err != nil {
		return wrapError("update paper", err)
	}
	changed, err := result.RowsAffected()
	if err != nil {
		return wrapError("check updated paper", err)
	}
	if changed == 0 {
		return domain.ErrPaperNotFound
	}

	return nil
}

func insertPaper(ctx context.Context, tx *sql.Tx, paper domain.Paper) error {
	args, err := paperArguments(paper.UUID, paper)
	if err != nil {
		return err
	}
	_, err = tx.ExecContext(ctx, `
		INSERT INTO public.paper (
			uuid, doi, title, title_short, publication_year, publication_month, publication_day,
			paper_type, publication_status, publication_status_timestamp, abstract, keywords, pdf_url
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)`, args...)
	return err
}

func paperArguments(uuid string, paper domain.Paper) ([]any, error) {
	statusTimestamp, err := timestampValue(paper.PublicationStatusTimestamp)
	if err != nil {
		return nil, fmt.Errorf("parse publication status timestamp: %w", err)
	}

	return []any{
		uuid,
		nullString(paper.DOI),
		nullString(paper.Title),
		nullString(paper.TitleShort),
		nullInt(paper.PublicationDate.Year),
		nullInt(paper.PublicationDate.Month),
		nullInt(paper.PublicationDate.Day),
		nullString(string(paper.Type)),
		nullString(paper.PublicationStatus),
		statusTimestamp,
		nullString(paper.Abstract),
		pgtype.FlatArray[string](paper.Keywords),
		nullString(paper.PDFURL),
	}, nil
}
