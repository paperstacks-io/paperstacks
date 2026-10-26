package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/paperstacks.io/paperstacks/internal/paper/domain"
)

const metadataColumns = `
	uuid_paper::text,
	publisher,
	journal_title,
	journal_abbrev,
	book_title,
	series_title,
	event_title,
	event_place,
	institution,
	pages,
	volume,
	issue,
	isbn,
	issn,
	"reference",
	license,
	copyright,
	funding,
	datasource,
	datasource_timestamp`

func loadMetadata(ctx context.Context, db *pgxpool.Pool, papers map[string]*domain.Paper, uuids []string) error {
	rows, err := db.Query(ctx, `SELECT `+metadataColumns+` FROM public.metadata WHERE uuid_paper = ANY($1::uuid[])`, uuids)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		uuid, metadata, err := scanMetadata(rows)
		if err != nil {
			return err
		}
		if paper := papers[uuid]; paper != nil {
			paper.Metadata = metadata
		}
	}

	return rows.Err()
}

func scanMetadata(row pgx.Row) (string, domain.Metadata, error) {
	var d metadataDAO
	err := row.Scan(
		&d.PaperUUID, &d.Publisher, &d.JournalTitle, &d.JournalAbbrev,
		&d.BookTitle, &d.SeriesTitle, &d.EventTitle, &d.EventLocation, &d.Institution,
		&d.Pages, &d.Volume, &d.Issue,
		&d.ISBN, &d.ISSN, &d.References,
		&d.License, &d.Copyright, &d.Funding, &d.DataSource, &d.DataSourceTimestamp,
	)
	if err != nil {
		return "", domain.Metadata{}, err
	}
	return d.PaperUUID, d.domain(), nil
}

func insertMetadata(ctx context.Context, tx pgx.Tx, uuid string, metadata domain.Metadata) error {
	timestamp, err := timestampValue(metadata.DataSourceTimestamp)
	if err != nil {
		return fmt.Errorf("parse data source timestamp: %w", err)
	}
	_, err = tx.Exec(ctx, `
		INSERT INTO public.metadata (
			uuid_paper, publisher, journal_title, journal_abbrev, book_title, series_title, event_title,
			event_place, institution, pages, volume, issue, isbn, issn, "reference", license, copyright,
			funding, datasource, datasource_timestamp
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20
		)`,
		uuid,
		nullString(metadata.Publisher),
		nullString(metadata.JournalTitle),
		nullString(metadata.JournalAbbrev),
		nullString(metadata.BookTitle),
		nullString(metadata.SeriesTitle),
		nullString(metadata.EventTitle),
		nullString(metadata.EventLocation),
		nullString(metadata.Institution),
		nullString(metadata.Pages),
		nullString(metadata.Volume),
		nullString(metadata.Issue),
		metadata.ISBN,
		metadata.ISSN,
		metadata.References,
		nullString(metadata.License),
		nullString(metadata.Copyright),
		nullString(metadata.Funding),
		nullString(metadata.DataSource),
		timestamp,
	)
	return err
}
