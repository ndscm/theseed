package service

import (
	"context"
	"encoding/json"

	"connectrpc.com/connect"
	"github.com/google/uuid"
	"github.com/ndscm/theseed/seed/cloud/login/go/login"
	"github.com/ndscm/theseed/seed/infra/error/go/seederr"
	"github.com/ndscm/theseed/seed/infra/log/go/seedlog"
	"github.com/ndscm/theseed/seed/office/stuff/database/ent"
	"github.com/ndscm/theseed/seed/office/stuff/database/stuffdb"
	"github.com/ndscm/theseed/seed/office/stuff/proto/stufferrorpb"
	"github.com/ndscm/theseed/seed/office/stuff/proto/stuffpb"
	"google.golang.org/grpc/codes"
	"google.golang.org/protobuf/types/known/emptypb"
)

type StuffService struct{}

func (svc *StuffService) ListCategories(
	ctx context.Context,
	req *connect.Request[stuffpb.ListCategoriesRequest],
) (*connect.Response[stuffpb.ListCategoriesResponse], error) {
	_, err := login.EnsureLoginUser(ctx)
	if err != nil {
		return nil, seederr.DefaultCode(codes.Unauthenticated, err)
	}
	db, err := stuffdb.Open(ctx)
	if err != nil {
		return nil, seederr.DefaultCode(stufferrorpb.Code_INTERNAL_OPEN_DATABASE_FAILED, err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			seedlog.Errorf("close db failed: %v", err)
		}
	}()

	rows, err := db.Category.Query().All(ctx)
	if err != nil {
		return nil, seederr.DefaultCode(codes.Unknown, err)
	}

	categories := make([]*stuffpb.Category, 0, len(rows))
	for _, row := range rows {
		categories = append(categories, getCategoryProtoFromEnt(row))
	}

	return connect.NewResponse(&stuffpb.ListCategoriesResponse{
		Categories: categories,
	}), nil
}

func (svc *StuffService) CreateStuff(
	ctx context.Context,
	req *connect.Request[stuffpb.CreateStuffRequest],
) (*connect.Response[stuffpb.Stuff], error) {
	_, err := login.EnsureLoginUser(ctx)
	if err != nil {
		return nil, seederr.DefaultCode(codes.Unauthenticated, err)
	}
	if req.Msg.Stuff == nil {
		return nil, seederr.CodeErrorf(stufferrorpb.Code_INVALID_STUFF, "stuff is required")
	}

	db, err := stuffdb.Open(ctx)
	if err != nil {
		return nil, seederr.DefaultCode(stufferrorpb.Code_INTERNAL_OPEN_DATABASE_FAILED, err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			seedlog.Errorf("close db failed: %v", err)
		}
	}()

	create := db.Stuff.Create()
	create.SetOrder(req.Msg.Stuff.Order)
	if req.Msg.Stuff.Data != "" {
		raw := json.RawMessage(req.Msg.Stuff.Data)
		create.SetData(&raw)
	}
	loginUser, err := login.LoginUser(ctx)
	if err == nil && loginUser != nil && loginUser.Email != "" {
		create.SetOwner(loginUser.Email)
	}

	row, err := create.Save(ctx)
	if err != nil {
		return nil, seederr.DefaultCode(codes.Unknown, err)
	}

	return connect.NewResponse(getStuffProtoFromEnt(row)), nil
}

func (svc *StuffService) GetStuff(
	ctx context.Context,
	req *connect.Request[stuffpb.GetStuffRequest],
) (*connect.Response[stuffpb.Stuff], error) {
	_, err := login.EnsureLoginUser(ctx)
	if err != nil {
		return nil, seederr.DefaultCode(codes.Unauthenticated, err)
	}
	id, err := uuid.Parse(req.Msg.Uuid)
	if err != nil {
		return nil, seederr.CodeErrorf(stufferrorpb.Code_INVALID_STUFF_ID, "stuff id is invalid")
	}

	db, err := stuffdb.Open(ctx)
	if err != nil {
		return nil, seederr.DefaultCode(stufferrorpb.Code_INTERNAL_OPEN_DATABASE_FAILED, err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			seedlog.Errorf("close db failed: %v", err)
		}
	}()

	row, err := db.Stuff.Get(ctx, id)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, seederr.CodeErrorf(stufferrorpb.Code_NOTFOUND_STUFF, "stuff not found: %s", id)
		}
		return nil, seederr.DefaultCode(codes.Unknown, err)
	}

	return connect.NewResponse(getStuffProtoFromEnt(row)), nil
}

