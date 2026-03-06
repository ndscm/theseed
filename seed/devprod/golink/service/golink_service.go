package service

import (
	"context"
	"database/sql"
	"errors"
	"slices"

	"connectrpc.com/connect"
	"github.com/ndscm/theseed/seed/devprod/golink/database/golinkdb"
	"github.com/ndscm/theseed/seed/devprod/golink/proto/golinkpb"
	"github.com/ndscm/theseed/seed/infra/http/go/seedjwt"
	"github.com/ndscm/theseed/seed/infra/error/go/seederr"
	"github.com/ndscm/theseed/seed/infra/log/go/seedlog"
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
	linkRow := getLinkRowFromLinkProto(req.Msg.Link)
	if linkRow == nil {
		return nil, seederr.WrapErrorf("link is required")
	}
	linkRow.Key = normalizeKey(linkRow.Key)
	if linkRow.Key == "" {
		return nil, seederr.WrapErrorf("key cannot be empty")
	}
	jwtUser, err := seedjwt.JwtUser(ctx)
	if err == nil && jwtUser != nil && jwtUser.Email != "" {
		linkRow.Owner = &jwtUser.Email
	}

	db, err := golinkdb.Open(ctx)
	if err != nil {
		return nil, seederr.Wrap(err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			seedlog.Errorf("close db failed: %v", err)
		}
	}()

	row, err := golinkdb.InsertLink(ctx, db, linkRow)
	if err != nil {
		return nil, seederr.Wrap(err)
	}

	return connect.NewResponse(getLinkProtoFromLinkRow(row)), nil
}

func (svc *GolinkService) GetLink(
	ctx context.Context,
	req *connect.Request[golinkpb.GetLinkRequest],
) (*connect.Response[golinkpb.Link], error) {
	key := normalizeKey(req.Msg.Key)
	if key == "" {
		return nil, seederr.WrapErrorf("key cannot be empty")
	}

	db, err := golinkdb.Open(ctx)
	if err != nil {
		return nil, seederr.Wrap(err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			seedlog.Errorf("close db failed: %v", err)
		}
	}()

	row, err := golinkdb.SelectLinkByKey(ctx, db, key)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, seederr.WrapErrorf("link not found: %s", key)
	}
	if err != nil {
		return nil, seederr.Wrap(err)
	}

	return connect.NewResponse(getLinkProtoFromLinkRow(row)), nil
}

func (svc *GolinkService) UpdateLink(
	ctx context.Context,
	req *connect.Request[golinkpb.UpdateLinkRequest],
) (*connect.Response[golinkpb.Link], error) {
	linkRow := getLinkRowFromLinkProto(req.Msg.Link)
	if linkRow == nil {
		return nil, seederr.WrapErrorf("link is required")
	}
	linkRow.Key = normalizeKey(linkRow.Key)
	if linkRow.Key == "" {
		return nil, seederr.WrapErrorf("key cannot be empty")
	}
	linkRow.Owner = nil // do not trust owner from request
	jwtUser, err := seedjwt.JwtUser(ctx)
	if err == nil && jwtUser != nil && jwtUser.Email != "" {
		linkRow.Owner = &jwtUser.Email
	}

	db, err := golinkdb.Open(ctx)
	if err != nil {
		return nil, seederr.Wrap(err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			seedlog.Errorf("close db failed: %v", err)
		}
	}()

	// Check ownership
	existingRow, err := golinkdb.SelectLinkByKey(ctx, db, linkRow.Key)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, seederr.WrapErrorf("link not found: %s", linkRow.Key)
	}
	if err != nil {
		return nil, seederr.Wrap(err)
	}
	existing := getLinkProtoFromLinkRow(existingRow)
	if existing.Owner != nil && *existing.Owner != "" {
		if linkRow.Owner == nil || *linkRow.Owner != *existing.Owner {
			return nil, seederr.WrapErrorf("not the owner of this link")
		}
	}

	// Check etag
	if req.Msg.Etag != "" && req.Msg.Etag != existing.Etag {
		return nil, seederr.WrapErrorf("etag mismatch")
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
	row, err := golinkdb.UpdateLink(ctx, db, linkRow.Key, linkRow, updateFields)
	if err != nil {
		return nil, seederr.Wrap(err)
	}

	return connect.NewResponse(getLinkProtoFromLinkRow(row)), nil
}

func (svc *GolinkService) DeleteLink(
	ctx context.Context,
	req *connect.Request[golinkpb.DeleteLinkRequest],
) (*connect.Response[emptypb.Empty], error) {
	key := normalizeKey(req.Msg.Key)
	if key == "" {
		return nil, seederr.WrapErrorf("key cannot be empty")
	}

	db, err := golinkdb.Open(ctx)
	if err != nil {
		return nil, seederr.Wrap(err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			seedlog.Errorf("close db failed: %v", err)
		}
	}()

	// Check ownership
	existingRow, err := golinkdb.SelectLinkByKey(ctx, db, key)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, seederr.WrapErrorf("link not found: %s", key)
	}
	if err != nil {
		return nil, seederr.Wrap(err)
	}
	existing := getLinkProtoFromLinkRow(existingRow)
	if existing.Owner != nil && *existing.Owner != "" {
		jwtUser, err := seedjwt.JwtUser(ctx)
		if err != nil {
			return nil, seederr.Wrap(err)
		}
		if jwtUser == nil || jwtUser.Email != *existing.Owner {
			return nil, seederr.WrapErrorf("not the owner of this link")
		}
	}

	// Check etag
	if req.Msg.Etag != "" && req.Msg.Etag != existing.Etag {
		return nil, seederr.WrapErrorf("etag mismatch")
	}

	err = golinkdb.DeleteLink(ctx, db, key)
	if err != nil {
		return nil, seederr.Wrap(err)
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
		return nil, seederr.Wrap(err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			seedlog.Errorf("close db failed: %v", err)
		}
	}()

	// page_token is the last key from previous page
	cursor := req.Msg.PageToken

	rows, err := golinkdb.SelectLinks(ctx, db, cursor, pageSize+1)
	if err != nil {
		return nil, seederr.Wrap(err)
	}

	nextPageToken := ""
	if len(rows) > pageSize {
		// More results available, set next_page_token to last key we'll return
		nextPageToken = rows[pageSize-1].Key
		rows = rows[:pageSize]
	}

	links := []*golinkpb.Link{}
	for _, row := range rows {
		links = append(links, getLinkProtoFromLinkRow(row))
	}

	totalSize, err := golinkdb.CountLinks(ctx, db)
	if err != nil {
		return nil, seederr.Wrap(err)
	}

	return connect.NewResponse(&golinkpb.ListLinksResponse{
		Links:         links,
		NextPageToken: nextPageToken,
		TotalSize:     totalSize,
	}), nil
}
