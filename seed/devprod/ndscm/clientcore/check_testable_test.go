package clientcore

import (
	"fmt"
	"testing"

	"github.com/ndscm/theseed/seed/devprod/ndscm/scm"
)

type fakeCheckTestableScmProvider struct {
	scm.Provider

	commitIds []string // oldest -> newest, matching ListCommitIds' contract

	metadata map[string]*scm.CommitMetadata
}

func (p *fakeCheckTestableScmProvider) GetCommitId(commit string) (string, error) {
	return commit, nil
}

func (p *fakeCheckTestableScmProvider) GetCommitMetadata(commit string) (*scm.CommitMetadata, error) {
	metadata, ok := p.metadata[commit]
	if !ok {
		return nil, fmt.Errorf("unknown commit %v", commit)
	}
	return metadata, nil
}

func (p *fakeCheckTestableScmProvider) GetBranchTracking(branchName string) (string, error) {
	return "", fmt.Errorf("no tracking branch for %v", branchName)
}

func (p *fakeCheckTestableScmProvider) ListCommitIds(from string, to string) ([]string, error) {
	return p.commitIds, nil
}

type FakeCommitMetadataOptions struct {
	SideEffectOf []string
}

func makeFakeCommitMetadata(id string, changeUuid string, opts *FakeCommitMetadataOptions) *scm.CommitMetadata {
	metadata := &scm.CommitMetadata{
		CommitId:   id,
		ChangeUuid: changeUuid,
	}
	if opts == nil {
		opts = &FakeCommitMetadataOptions{}
	}
	for _, breakUuid := range opts.SideEffectOf {
		metadata.Extended = append(metadata.Extended, scm.KeyValue{
			Key:   "side-effect-of-change-uuid",
			Value: breakUuid,
		})
	}
	return metadata
}

// newFakeCheckTestableScmProvider builds the continuous-range example chain:
//
//	c1(testable) <- c2(break) <- c3(se 2) <- c4(se 2, break) <- c5(se 4) <-
//	c6(se 2) <- c7(se 2) <- c8(clean) <- c9(se 2)
//
// A commit is a break purely because a later commit points a side-effect back
// at it. The untestable range must stay continuous: the clean c8 ends the run,
// so c7 (the run's newest commit) and c9 (pointing back across the gap to the
// ancient c2) are both testable.
func newFakeCheckTestableScmProvider() *fakeCheckTestableScmProvider {
	commits := []*scm.CommitMetadata{
		makeFakeCommitMetadata("c1", "u1", nil),
		makeFakeCommitMetadata("c2", "u2", nil),
		makeFakeCommitMetadata("c3", "u3", &FakeCommitMetadataOptions{SideEffectOf: []string{"u2"}}),
		makeFakeCommitMetadata("c4", "u4", &FakeCommitMetadataOptions{SideEffectOf: []string{"u2"}}),
		makeFakeCommitMetadata("c5", "u5", &FakeCommitMetadataOptions{SideEffectOf: []string{"u4"}}),
		makeFakeCommitMetadata("c6", "u6", &FakeCommitMetadataOptions{SideEffectOf: []string{"u2"}}),
		makeFakeCommitMetadata("c7", "u7", &FakeCommitMetadataOptions{SideEffectOf: []string{"u2"}}),
		makeFakeCommitMetadata("c8", "u8", nil),
		makeFakeCommitMetadata("c9", "u9", &FakeCommitMetadataOptions{SideEffectOf: []string{"u2"}}),
	}
	provider := &fakeCheckTestableScmProvider{metadata: map[string]*scm.CommitMetadata{}}
	for _, commit := range commits {
		provider.commitIds = append(provider.commitIds, commit.CommitId)
		provider.metadata[commit.CommitId] = commit
	}
	return provider
}

func TestCheckTestable(t *testing.T) {
	provider := newFakeCheckTestableScmProvider()
	belong := "test"

	t.Run("c1 before any break is testable", func(t *testing.T) {
		testable, err := CheckTestable(provider, "c1", &belong)
		if err != nil {
			t.Fatalf("CheckTestable(c1) returned unexpected error: %v", err)
		}
		if !testable {
			t.Errorf("c1 should be testable: it precedes every break")
		}
	})

	t.Run("c2 break is untestable", func(t *testing.T) {
		testable, err := CheckTestable(provider, "c2", &belong)
		if err != nil {
			t.Fatalf("CheckTestable(c2) returned unexpected error: %v", err)
		}
		if testable {
			t.Errorf("c2 should be untestable: it is a break with live side-effects")
		}
	})

	t.Run("c3 side-effect of c2 is untestable", func(t *testing.T) {
		testable, err := CheckTestable(provider, "c3", &belong)
		if err != nil {
			t.Fatalf("CheckTestable(c3) returned unexpected error: %v", err)
		}
		if testable {
			t.Errorf("c3 should be untestable: it sits inside c2's open range")
		}
	})

	t.Run("c4 nested break is untestable", func(t *testing.T) {
		testable, err := CheckTestable(provider, "c4", &belong)
		if err != nil {
			t.Fatalf("CheckTestable(c4) returned unexpected error: %v", err)
		}
		if testable {
			t.Errorf("c4 should be untestable: it is inside c2's range and opens its own")
		}
	})

	t.Run("c5 side-effect of nested break is untestable", func(t *testing.T) {
		testable, err := CheckTestable(provider, "c5", &belong)
		if err != nil {
			t.Fatalf("CheckTestable(c5) returned unexpected error: %v", err)
		}
		if testable {
			t.Errorf("c5 should be untestable: it sits inside the open range")
		}
	})

	t.Run("c6 side-effect of c2 is untestable", func(t *testing.T) {
		testable, err := CheckTestable(provider, "c6", &belong)
		if err != nil {
			t.Fatalf("CheckTestable(c6) returned unexpected error: %v", err)
		}
		if testable {
			t.Errorf("c6 should be untestable: c7 keeps c2's range open above it")
		}
	})

	t.Run("c7 newest commit of the continuous run is testable", func(t *testing.T) {
		testable, err := CheckTestable(provider, "c7", &belong)
		if err != nil {
			t.Fatalf("CheckTestable(c7) returned unexpected error: %v", err)
		}
		if !testable {
			t.Errorf("c7 should be testable: the clean c8 above it resets the range")
		}
	})

	t.Run("c8 clean commit is testable", func(t *testing.T) {
		testable, err := CheckTestable(provider, "c8", &belong)
		if err != nil {
			t.Fatalf("CheckTestable(c8) returned unexpected error: %v", err)
		}
		if !testable {
			t.Errorf("c8 should be testable: a clean commit is a consistent checkpoint")
		}
	})

	t.Run("c9 side-effect pointing back across the gap is testable", func(t *testing.T) {
		testable, err := CheckTestable(provider, "c9", &belong)
		if err != nil {
			t.Fatalf("CheckTestable(c9) returned unexpected error: %v", err)
		}
		if !testable {
			t.Errorf("c9 should be testable: the gap at c8 breaks continuity to c2")
		}
	})
}
