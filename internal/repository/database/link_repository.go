package database

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"short_link/internal/model"

	"github.com/lib/pq"
)

type LinkRepository struct {
	db *sql.DB
}

func NewLinkRepository(db *sql.DB) *LinkRepository {
	return &LinkRepository{db: db}
}

func (r *LinkRepository) BeginTx(ctx context.Context) (*sql.Tx, error) {
	return r.db.BeginTx(ctx, &sql.TxOptions{
		Isolation: sql.LevelReadCommitted,
		ReadOnly:  false,
	})
}

func (r *LinkRepository) CreateOriginalLink(ctx context.Context, tx *sql.Tx, originalUrl string) (*model.Link, error) {
	query := `INSERT INTO links (url) 
	VALUES ($1) 
	ON CONFLICT (url) DO UPDATE SET url = EXCLUDED.url
	RETURNING id, url`

	var link model.Link

	err := tx.QueryRowContext(ctx, query, originalUrl).Scan(&link.Id, &link.URL)

	if err != nil {
		return nil, fmt.Errorf("Error when adding url %s: %w", originalUrl, err)
	}

	return &link, nil
}

func (r *LinkRepository) ExistsOriginalLink(ctx context.Context, tx *sql.Tx, originalUrl string) (*model.Link, error) {
	query := `SELECT id, url 
			FROM links 
			WHERE links.url = $1`

	var link model.Link

	err := tx.QueryRowContext(ctx, query, originalUrl).Scan(&link.Id, &link.URL)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("check URL '%s': %w", originalUrl, err)
	}

	return &link, nil
}

func (r *LinkRepository) CreateShortLink(ctx context.Context, tx *sql.Tx, shortUrl string, id_url int64) (*model.ShortLink, error) {
	query := `INSERT INTO short_links (id_url, short_url)
	VALUES ($1, $2)
	RETURNING id, id_url, short_url, created_at, accessed_at, accessed_count`

	var shortLink model.ShortLink

	err := tx.QueryRowContext(ctx, query, id_url, shortUrl).Scan(
		&shortLink.Id, &shortLink.IdURL, &shortLink.ShortURL, &shortLink.CreatedAt, &shortLink.AccessedAt, &shortLink.AccessedCount)

	if err != nil {
		return nil, fmt.Errorf("Error when adding short url %s: %w", shortUrl, err)
	}

	return &shortLink, nil
}

func (r *LinkRepository) ExistsShortLink(ctx context.Context, tx *sql.Tx, shortUrl string) (*model.ShortLink, error) {
	query := `SELECT id, short_url 
			FROM short_links 
			WHERE short_links.short_url = $1`

	var shortLink model.ShortLink

	err := tx.QueryRowContext(ctx, query, shortUrl).Scan(&shortLink.Id, &shortLink.ShortURL)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("check URL '%s': %w", shortUrl, err)
	}

	return &shortLink, nil
}

func (r *LinkRepository) GetOriginalLink(ctx context.Context, tx *sql.Tx, shortUrl string) (string, error) {
	query := `WITH updated AS (
        UPDATE short_links 
        SET accessed_at = CURRENT_TIMESTAMP, 
            accessed_count = accessed_count + 1
        WHERE short_url = $1
        RETURNING id_url
    )
    SELECT l.url
    FROM links l
    WHERE l.id = (SELECT id_url FROM updated)`

	var link string

	err := tx.QueryRowContext(ctx, query, shortUrl).Scan(&link)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", nil
		}
		return "", fmt.Errorf("failed to get original link for short URL '%s': %w", shortUrl, err)
	}
	return link, nil
}

func (r *LinkRepository) AccessedCountIncrement(ctx context.Context, tx *sql.Tx, shortURL string) error {
	query := `UPDATE short_links
			SET accessed_count = accessed_count + 1,
			accessed_at = CURRENT_TIMESTAMP
			WHERE short_url = $1`

	_, err := tx.ExecContext(ctx, query, shortURL)

	if err != nil {
		return fmt.Errorf(err.Error())
	}

	return nil
}

func (r *LinkRepository) GetAllShortLink(ctx context.Context, tx *sql.Tx) ([]model.LinkStatsDTO, error) {
	query := `SELECT l.url, sl.short_url, sl.created_at, sl.accessed_at, sl.accessed_count
			FROM short_links AS sl
			INNER JOIN links AS l 
			ON sl.id_url = l.id
			ORDER BY sl.created_at DESC`

	var linkStats []model.LinkStatsDTO

	rows, err := tx.QueryContext(ctx, query)

	if err != nil {
		return nil, fmt.Errorf("Error: method get all links: %w", err)
	}

	defer rows.Close()

	for rows.Next() {
		var linkStat model.LinkStatsDTO
		err := rows.Scan(
			&linkStat.URL,
			&linkStat.ShortURL,
			&linkStat.CreatedAt,
			&linkStat.AccessedAt,
			&linkStat.AccessedCount,
		)

		if err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}

		linkStats = append(linkStats, linkStat)
	}

	return linkStats, nil
}

func (r *LinkRepository) FindExpiredLinks(ctx context.Context, batchSize int, tx *sql.Tx) ([]string, error) {
	query := `SELECT short_url
			FROM short_links
			WHERE accessed_at < CURRENT_TIMESTAMP - INTERVAL '24 hour'
			LIMIT ($1)`

	rows, err := tx.QueryContext(ctx, query, batchSize)

	if err != nil {
		return nil, fmt.Errorf("Request execution error: %w", err)
	}

	defer rows.Close()

	var expiredLinks []string

	for rows.Next() {
		var expiredLink string

		err := rows.Scan(&expiredLink)

		if err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}

		expiredLinks = append(expiredLinks, expiredLink)
	}

	return expiredLinks, nil
}

func (r *LinkRepository) DeleteExpiredShortLinks(ctx context.Context, expLinks []string, tx *sql.Tx) error {
	query := `DELETE FROM short_links
			WHERE short_url = ANY($1)
			RETURNING id_url, short_url`

	_, err := tx.ExecContext(ctx, query, pq.Array(expLinks))

	if err != nil {
		return fmt.Errorf("Request execution error: %w", err)
	}

	return nil
}

func (r *LinkRepository) DeleteExpiredOriginalLinks(ctx context.Context, tx *sql.Tx) error {
	query := `DELETE FROM links
			WHERE id NOT IN(
				SELECT id_url
				FROM short_links
				WHERE id_url IS NOT NULL)`

	_, err := tx.ExecContext(ctx, query)

	if err != nil {
		return fmt.Errorf("Request execution error: %w", err)
	}

	return nil
}
