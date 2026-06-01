package clientcore

import (
	"strings"

	"github.com/google/uuid"
	"github.com/ndscm/theseed/seed/devprod/ndscm/scm"
	"github.com/ndscm/theseed/seed/infra/error/go/seederr"
)

type NdAmendOptions struct {
	Break        string
	Migrate      string
	SideEffectOf string
}

func NdAmend(scmProvider scm.Provider, options NdAmendOptions) error {
	if options.Break == "" && options.Migrate == "" && options.SideEffectOf == "" {
		return seederr.WrapErrorf("nd-amend requires at least one of --break, --migrate, or --side-effect-of")
	}
	if options.Break != "" {
		if options.Break == "lock" && options.Migrate == "" {
			options.Migrate = "update lock"
		}
		if options.Break == "melt" && options.Migrate == "" {
			options.Migrate = "drop"
		}
		err := scmProvider.AmendHeadCommit("Break", options.Break)
		if err != nil {
			return seederr.Wrap(err)
		}
	}
	if options.Migrate != "" {
		err := scmProvider.AmendHeadCommit("Migrate", options.Migrate)
		if err != nil {
			return seederr.Wrap(err)
		}
	}
	if options.SideEffectOf != "" {
		sideEffectOf := strings.TrimSpace(strings.ToLower(options.SideEffectOf))
		switch sideEffectOf {
		case "break":
			commitId, err := scmProvider.SearchTrailer("Break", "")
			if err != nil {
				return seederr.Wrap(err)
			}
			trailers, err := scmProvider.ListCommitTrailers(commitId)
			if err != nil {
				return seederr.Wrap(err)
			}
			found := false
			for _, trailer := range trailers {
				if trailer.Key == "Change-uuid" {
					if found {
						return seederr.WrapErrorf("multiple Change-uuid trailers found in commit %s", commitId)
					}
					sideEffectOf = trailer.Text
					found = true
				}
			}
			if !found {
				return seederr.WrapErrorf("no Change-uuid trailer found in commit %s", commitId)
			}
		default:
			parsedUuid, err := uuid.Parse(options.SideEffectOf)
			if err != nil {
				return seederr.WrapErrorf("--side-effect-of must be a valid UUID or search term, but got '%s'", options.SideEffectOf)
			}
			sideEffectOf = parsedUuid.String()
		}
		err := scmProvider.AmendHeadCommit("Side-effect-of-change-uuid", sideEffectOf)
		if err != nil {
			return seederr.Wrap(err)
		}
	}
	return nil
}
