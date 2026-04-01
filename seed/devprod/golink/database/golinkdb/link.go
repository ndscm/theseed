package golinkdb

import (
	"context"

	"entgo.io/ent/dialect/sql"
	"github.com/ndscm/theseed/seed/devprod/golink/database/ent"
	"github.com/ndscm/theseed/seed/devprod/golink/database/ent/link"
	"github.com/ndscm/theseed/seed/devprod/golink/proto/golinkerrorpb"
	"github.com/ndscm/theseed/seed/infra/error/go/seederr"
	"github.com/ndscm/theseed/seed/infra/log/go/seedlog"
)

func InsertLink(ctx context.Context, db *ent.Client, key string, row *ent.Link) (*ent.Link, error) {
	if row == nil {
		return nil, seederr.CodeErrorf(golinkerrorpb.Code_INVALID_LINK, "link is required")
	}
	createQuery := db.Link.Create().
		SetID(key).
		SetTarget(row.Target).
		SetPublic(row.Public)
	if row.Owner != nil {
		createQuery.SetOwner(*row.Owner)
	}
	resultRow, err := createQuery.Save(ctx)
	if err != nil {
		return nil, seederr.Wrap(err)
	}
	return resultRow, nil
}

func SelectLink(ctx context.Context, db *ent.Client, key string) (*ent.Link, error) {
	row, err := db.Link.Get(ctx, key)
	if err != nil {
		return nil, seederr.Wrap(err)
	}
	return row, nil
}

func UpdateLink(ctx context.Context, db *ent.Client, key string, row *ent.Link, updateFields []string) (*ent.Link, error) {
	if row == nil {
		return nil, seederr.WrapErrorf("link is required")
	}
	updateQuery := db.Link.UpdateOneID(key)
	for _, field := range updateFields {
		switch field {
		case "target":
			updateQuery.SetTarget(row.Target)
		case "public":
			updateQuery.SetPublic(row.Public)
		case "owner":
			if row.Owner == nil {
				updateQuery.ClearOwner()
			} else {
				updateQuery.SetOwner(*row.Owner)
			}
		default:
			seedlog.Warnf("Unknown field in update fields: %s", field)
		}
	}
	resultRow, err := updateQuery.Save(ctx)
	if err != nil {
		return nil, seederr.Wrap(err)
	}
	return resultRow, nil
}

func DeleteLink(ctx context.Context, db *ent.Client, key string) error {
	err := db.Link.DeleteOneID(key).Exec(ctx)
	if err != nil {
		return seederr.Wrap(err)
	}
	return nil
}

func SelectLinks(ctx context.Context, db *ent.Client, cursor string, limit int) ([]*ent.Link, int64, error) {
	query := db.Link.Query().Order(link.ByID(sql.OrderAsc()))
	total, err := query.Clone().Count(ctx)
	if err != nil {
		return nil, 0, seederr.Wrap(err)
	}
	if cursor != "" {
		query = query.Where(link.IDGT(cursor))
	}
	if limit > 0 {
		query = query.Limit(limit)
	}
	rows, err := query.All(ctx)
	if err != nil {
		return nil, 0, seederr.Wrap(err)
	}
	return rows, int64(total), nil
}

func IncrementHitCount(ctx context.Context, db *ent.Client, key string) error {
	err := db.Link.UpdateOneID(key).AddHitCount(1).Exec(ctx)
	if err != nil {
		return seederr.Wrap(err)
	}
	return nil
}
