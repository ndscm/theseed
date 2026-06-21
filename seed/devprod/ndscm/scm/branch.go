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

// BaseBranchName returns the base branch tracking branchName. The base is the
// same owner/focus under the "base" branch type
// (e.g. "alice/dev/web" -> "alice/base/dev/web").
func BaseBranchName(branchName string) string {
	ownerHandle, branchType, remain, err := ParseCanonicalBranch(branchName)
	if err != nil {
		return ""
	}
	return ownerHandle + "/base/" + branchType + "/" + remain
}

// ChangeBranchName returns the change branch for featureName on devBranch. It is
// "<owner>/change/<focus>/<featureName>" derived from the dev branch.
func ChangeBranchName(devBranch string, featureName string) (string, error) {
	ownerHandle, branchType, focus, err := ParseCanonicalBranch(devBranch)
	if err != nil {
		return "", seederr.Wrap(err)
	}
	if branchType != "dev" {
		return "", seederr.WrapErrorf("workspace branch is not a dev branch: %v", devBranch)
	}
	return ownerHandle + "/change/" + focus + "/" + featureName, nil
}
