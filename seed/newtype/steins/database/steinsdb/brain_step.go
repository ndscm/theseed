package steinsdb

import (
	"context"

	"entgo.io/ent/dialect/sql"
	"github.com/ndscm/theseed/seed/infra/error/go/seederr"
	"github.com/ndscm/theseed/seed/infra/grpc/go/aips"
	"github.com/ndscm/theseed/seed/newtype/steins/database/ent"
)

// SelectBrainStepRows returns the BrainStep rows matching filters, ordered by
// orders and restricted to the keyset window after cursor, capped at limit.
// It also returns the total number of rows matching filters, ignoring the
// cursor and limit (i.e. the full result-set size for pagination metadata).
func SelectBrainStepRows(
	ctx context.Context, db *ent.Client,
	filters []*sql.Predicate, orders []aips.FieldOrder,
	cursor *sql.Predicate, limit int,
) ([]*ent.BrainStep, int64, error) {
	query := db.BrainStep.Query()
	for _, filter := range filters {
		if filter != nil {
			query = query.Where(func(s *sql.Selector) {
				s.Where(filter)
			})
		}
	}

	total, err := query.Count(ctx)
	if err != nil {
		return nil, 0, seederr.Wrap(err)
	}

	for _, order := range orders {
		query = query.Order(func(s *sql.Selector) {
			if order.Desc {
				s.OrderBy(sql.Desc(order.Field))
			} else {
				s.OrderBy(sql.Asc(order.Field))
			}
		})
	}
	if cursor != nil {
		query = query.Where(func(s *sql.Selector) {
			s.Where(cursor)
		})
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
