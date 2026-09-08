package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	_ "github.com/jackc/pgx/v5/stdlib"
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

var pgTypes = pgtype.NewMap()

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

	args, err := paperArguments(paper)
	if err != nil {
		return wrapError("prepare paper update", err)
	}
	args = append(args, uuid)
	result, err := tx.ExecContext(ctx, `
		UPDATE public.paper
		SET doi = $1,
			title = $2,
			title_short = $3,
			publication_year = $4,
			publication_month = $5,
			publication_day = $6,
			paper_type = $7,
			publication_status = $8,
			publication_status_timestamp = $9,
			abstract = $10,
			keywords = $11,
			pdf_url = $12
		WHERE uuid = $13`, args...)
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
		return err
	}
	return loadAuthors(ctx, r.db, byUUID, uuids)
}

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
	var paper domain.Paper
	var doi, title, titleShort, paperType, publicationStatus, abstract, pdfURL sql.NullString
	var year, month, day sql.NullInt16
	var statusTimestamp sql.NullTime

	err := row.Scan(
		&paper.UUID,
		&doi,
		&title,
		&titleShort,
		&year,
		&month,
		&day,
		&paperType,
		&publicationStatus,
		&statusTimestamp,
		&abstract,
		pgTypes.SQLScanner(&paper.Keywords),
		&pdfURL,
	)
	if err != nil {
		return domain.Paper{}, err
	}

	paper.DOI = doi.String
	paper.Title = title.String
	paper.TitleShort = titleShort.String
	paper.PublicationDate = domain.Date{Year: int(year.Int16), Month: int(month.Int16), Day: int(day.Int16)}
	paper.Type = domain.PublicationType(paperType.String)
	paper.PublicationStatus = publicationStatus.String
	paper.PublicationStatusTimestamp = formatTimestamp(statusTimestamp)
	paper.Abstract = abstract.String
	paper.PDFURL = pdfURL.String

	return paper, nil
}

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

func loadAuthors(ctx context.Context, db *sql.DB, papers map[string]*domain.Paper, uuids []string) error {
	rows, err := db.QueryContext(ctx, `
		SELECT pa.uuid_paper::text, a.name_first, a.name_middle, a.name_last, af.name, a.orcid
		FROM public.paper_author pa
		JOIN public.author a ON a.key = pa.key_author
		LEFT JOIN public.affiliation af ON af.key = a.key_affiliation
		WHERE pa.uuid_paper = ANY($1::uuid[])
		ORDER BY pa.uuid_paper ASC, pa.position ASC`, pgtype.FlatArray[string](uuids))
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var uuid string
		var first, middle, last, affiliation, orcid sql.NullString
		if err := rows.Scan(&uuid, &first, &middle, &last, &affiliation, &orcid); err != nil {
			return err
		}
		if paper := papers[uuid]; paper != nil {
			paper.Authors = append(paper.Authors, domain.Author{
				NameFirst:   first.String,
				NameMiddle:  middle.String,
				NameLast:    last.String,
				Affiliation: affiliation.String,
				ORCID:       orcid.String,
			})
		}
	}

	return rows.Err()
}

func insertPaper(ctx context.Context, tx *sql.Tx, paper domain.Paper) error {
	args, err := paperArguments(paper)
	if err != nil {
		return err
	}
	args = append([]any{paper.UUID}, args...)
	_, err = tx.ExecContext(ctx, `
		INSERT INTO public.paper (
			uuid, doi, title, title_short, publication_year, publication_month, publication_day,
			paper_type, publication_status, publication_status_timestamp, abstract, keywords, pdf_url
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)`, args...)
	return err
}

func paperArguments(paper domain.Paper) ([]any, error) {
	statusTimestamp, err := timestampValue(paper.PublicationStatusTimestamp)
	if err != nil {
		return nil, err
	}

	return []any{
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

func insertChildren(ctx context.Context, tx *sql.Tx, paper domain.Paper) error {
	if err := insertMetadata(ctx, tx, paper.UUID, paper.Metadata); err != nil {
		return err
	}
	return insertAuthors(ctx, tx, paper.UUID, paper.Authors)
}

func insertMetadata(ctx context.Context, tx *sql.Tx, uuid string, metadata domain.Metadata) error {
	timestamp, err := timestampValue(metadata.DataSourceTimestamp)
	if err != nil {
		return err
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

func insertAuthors(ctx context.Context, tx *sql.Tx, uuid string, authors []domain.Author) error {
	for position, author := range authors {
		var affiliation any
		if author.Affiliation != "" {
			if err := tx.QueryRowContext(ctx, `INSERT INTO public.affiliation (name) VALUES ($1) RETURNING key`, author.Affiliation).Scan(&affiliation); err != nil {
				return err
			}
		}

		var authorKey int64
		if err := tx.QueryRowContext(ctx, `
			INSERT INTO public.author (name_first, name_middle, name_last, orcid, key_affiliation)
			VALUES ($1, $2, $3, $4, $5)
			RETURNING key`,
			nullString(author.NameFirst),
			nullString(author.NameMiddle),
			nullString(author.NameLast),
			nullString(author.ORCID),
			affiliation,
		).Scan(&authorKey); err != nil {
			return err
		}
		if _, err := tx.ExecContext(ctx, `INSERT INTO public.paper_author (uuid_paper, key_author, position) VALUES ($1, $2, $3)`, uuid, authorKey, position); err != nil {
			return err
		}
	}
	return nil
}

func removePaperAuthors(ctx context.Context, tx *sql.Tx, uuid string) error {
	_, err := tx.ExecContext(ctx, `
		WITH deleted_links AS (
			DELETE FROM public.paper_author
			WHERE uuid_paper = $1
			RETURNING key_author
		), deleted_authors AS (
			DELETE FROM public.author
			WHERE key IN (SELECT key_author FROM deleted_links)
				AND NOT EXISTS (
					SELECT 1 FROM public.paper_author
					WHERE key_author = author.key
				)
			RETURNING key_affiliation
		)
		DELETE FROM public.affiliation
		WHERE key IN (SELECT key_affiliation FROM deleted_authors)
			AND NOT EXISTS (
				SELECT 1 FROM public.author
				WHERE key_affiliation = affiliation.key
			)`, uuid)
	return err
}

func timestampValue(value string) (any, error) {
	if value == "" {
		return nil, nil
	}
	parsed, err := time.Parse(time.RFC3339, value)
	if err != nil {
		return nil, err
	}
	return parsed, nil
}

func formatTimestamp(value sql.NullTime) string {
	if !value.Valid {
		return ""
	}
	return value.Time.UTC().Format(time.RFC3339Nano)
}

func nullString(value string) any {
	if value == "" {
		return nil
	}
	return value
}

func nullInt(value int) any {
	if value == 0 {
		return nil
	}
	return value
}

func wrapError(operation string, err error) error {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23505" {
		return fmt.Errorf("%s: %w", operation, domain.ErrPaperAlreadyExists)
	}
	return fmt.Errorf("%s: %w", operation, err)
}
