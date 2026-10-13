package postgres

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/jackc/pgx/v5/pgtype"
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

func loadMetadata(ctx context.Context, db *sql.DB, papers map[string]*domain.Paper, uuids []string) error {
	rows, err := db.QueryContext(ctx, `SELECT `+metadataColumns+` FROM public.metadata WHERE uuid_paper = ANY($1::uuid[])`, pgtype.FlatArray[string](uuids))
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

func scanMetadata(row rowScanner) (string, domain.Metadata, error) {
	pgTypes := pgtype.NewMap()
	var uuid string
	var metadata domain.Metadata
	var publisher, journalTitle, journalAbbrev, bookTitle, seriesTitle, eventTitle, eventLocation, institution sql.NullString
	var pages, volume, issue, license, copyright, funding, dataSource sql.NullString
	var dataSourceTimestamp sql.NullTime

	err := row.Scan(
		&uuid,
		&publisher,
		&journalTitle,
		&journalAbbrev,
		&bookTitle,
		&seriesTitle,
		&eventTitle,
		&eventLocation,
		&institution,
		&pages,
		&volume,
		&issue,
		pgTypes.SQLScanner(&metadata.ISBN),
		pgTypes.SQLScanner(&metadata.ISSN),
		pgTypes.SQLScanner(&metadata.References),
		&license,
		&copyright,
		&funding,
		&dataSource,
		&dataSourceTimestamp,
	)
	if err != nil {
		return "", domain.Metadata{}, err
	}

	metadata.Publisher = publisher.String
	metadata.JournalTitle = journalTitle.String
	metadata.JournalAbbrev = journalAbbrev.String
	metadata.BookTitle = bookTitle.String
	metadata.SeriesTitle = seriesTitle.String
	metadata.EventTitle = eventTitle.String
	metadata.EventLocation = eventLocation.String
	metadata.Institution = institution.String
	metadata.Pages = pages.String
	metadata.Volume = volume.String
	metadata.Issue = issue.String
	metadata.License = license.String
	metadata.Copyright = copyright.String
	metadata.Funding = funding.String
	metadata.DataSource = dataSource.String
	metadata.DataSourceTimestamp = formatTimestamp(dataSourceTimestamp)

	return uuid, metadata, nil
}

func insertMetadata(ctx context.Context, tx *sql.Tx, uuid string, metadata domain.Metadata) error {
	timestamp, err := timestampValue(metadata.DataSourceTimestamp)
	if err != nil {
		return fmt.Errorf("parse data source timestamp: %w", err)
	}
	_, err = tx.ExecContext(ctx, `
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
		pgtype.FlatArray[string](metadata.ISBN),
		pgtype.FlatArray[string](metadata.ISSN),
		pgtype.FlatArray[string](metadata.References),
		nullString(metadata.License),
		nullString(metadata.Copyright),
		nullString(metadata.Funding),
		nullString(metadata.DataSource),
		timestamp,
	)
	return err
}
