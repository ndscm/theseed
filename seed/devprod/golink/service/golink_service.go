package service

import (
	"context"
	"slices"

	"connectrpc.com/connect"
	"github.com/ndscm/theseed/seed/cloud/login/go/login"
	"github.com/ndscm/theseed/seed/devprod/golink/database/ent"
	"github.com/ndscm/theseed/seed/devprod/golink/database/golinkdb"
	"github.com/ndscm/theseed/seed/devprod/golink/proto/golinkerrorpb"
	"github.com/ndscm/theseed/seed/devprod/golink/proto/golinkpb"
	"github.com/ndscm/theseed/seed/infra/error/go/seederr"
	"github.com/ndscm/theseed/seed/infra/log/go/seedlog"
	"google.golang.org/grpc/codes"
	"google.golang.org/protobuf/types/known/emptypb"
)

var whitelistPathsOfUpdateLink = []string{
	"target",
	"public",
}

type GolinkService struct{}

func (svc *GolinkService) CreateLink(
	ctx context.Context,
	req *connect.Request[golinkpb.CreateLinkRequest],
) (*connect.Response[golinkpb.Link], error) {
	row := getLinkEntFromProto(req.Msg.Link)
	if row == nil {
		return nil, seederr.CodeErrorf(golinkerrorpb.Code_INVALID_LINK, "link is required")
	}
	key := normalizeKey(row.ID)
	if key == "" {
		return nil, seederr.CodeErrorf(golinkerrorpb.Code_INVALID_KEY, "key is invalid")
	}
	loginUser, err := login.LoginUser(ctx)
	if err == nil && loginUser != nil && loginUser.Email != "" {
		row.Owner = &loginUser.Email
	}

	db, err := golinkdb.Open(ctx)
	if err != nil {
		return nil, seederr.DefaultCode(golinkerrorpb.Code_INTERNAL_OPEN_DATABASE_FAILED, err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			seedlog.Errorf("close db failed: %v", err)
		}
	}()

	resultRow, err := golinkdb.InsertLink(ctx, db, key, row)
	if err != nil {
		if ent.IsConstraintError(err) {
			return nil, seederr.CodeErrorf(golinkerrorpb.Code_EXISTS_LINK, "link already exists for: %s", key)
		}
		return nil, seederr.DefaultCode(codes.Unknown, err)
	}

	responsePb := getLinkProtoFromEnt(resultRow)
	return connect.NewResponse(responsePb), nil
}

func (svc *GolinkService) GetLink(
	ctx context.Context,
	req *connect.Request[golinkpb.GetLinkRequest],
) (*connect.Response[golinkpb.Link], error) {
	key := normalizeKey(req.Msg.Key)
	if key == "" {
		return nil, seederr.CodeErrorf(golinkerrorpb.Code_INVALID_KEY, "key is invalid")
	}

	db, err := golinkdb.Open(ctx)
	if err != nil {
		return nil, seederr.DefaultCode(golinkerrorpb.Code_INTERNAL_OPEN_DATABASE_FAILED, err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			seedlog.Errorf("close db failed: %v", err)
		}
	}()

	row, err := golinkdb.SelectLink(ctx, db, key)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, seederr.CodeErrorf(golinkerrorpb.Code_NOTFOUND_LINK, "link is not found for: %s", key)
		}
		return nil, seederr.DefaultCode(codes.Unknown, err)
	}

	responsePb := getLinkProtoFromEnt(row)
	return connect.NewResponse(responsePb), nil
}

