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

	// Side-effect-of-change-uuid is usually added before the new break and migrate metadata.
	if options.SideEffectOf != "" {
		sideEffectOf := strings.TrimSpace(strings.ToLower(options.SideEffectOf))
		switch sideEffectOf {
		case "break":
			breakCommitId, err := scmProvider.SearchExtendedMetadata("HEAD^", "Break", "")
			if err != nil {
				return seederr.Wrap(err)
			}
			breakCommitMetadata, err := scmProvider.GetCommitMetadata(breakCommitId)
			if err != nil {
				return seederr.Wrap(err)
			}
			parsedUuid, err := uuid.Parse(breakCommitMetadata.ChangeUuid)
			if err != nil {
				return seederr.WrapErrorf("break commit does not have a valid change UUID. commit=%s", breakCommitId)
			}
			sideEffectOf = parsedUuid.String()
		default:
			parsedUuid, err := uuid.Parse(options.SideEffectOf)
			if err != nil {
				return seederr.WrapErrorf("side effect is not a valid UUID. sideEffectOf='%s'", options.SideEffectOf)
			}
			sideEffectOf = parsedUuid.String()
		}
		err := scmProvider.AmendAppendExtendedMetadata("Side-effect-of-change-uuid", sideEffectOf)
		if err != nil {
			return seederr.Wrap(err)
		}
	}

	if options.Break != "" {
		if options.Break == "lock" && options.Migrate == "" {
			options.Migrate = "update lock"
		}
		if options.Break == "melt" && options.Migrate == "" {
			options.Migrate = "drop"
		}
		err := scmProvider.AmendAppendExtendedMetadata("Break", options.Break)
		if err != nil {
			return seederr.Wrap(err)
		}
	}
	if options.Migrate != "" {
		err := scmProvider.AmendAppendExtendedMetadata("Migrate", options.Migrate)
		if err != nil {
			return seederr.Wrap(err)
		}
	}
	return nil
}
