package postgres

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/paperstacks.io/paperstacks/internal/paper/domain"
)

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

func insertAuthors(ctx context.Context, tx *sql.Tx, uuid string, authors []domain.Author) error {
	for position, author := range authors {
		var affiliation any
		if author.Affiliation != "" {
			if err := tx.QueryRowContext(ctx, `INSERT INTO public.affiliation (name) VALUES ($1) RETURNING key`, author.Affiliation).Scan(&affiliation); err != nil {
				return fmt.Errorf("insert affiliation for author %d: %w", position, err)
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
			return fmt.Errorf("insert author %d: %w", position, err)
		}
		if _, err := tx.ExecContext(ctx, `INSERT INTO public.paper_author (uuid_paper, key_author, position) VALUES ($1, $2, $3)`, uuid, authorKey, position); err != nil {
			return fmt.Errorf("link author %d: %w", position, err)
		}
	}
	return nil
}

func removePaperAuthors(ctx context.Context, tx *sql.Tx, uuid string) error {
	pgTypes := pgtype.NewMap()
	var authorKeys, affiliationKeys []int64

	// Use separate statements so each orphan check sees the preceding deletes.
	if err := tx.QueryRowContext(ctx, `
		WITH deleted_links AS (
			DELETE FROM public.paper_author WHERE uuid_paper = $1
			RETURNING key_author
		)
		SELECT array_agg(key_author) FROM deleted_links`, uuid).Scan(pgTypes.SQLScanner(&authorKeys)); err != nil {
		return fmt.Errorf("delete author links: %w", err)
	}
	if err := tx.QueryRowContext(ctx, `
		WITH deleted_authors AS (
			DELETE FROM public.author
			WHERE key = ANY($1::bigint[])
				AND NOT EXISTS (
					SELECT 1 FROM public.paper_author WHERE key_author = author.key
				)
			RETURNING key_affiliation
		)
		SELECT array_agg(key_affiliation) FILTER (WHERE key_affiliation IS NOT NULL)
		FROM deleted_authors`, pgtype.FlatArray[int64](authorKeys)).Scan(pgTypes.SQLScanner(&affiliationKeys)); err != nil {
		return fmt.Errorf("delete unused authors: %w", err)
	}
	if _, err := tx.ExecContext(ctx, `
		DELETE FROM public.affiliation
		WHERE key = ANY($1::bigint[])
			AND NOT EXISTS (
				SELECT 1 FROM public.author WHERE key_affiliation = affiliation.key
			)`, pgtype.FlatArray[int64](affiliationKeys)); err != nil {
		return fmt.Errorf("delete unused affiliations: %w", err)
	}
	return nil
}
