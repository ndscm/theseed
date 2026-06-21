package git

import (
	"strings"

	"github.com/ndscm/theseed/seed/infra/error/go/seederr"
	"github.com/ndscm/theseed/seed/infra/log/go/seedlog"
	"github.com/ndscm/theseed/seed/infra/shell/go/seedshell"
)

type SeenEntry struct {
	OurHash   string
	TheirHash string
}

func SearchForkPoint(gitDir string, ourTipPoint string, theirTipPoint string) (string, string, error) {
	seen := map[string]SeenEntry{}

	ourHash := ourTipPoint
	ourMeta, err := GetCommitMetadata(gitDir, ourHash)
	if err != nil {
		return "", "", seederr.Wrap(err)
	}
	ourChangeUuid, err := ourMeta.GetChangeUuid()
	if err != nil {
		return "", "", seederr.Wrap(err)
	}

	theirHash := theirTipPoint
	theirMeta, err := GetCommitMetadata(gitDir, theirHash)
	if err != nil {
		return "", "", seederr.Wrap(err)
	}
	theirChangeUuid, err := theirMeta.GetChangeUuid()
	if err != nil {
		return "", "", seederr.Wrap(err)
	}

	result := SeenEntry{}
	for {
		ourEntry := seen[ourChangeUuid]
		ourEntry.OurHash = ourHash
		seen[ourChangeUuid] = ourEntry

		theirEntry := seen[theirChangeUuid]
		theirEntry.TheirHash = theirHash
		seen[theirChangeUuid] = theirEntry

		ourCandidate := seen[ourChangeUuid]
		if ourCandidate.OurHash != "" && ourCandidate.TheirHash != "" {
			result = ourCandidate
			break
		}
		theirCandidate := seen[theirChangeUuid]
		if theirCandidate.OurHash != "" && theirCandidate.TheirHash != "" {
			result = theirCandidate
			break
		}

		ourCanMove := len(ourMeta.Parents) == 1
		theirCanMove := len(theirMeta.Parents) == 1

		if len(ourMeta.Parents) > 1 {
			return "", "", seederr.WrapErrorf("expected at most one parent for commit %v, got %d", ourHash, len(ourMeta.Parents))
		}
		if len(theirMeta.Parents) > 1 {
			return "", "", seederr.WrapErrorf("expected at most one parent for commit %v, got %d", theirHash, len(theirMeta.Parents))
		}
		if !ourCanMove && !theirCanMove {
			return "", "", seederr.WrapErrorf("no fork point found: both chains reached root at %v and %v", ourHash, theirHash)
		}

		if !theirCanMove || (ourCanMove && !ourMeta.CommitterTime.Before(theirMeta.CommitterTime)) {
			ourHash = ourMeta.Parents[0]
			ourMeta, err = GetCommitMetadata(gitDir, ourHash)
			if err != nil {
				return "", "", seederr.Wrap(err)
			}
			ourChangeUuid, err = ourMeta.GetChangeUuid()
			if err != nil {
				return "", "", seederr.Wrap(err)
			}
		} else {
			theirHash = theirMeta.Parents[0]
			theirMeta, err = GetCommitMetadata(gitDir, theirHash)
			if err != nil {
				return "", "", seederr.Wrap(err)
			}
			theirChangeUuid, err = theirMeta.GetChangeUuid()
			if err != nil {
				return "", "", seederr.Wrap(err)
			}
		}
	}
	seedlog.Debugf("Found fork point. seen=%v", result)

	ourForkPoint := result.OurHash
	theirForkPoint := result.TheirHash
	if ourForkPoint == "" || theirForkPoint == "" {
		return "", "", seederr.WrapErrorf("unexpected error: fork point not found in seen map")
	}
	return ourForkPoint, theirForkPoint, nil
}

func SearchTrailer(gitDir string, tipPoint string, trailerKey string, text string) (string, error) {
	gitArgs := []string{}
	if gitDir != "" {
		gitArgs = append(gitArgs, "--git-dir", gitDir)
	}
	gitArgs = append(gitArgs, "log", "--format=%H%x00%(trailers:key="+trailerKey+",valueonly,separator=%x01)", tipPoint)
	output, err := seedshell.PureOutput("git", gitArgs...)
	if err != nil {
		return "", seederr.WrapErrorf("failed to search trailer %v: %w", trailerKey, err)
	}
	for _, line := range strings.Split(strings.TrimSpace(string(output)), "\n") {
		if line == "" {
			continue
		}
		hash, trailerValues, found := strings.Cut(line, "\x00")
		if !found {
			continue
		}
		trailerValues = strings.TrimSpace(trailerValues)
		if trailerValues == "" {
			continue
		}
		if text == "" {
			return hash, nil
		}
		for _, val := range strings.Split(trailerValues, "\x01") {
			if strings.TrimSpace(val) == text {
				return hash, nil
			}
		}
	}
	return "", seederr.WrapErrorf("no commit found with trailer %v: %v", trailerKey, text)
}