func (svc *GolinkService) UpdateLink(
	ctx context.Context,
	req *connect.Request[golinkpb.UpdateLinkRequest],
) (*connect.Response[golinkpb.Link], error) {
	row := getLinkEntFromProto(req.Msg.Link)
	if row == nil {
		return nil, seederr.CodeErrorf(golinkerrorpb.Code_INVALID_LINK, "link is required")
	}
	key := normalizeKey(row.ID)
	if key == "" {
		return nil, seederr.CodeErrorf(golinkerrorpb.Code_INVALID_KEY, "key is invalid")
	}
	row.Owner = nil // do not trust owner from request
	loginUser, err := login.LoginUser(ctx)
	if err == nil && loginUser != nil && loginUser.Email != "" {
		row.Owner = &loginUser.Email
	}

	db, err := golinkdb.Open(ctx)
	if err != nil {
		return nil, seederr.DefaultCode(golinkerrorpb.Code_INTERNAL_OPEN_DATABASE_FAILED, err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			seedlog.Errorf("close db failed: %v", err)
		}
	}()

	// Check ownership
	existingRow, err := golinkdb.SelectLink(ctx, db, key)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, seederr.CodeErrorf(golinkerrorpb.Code_NOTFOUND_LINK, "link is not found for: %s", key)
		}
		return nil, seederr.DefaultCode(codes.Unknown, err)
	}
	existingPb := getLinkProtoFromEnt(existingRow)
	if existingPb.Owner != nil && *existingPb.Owner != "" {
		if row.Owner == nil || *row.Owner != *existingPb.Owner {
			return nil, seederr.CodeErrorf(golinkerrorpb.Code_DENIED_NOT_LINK_OWNER, "user is not the link owner")
		}
	}

	// Check etag
	if req.Msg.Etag != "" && req.Msg.Etag != existingPb.Etag {
		return nil, seederr.CodeErrorf(golinkerrorpb.Code_ABORTED_ETAG_MISMATCH, "etag mismatch")
	}

	updatePaths := []string{}
	if req.Msg.UpdateMask != nil && len(req.Msg.UpdateMask.Paths) > 0 {
		updatePaths = append(updatePaths, req.Msg.UpdateMask.Paths...)
	}
	if len(updatePaths) == 0 || slices.Contains(updatePaths, "*") {
		updatePaths = whitelistPathsOfUpdateLink
	}
	updatePaths = append(updatePaths, "owner")
	updateFields := updatePaths // translate proto fields to db fields
	resultRow, err := golinkdb.UpdateLink(ctx, db, key, row, updateFields)
	if err != nil {
		return nil, seederr.DefaultCode(codes.Unknown, err)
	}

	responsePb := getLinkProtoFromEnt(resultRow)
	return connect.NewResponse(responsePb), nil
}

func (svc *GolinkService) DeleteLink(
	ctx context.Context,
	req *connect.Request[golinkpb.DeleteLinkRequest],
) (*connect.Response[emptypb.Empty], error) {
	key := normalizeKey(req.Msg.Key)
	if key == "" {
		return nil, seederr.CodeErrorf(golinkerrorpb.Code_INVALID_KEY, "key is invalid")
	}

	db, err := golinkdb.Open(ctx)
	if err != nil {
		return nil, seederr.DefaultCode(golinkerrorpb.Code_INTERNAL_OPEN_DATABASE_FAILED, err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			seedlog.Errorf("close db failed: %v", err)
		}
	}()

	// Check ownership
	existingRow, err := golinkdb.SelectLink(ctx, db, key)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, seederr.CodeErrorf(golinkerrorpb.Code_NOTFOUND_LINK, "link is not found for: %s", key)
		}
		return nil, seederr.DefaultCode(codes.Unknown, err)
	}
	existingPb := getLinkProtoFromEnt(existingRow)
	if existingPb.Owner != nil && *existingPb.Owner != "" {
		loginUser, err := login.LoginUser(ctx)
		if err != nil {
			return nil, seederr.DefaultCode(codes.Unknown, err)
		}
		if loginUser == nil || loginUser.Email != *existingPb.Owner {
			return nil, seederr.CodeErrorf(golinkerrorpb.Code_DENIED_NOT_LINK_OWNER, "user is not the link owner")
		}
	}

	// Check etag
	if req.Msg.Etag != "" && req.Msg.Etag != existingPb.Etag {
		return nil, seederr.CodeErrorf(golinkerrorpb.Code_ABORTED_ETAG_MISMATCH, "etag mismatch")
	}

	err = golinkdb.DeleteLink(ctx, db, key)
	if err != nil {
		return nil, seederr.DefaultCode(codes.Unknown, err)
	}

	return connect.NewResponse(&emptypb.Empty{}), nil
}

func (svc *GolinkService) ListLinks(
	ctx context.Context,
	req *connect.Request[golinkpb.ListLinksRequest],
) (*connect.Response[golinkpb.ListLinksResponse], error) {
	pageSize := int(req.Msg.PageSize)
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 50
	}

	db, err := golinkdb.Open(ctx)
	if err != nil {
		return nil, seederr.DefaultCode(golinkerrorpb.Code_INTERNAL_OPEN_DATABASE_FAILED, err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			seedlog.Errorf("close db failed: %v", err)
		}
	}()

	cursor := req.Msg.PageToken

	rows, total, err := golinkdb.SelectLinks(ctx, db, cursor, pageSize)
	if err != nil {
		return nil, seederr.Wrap(err)
	}

	nextPageToken := ""
	if len(rows) >= pageSize {
		nextPageToken = rows[len(rows)-1].ID
	}

	links := []*golinkpb.Link{}
	for _, row := range rows {
		links = append(links, getLinkProtoFromEnt(row))
	}

	responsePb := &golinkpb.ListLinksResponse{
		Links:         links,
		NextPageToken: nextPageToken,
		TotalSize:     total,
	}
	return connect.NewResponse(responsePb), nil
}
