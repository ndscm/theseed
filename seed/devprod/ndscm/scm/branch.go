package scm

import (
	"errors"
	"strings"

	"github.com/ndscm/theseed/seed/infra/error/go/seederr"
	"github.com/ndscm/theseed/seed/infra/flag/go/seedflag"
)

var flagCanonicalBranch = seedflag.DefineBool("canonical_branch", false, "whether to use canonical branch names")

var ErrBranchNotFound = errors.New("branch not found")

func CanonicalBranch() bool {
	return flagCanonicalBranch.Get()
}

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
	if !CanonicalBranch() {
		return strings.HasPrefix(branchName, expectedBranchType+"/")
	}
	_, branchType, _, err := ParseCanonicalBranch(branchName)
	if err != nil {
		return false
	}
	return branchType == expectedBranchType
}

// BaseBranchName returns the base branch tracking branchName. In canonical
// mode the base is the same owner/focus under the "base" branch type
// (e.g. "alice/dev/web" -> "alice/base/dev/web"); otherwise it is the
// legacy "base/<branchName>" form.
func BaseBranchName(branchName string, canonicalBranch bool) string {
	if !canonicalBranch {
		return "base/" + branchName
	}
	ownerHandle, branchType, remain, err := ParseCanonicalBranch(branchName)
	if err != nil {
		return ""
	}
	return ownerHandle + "/base/" + branchType + "/" + remain
}

// ChangeBranchName returns the change branch for featureName on devBranch. In
// canonical mode it is "<owner>/change/<focus>/<featureName>" derived from the
// dev branch; otherwise it is the legacy "change/<featureName>" form.
func ChangeBranchName(devBranch string, featureName string, canonicalBranch bool) (string, error) {
	if !canonicalBranch {
		return "change/" + featureName, nil
	}
	ownerHandle, branchType, focus, err := ParseCanonicalBranch(devBranch)
	if err != nil {
		return "", seederr.Wrap(err)
	}
	if branchType != "dev" {
		return "", seederr.WrapErrorf("workspace branch is not a dev branch: %v", devBranch)
	}
	return ownerHandle + "/change/" + focus + "/" + featureName, nil
}
