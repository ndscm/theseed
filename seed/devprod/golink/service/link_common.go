package service

import (
	"crypto/sha256"
	"encoding/hex"
	"strings"
	"time"

	"github.com/ndscm/theseed/seed/devprod/golink/database/ent"
	"github.com/ndscm/theseed/seed/devprod/golink/proto/golinkpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// normalizeKey converts key to lowercase, replaces underscores with dashes,
// and removes invalid characters (only keeps a-z, 0-9, -).
func normalizeKey(key string) string {
	var result strings.Builder
	for _, r := range key {
		switch {
		case r >= 'A' && r <= 'Z':
			result.WriteRune(r + 32) // to lowercase
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9', r == '-':
			result.WriteRune(r)
		case r == '_':
			result.WriteRune('-')
		}
	}
	return result.String()
}

// computeEtag generates an etag from the link data.
func computeEtag(key, target string, updatedTime time.Time) string {
	h := sha256.New()
	h.Write([]byte(key))
	h.Write([]byte(target))
	h.Write([]byte(updatedTime.String()))
	return hex.EncodeToString(h.Sum(nil))[:16]
}

func getLinkProtoFromEnt(row *ent.Link) *golinkpb.Link {
	if row == nil {
		return nil
	}
	link := golinkpb.Link{
		Key:         row.ID,
		Target:      row.Target,
		Public:      row.Public,
		Owner:       row.Owner,
		HitCount:    row.HitCount,
		CreatedTime: timestamppb.New(row.CreatedTime),
		UpdatedTime: timestamppb.New(row.UpdatedTime),
		Etag:        computeEtag(row.ID, row.Target, row.UpdatedTime),
	}
	return &link
}

func getLinkEntFromProto(link *golinkpb.Link) *ent.Link {
	if link == nil {
		return nil
	}
	row := ent.Link{
		ID:       normalizeKey(link.Key),
		Target:   link.Target,
		Public:   link.Public,
		Owner:    link.Owner,
		HitCount: link.HitCount,
	}
	if link.CreatedTime != nil {
		row.CreatedTime = link.CreatedTime.AsTime()
	}
	if link.UpdatedTime != nil {
		row.UpdatedTime = link.UpdatedTime.AsTime()
	}
	return &row
}
