package service

import (
	"encoding/base64"
	"encoding/json"
	"strings"

	"entgo.io/ent/dialect/sql"
	"github.com/google/uuid"
	"github.com/ndscm/theseed/seed/infra/error/go/seederr"
	"github.com/ndscm/theseed/seed/infra/grpc/go/aips"
	"github.com/ndscm/theseed/seed/newtype/steins/database/ent"
	"github.com/ndscm/theseed/seed/newtype/steins/database/ent/brainstep"
)

const defaultPageSize = 25
const maxPageSize = 100

// BrainStepPageToken is a cursor for AIP-158 pagination. It stores only the
// primary key of the last returned row; the sort order and filter are
// re-supplied by the caller via order_by and filter on the next request.
type BrainStepPageToken struct {
	Uuid string `json:"uuid"`
}

// encodeBrainStepPageToken creates a base64-encoded cursor from the last step's UUID.
func encodeBrainStepPageToken(stepUuid uuid.UUID) string {
	cursor := BrainStepPageToken{
		Uuid: stepUuid.String(),
	}
	jsonBytes, err := json.Marshal(cursor)
	if err != nil {
		return ""
	}
	return base64.URLEncoding.EncodeToString(jsonBytes)
}

// decodeBrainStepPageToken parses a base64-encoded cursor string into a step UUID.
func decodeBrainStepPageToken(pageToken string) (uuid.UUID, error) {
	jsonBytes, err := base64.URLEncoding.DecodeString(pageToken)
	if err != nil {
		return uuid.UUID{}, seederr.WrapErrorf("invalid page_token encoding: %w", err)
	}
	cursor := BrainStepPageToken{}
	err = json.Unmarshal(jsonBytes, &cursor)
	if err != nil {
		return uuid.UUID{}, seederr.WrapErrorf("invalid page_token format: %w", err)
	}
	stepUuid, err := uuid.Parse(cursor.Uuid)
	if err != nil {
		return uuid.UUID{}, seederr.WrapErrorf("invalid page_token id: %w", err)
	}
	return stepUuid, nil
}

// parseBrainStepOrderBy parses an AIP-132 order_by string (e.g.
// "emit_time desc, type asc") into an ordered list of field orders. When no
// order is given it defaults to newest-first by emit_time. A tiebreak on the
// primary key is always appended so the total order is stable and keyset
// pagination is well defined.
func parseBrainStepOrderBy(orderBy string) ([]aips.FieldOrder, error) {
	orders := []aips.FieldOrder{}
	for _, field := range strings.Split(orderBy, ",") {
		field = strings.TrimSpace(field)
		if field == "" {
			continue
		}
		parts := strings.Fields(field)
		fieldName := parts[0]
		isDesc := false
		if len(parts) > 1 {
			direction := strings.ToLower(parts[1])
			switch direction {
			case "asc":
				// pass
			case "desc":
				isDesc = true
			default:
				return nil, seederr.WrapErrorf("invalid order direction: %s", parts[1])
			}
		}
		switch fieldName {
		case brainstep.FieldID:
		case brainstep.FieldCreateTime:
		case brainstep.FieldEmitTime:
		case brainstep.FieldType:
		case brainstep.FieldTopic:
		case brainstep.FieldThreadUUID:
		default:
			return nil, seederr.WrapErrorf("unknown order_by field: %s", fieldName)
		}
		orders = append(orders, aips.FieldOrder{Field: fieldName, Desc: isDesc})
	}
	if len(orders) == 0 {
		orders = append(orders, aips.FieldOrder{Field: brainstep.FieldID})
	}
	return orders, nil
}

func buildBrainStepEqualityPredicate(row *ent.BrainStep, field string) *sql.Predicate {
	switch field {
	case "uuid":
		return sql.EQ(brainstep.FieldID, row.ID)
	case "create_time":
		return sql.EQ(brainstep.FieldCreateTime, row.CreateTime)
	case "emit_time":
		return sql.EQ(brainstep.FieldEmitTime, row.EmitTime)
	case "type":
		return sql.EQ(brainstep.FieldType, row.Type)
	case "topic":
		return sql.EQ(brainstep.FieldTopic, row.Topic)
	case "thread_uuid":
		return sql.EQ(brainstep.FieldThreadUUID, row.ThreadUUID)
	}
	return nil
}

func buildBrainStepComparisonPredicate(row *ent.BrainStep, field string, desc bool) *sql.Predicate {
	switch field {
	case "uuid":
		if desc {
			return sql.LT(brainstep.FieldID, row.ID)
		}
		return sql.GT(brainstep.FieldID, row.ID)
	case "create_time":
		if desc {
			return sql.LT(brainstep.FieldCreateTime, row.CreateTime)
		}
		return sql.GT(brainstep.FieldCreateTime, row.CreateTime)
	case "emit_time":
		if desc {
			return sql.LT(brainstep.FieldEmitTime, row.EmitTime)
		}
		return sql.GT(brainstep.FieldEmitTime, row.EmitTime)
	case "type":
		if desc {
			return sql.LT(brainstep.FieldType, row.Type)
		}
		return sql.GT(brainstep.FieldType, row.Type)
	case "topic":
		if desc {
			return sql.LT(brainstep.FieldTopic, row.Topic)
		}
		return sql.GT(brainstep.FieldTopic, row.Topic)
	case "thread_uuid":
		if desc {
			return sql.LT(brainstep.FieldThreadUUID, row.ThreadUUID)
		}
		return sql.GT(brainstep.FieldThreadUUID, row.ThreadUUID)
	}
	return nil
}

// buildBrainStepCursorPredicate builds the keyset-pagination predicate that
// selects rows strictly after cursorRow under the given order. For
// ORDER BY col1 ASC, col2 DESC it generates:
//
//	(col1 > c1) OR (col1 = c1 AND col2 < c2)
func buildBrainStepCursorPredicate(cursorRow *ent.BrainStep, orders []aips.FieldOrder) *sql.Predicate {
	if cursorRow == nil || len(orders) == 0 {
		return nil
	}
	orPredicates := []*sql.Predicate{}
	for i := range orders {
		andPredicates := []*sql.Predicate{}
		for j := 0; j < i; j++ {
			andPredicates = append(andPredicates, buildBrainStepEqualityPredicate(cursorRow, orders[j].Field))
		}
		andPredicates = append(andPredicates, buildBrainStepComparisonPredicate(cursorRow, orders[i].Field, orders[i].Desc))
		orPredicates = append(orPredicates, sql.And(andPredicates...))
	}
	return sql.Or(orPredicates...)
}