func (svc *StuffService) UpdateStuff(
	ctx context.Context,
	req *connect.Request[stuffpb.UpdateStuffRequest],
) (*connect.Response[stuffpb.Stuff], error) {
	_, err := login.EnsureLoginUser(ctx)
	if err != nil {
		return nil, seederr.DefaultCode(codes.Unauthenticated, err)
	}
	if req.Msg.Stuff == nil {
		return nil, seederr.CodeErrorf(stufferrorpb.Code_INVALID_STUFF, "stuff is required")
	}
	id, err := uuid.Parse(req.Msg.Stuff.Uuid)
	if err != nil {
		return nil, seederr.CodeErrorf(stufferrorpb.Code_INVALID_STUFF_ID, "stuff id is invalid")
	}

	db, err := stuffdb.Open(ctx)
	if err != nil {
		return nil, seederr.DefaultCode(stufferrorpb.Code_INTERNAL_OPEN_DATABASE_FAILED, err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			seedlog.Errorf("close db failed: %v", err)
		}
	}()

	// Check existence and ownership
	existing, err := db.Stuff.Get(ctx, id)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, seederr.CodeErrorf(stufferrorpb.Code_NOTFOUND_STUFF, "stuff not found: %s", id)
		}
		return nil, seederr.DefaultCode(codes.Unknown, err)
	}
	existingPb := getStuffProtoFromEnt(existing)
	if existingPb.Owner != nil && *existingPb.Owner != "" {
		loginUser, err := login.LoginUser(ctx)
		if err != nil || loginUser == nil || loginUser.Email != *existingPb.Owner {
			return nil, seederr.CodeErrorf(stufferrorpb.Code_DENIED_NOT_OWNER, "user is not the owner")
		}
	}

	update := db.Stuff.UpdateOneID(id)
	if req.Msg.Stuff.Order != "" {
		update.SetOrder(req.Msg.Stuff.Order)
	}
	if req.Msg.Stuff.Data != "" {
		raw := json.RawMessage(req.Msg.Stuff.Data)
		update.SetData(&raw)
	}

	row, err := update.Save(ctx)
	if err != nil {
		return nil, seederr.DefaultCode(codes.Unknown, err)
	}

	return connect.NewResponse(getStuffProtoFromEnt(row)), nil
}

func (svc *StuffService) DeleteStuff(
	ctx context.Context,
	req *connect.Request[stuffpb.DeleteStuffRequest],
) (*connect.Response[emptypb.Empty], error) {
	_, err := login.EnsureLoginUser(ctx)
	if err != nil {
		return nil, seederr.DefaultCode(codes.Unauthenticated, err)
	}
	id, err := uuid.Parse(req.Msg.Uuid)
	if err != nil {
		return nil, seederr.CodeErrorf(stufferrorpb.Code_INVALID_STUFF_ID, "stuff id is invalid")
	}

	db, err := stuffdb.Open(ctx)
	if err != nil {
		return nil, seederr.DefaultCode(stufferrorpb.Code_INTERNAL_OPEN_DATABASE_FAILED, err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			seedlog.Errorf("close db failed: %v", err)
		}
	}()

	// Check ownership
	existing, err := db.Stuff.Get(ctx, id)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, seederr.CodeErrorf(stufferrorpb.Code_NOTFOUND_STUFF, "stuff not found: %s", id)
		}
		return nil, seederr.DefaultCode(codes.Unknown, err)
	}
	existingPb := getStuffProtoFromEnt(existing)
	if existingPb.Owner != nil && *existingPb.Owner != "" {
		loginUser, err := login.LoginUser(ctx)
		if err != nil || loginUser == nil || loginUser.Email != *existingPb.Owner {
			return nil, seederr.CodeErrorf(stufferrorpb.Code_DENIED_NOT_OWNER, "user is not the owner")
		}
	}

	err = db.Stuff.DeleteOneID(id).Exec(ctx)
	if err != nil {
		return nil, seederr.DefaultCode(codes.Unknown, err)
	}

	return connect.NewResponse(&emptypb.Empty{}), nil
}

func (svc *StuffService) ListStuff(
	ctx context.Context,
	req *connect.Request[stuffpb.ListStuffRequest],
) (*connect.Response[stuffpb.ListStuffResponse], error) {
	_, err := login.EnsureLoginUser(ctx)
	if err != nil {
		return nil, seederr.DefaultCode(codes.Unauthenticated, err)
	}
	db, err := stuffdb.Open(ctx)
	if err != nil {
		return nil, seederr.DefaultCode(stufferrorpb.Code_INTERNAL_OPEN_DATABASE_FAILED, err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			seedlog.Errorf("close db failed: %v", err)
		}
	}()

	rows, err := db.Stuff.Query().All(ctx)
	if err != nil {
		return nil, seederr.DefaultCode(codes.Unknown, err)
	}

	stuffList := make([]*stuffpb.Stuff, 0, len(rows))
	for _, row := range rows {
		stuffList = append(stuffList, getStuffProtoFromEnt(row))
	}

	return connect.NewResponse(&stuffpb.ListStuffResponse{
		Stuff: stuffList,
	}), nil
}
