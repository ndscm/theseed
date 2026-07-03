package scm

import (
	"errors"
	"strings"

	"github.com/ndscm/theseed/seed/infra/error/go/seederr"
)

var ErrBranchNotFound = errors.New("branch not found")

func ParseCanonicalBranch(branchName string) (string, string, string, error) {
	ownerHandle, remain, found := strings.Cut(branchName, "/")
	if !found || remain == "" {
		return "", "", "", seederr.WrapErrorf("invalid canonical branch name. branch=%v", branchName)
	}
	branchType, remain, found := strings.Cut(remain, "/")
	if !found || remain == "" {
		return "", "", "", seederr.WrapErrorf("invalid canonical branch name. branch=%v", branchName)
	}
	return ownerHandle, branchType, remain, nil
}

func IsBranchType(branchName string, expectedBranchType string) bool {
	_, branchType, _, err := ParseCanonicalBranch(branchName)
	if err != nil {
		return false
	}
	return branchType == expectedBranchType
}

// GetBaseBranchName returns the name of the base branch tracking branchName. The
// base branch has the same owner and remainder as branchName, with "base"
// inserted as the branch type (e.g. "christina/dev/web" -> "christina/base/dev/web").
func GetBaseBranchName(branchName string) (string, error) {
	ownerHandle, branchType, remain, err := ParseCanonicalBranch(branchName)
	if err != nil {
		return "", seederr.Wrap(err)
	}
	return ownerHandle + "/base/" + branchType + "/" + remain, nil
}

// GetChangeBranchName returns the change branch for featureName on devBranch. It is
// "<owner>/change/<focus>/<featureName>" derived from the dev branch.
func GetChangeBranchName(devBranchName string, featureName string) (string, error) {
	ownerHandle, branchType, focus, err := ParseCanonicalBranch(devBranchName)
	if err != nil {
		return "", seederr.Wrap(err)
	}
	if branchType != "dev" {
		return "", seederr.WrapErrorf("workspace branch is not a dev branch: %v", devBranchName)
	}
	return ownerHandle + "/change/" + focus + "/" + featureName, nil
}

// GetWipBranchName returns the wip branch tracking branchName. The wip branch is the
// same owner/focus under the "wip" branch type
// (e.g. "alice/dev/web" -> "alice/wip/dev/web").
func GetWipBranchName(branchName string) (string, error) {
	ownerHandle, branchType, remain, err := ParseCanonicalBranch(branchName)
	if err != nil {
		return "", seederr.Wrap(err)
	}
	return ownerHandle + "/wip/" + branchType + "/" + remain, nil
}
