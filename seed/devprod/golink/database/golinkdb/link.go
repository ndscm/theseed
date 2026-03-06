package golinkdb

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/ndscm/theseed/seed/infra/error/go/seederr"
	"github.com/ndscm/theseed/seed/infra/log/go/seedlog"
)

type LinkRow struct {
	Key         string
	Target      string
	Public      bool
	Owner       *string
	HitCount    int64
	CreatedTime time.Time
	UpdatedTime time.Time
}

func InsertLink(ctx context.Context, db *sql.DB, link *LinkRow) (*LinkRow, error) {
	if link == nil {
		return nil, seederr.WrapErrorf("link is required")
	}
	row := LinkRow{}
	err := db.QueryRowContext(ctx, `
INSERT INTO "golink" (
  "key",
  "target",
  "public",
  "owner"
) VALUES ($1, $2, $3, $4)
RETURNING "key",
  "target",
  "public",
  "owner",
  "hit_count",
  "created_time",
  "updated_time";
`,
		link.Key,
		link.Target,
		link.Public,
		link.Owner,
	).Scan(
		&row.Key,
		&row.Target,
		&row.Public,
		&row.Owner,
		&row.HitCount,
		&row.CreatedTime,
		&row.UpdatedTime,
	)
	if err != nil {
		return nil, seederr.Wrap(err)
	}

	return &row, nil
}

func SelectLinkByKey(ctx context.Context, db *sql.DB, key string) (*LinkRow, error) {
	row := LinkRow{}
	err := db.QueryRowContext(ctx, `
SELECT "key",
  "target",
  "public",
  "owner",
  "hit_count",
  "created_time",
  "updated_time"
FROM "golink"
WHERE "key" = $1;
`, key).Scan(
		&row.Key,
		&row.Target,
		&row.Public,
		&row.Owner,
		&row.HitCount,
		&row.CreatedTime,
		&row.UpdatedTime,
	)
	if err != nil {
		return nil, seederr.Wrap(err)
	}

	return &row, nil
}

func UpdateLink(ctx context.Context, db *sql.DB, key string, link *LinkRow, updateFields []string) (*LinkRow, error) {
	if link == nil {
		return nil, seederr.WrapErrorf("link is required")
	}
	setSegments := []string{}
	values := []interface{}{key}
	for _, field := range updateFields {
		switch field {
		case "target":
			setSegments = append(setSegments, fmt.Sprintf(`"target" = $%d`, len(values)+1))
			values = append(values, link.Target)
		case "public":
			setSegments = append(setSegments, fmt.Sprintf(`"public" = $%d`, len(values)+1))
			values = append(values, link.Public)
		case "owner":
			setSegments = append(setSegments, fmt.Sprintf(`"owner" = $%d`, len(values)+1))
			values = append(values, link.Owner)
		default:
			seedlog.Warnf("Unknown field in update fields: %s", field)
		}
	}
	setSegments = append(setSegments, `"updated_time" = CURRENT_TIMESTAMP`)
	row := LinkRow{}
	err := db.QueryRowContext(ctx, `
UPDATE "golink"
SET `+strings.Join(setSegments, ", ")+`
WHERE "key" = $1
RETURNING "key",
  "target",
  "public",
  "owner",
  "hit_count",
  "created_time",
  "updated_time";
`, values...,
	).Scan(
		&row.Key,
		&row.Target,
		&row.Public,
		&row.Owner,
		&row.HitCount,
		&row.CreatedTime,
		&row.UpdatedTime,
	)
	if err != nil {
		return nil, seederr.Wrap(err)
	}

	return &row, nil
}

func DeleteLink(ctx context.Context, db *sql.DB, key string) error {
	result, err := db.ExecContext(ctx, `
DELETE FROM "golink" WHERE "key" = $1;
`, key)
	if err != nil {
		return seederr.Wrap(err)
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return seederr.Wrap(err)
	}
	if rowsAffected == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func SelectLinks(ctx context.Context, db *sql.DB, cursor string, limit int) ([]*LinkRow, error) {
	var rows *sql.Rows
	var err error

	whereClause := ""
	values := []interface{}{limit}
	if cursor != "" {
		whereClause = `WHERE "key" > $2`
		values = append(values, cursor)
	}
	rows, err = db.QueryContext(ctx, `
SELECT "key",
  "target",
  "public",
  "owner",
  "hit_count",
  "created_time",
  "updated_time"
FROM "golink"
`+whereClause+`
ORDER BY "key" ASC
LIMIT $1;
`, values...)
	if err != nil {
		return nil, seederr.Wrap(err)
	}
	defer rows.Close()

	var links []*LinkRow
	for rows.Next() {
		row := LinkRow{}
		err := rows.Scan(
			&row.Key,
			&row.Target,
			&row.Public,
			&row.Owner,
			&row.HitCount,
			&row.CreatedTime,
			&row.UpdatedTime,
		)
		if err != nil {
			return nil, seederr.Wrap(err)
		}
		links = append(links, &row)
	}

	return links, nil
}

func IncrementHitCount(ctx context.Context, db *sql.DB, key string) error {
	_, err := db.ExecContext(ctx, `
UPDATE "golink"
SET "hit_count" = "hit_count" + 1
WHERE "key" = $1;
`, key)
	if err != nil {
		return seederr.Wrap(err)
	}
	return nil
}

func CountLinks(ctx context.Context, db *sql.DB) (int64, error) {
	var count int64
	err := db.QueryRowContext(ctx, `
SELECT COUNT(*) FROM "golink";
`).Scan(&count)
	if err != nil {
		return 0, seederr.Wrap(err)
	}
	return count, nil
}
