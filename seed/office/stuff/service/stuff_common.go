package service

import (
	"github.com/ndscm/theseed/seed/office/stuff/database/ent"
	"github.com/ndscm/theseed/seed/office/stuff/proto/stuffpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func getStuffProtoFromEnt(row *ent.Stuff) *stuffpb.Stuff {
	if row == nil {
		return nil
	}
	s := &stuffpb.Stuff{
		Uuid:        row.ID.String(),
		Order:       row.Order,
		Owner:       row.Owner,
		CreatedTime: timestamppb.New(row.CreatedTime),
		UpdatedTime: timestamppb.New(row.UpdatedTime),
	}
	if row.Data != nil {
		s.Data = string(*row.Data)
	}
	return s
}

func getCategoryProtoFromEnt(row *ent.Category) *stuffpb.Category {
	if row == nil {
		return nil
	}
	c := &stuffpb.Category{
		Uuid: row.ID.String(),
	}
	if row.Primary != nil {
		c.Primary = *row.Primary
	}
	if row.Secondary != nil {
		c.Secondary = *row.Secondary
	}
	return c
}
