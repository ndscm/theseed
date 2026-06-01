package git

import (
	"strings"
	"time"

	"github.com/ndscm/theseed/seed/infra/error/go/seederr"
	"github.com/ndscm/theseed/seed/infra/log/go/seedlog"
	"github.com/ndscm/theseed/seed/infra/shell/go/seedshell"
)

type Trailer struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type CommitMetadata struct {
	Hash           string    `json:"hash"`
	Tree           string    `json:"tree"`
	Parents        []string  `json:"parents"`
	Author         string    `json:"author"`
	AuthorEmail    string    `json:"authorEmail"`
	AuthorTime     time.Time `json:"authorTime"`
	Committer      string    `json:"committer"`
	CommitterEmail string    `json:"committerEmail"`
	CommitterTime  time.Time `json:"committerTime"`
	Subject        string    `json:"subject"`
	Body           string    `json:"body"`

	Trailers []Trailer `json:"trailers"`
}

func (c *CommitMetadata) GetChangeUuid() (string, error) {
	changeUuid := ""
	for _, trailer := range c.Trailers {
		if trailer.Key == "Change-uuid" {
			if changeUuid != "" {
				return "", seederr.WrapErrorf("multiple Change-uuid trailers found for commit %v", c.Hash)
			}
			changeUuid = trailer.Value
		}
	}
	if changeUuid == "" {
		return "", seederr.WrapErrorf("Change-uuid not found for commit %v", c.Hash)
	}
	seedlog.Debugf("Found commit change uuid. commit=%v changeUuid=%v", c.Hash, changeUuid)
	return changeUuid, nil
}

func GetCommitHash(gitDir string, commit string) (string, error) {
	gitArgs := []string{}
	if gitDir != "" {
		gitArgs = append(gitArgs, "--git-dir", gitDir)
	}
	commitOutput, err := seedshell.PureOutput("git", append(gitArgs, "rev-parse", commit)...)
	if err != nil {
		return "", seederr.WrapErrorf("failed to get commit hash for %v: %w", commit, err)
	}
	commitHash := strings.TrimSpace(string(commitOutput))
	return commitHash, nil
}

func ListCommitHash(gitDir string, from string, to string) ([]string, error) {
	gitArgs := []string{}
	if gitDir != "" {
		gitArgs = append(gitArgs, "--git-dir", gitDir)
	}
	listOutput, err := seedshell.PureOutput("git", append(gitArgs, "rev-list", "--ancestry-path", from+".."+to)...)
	if err != nil {
		return nil, seederr.WrapErrorf("failed to get commit hash for range %v..%v: %w", from, to, err)
	}
	trimmed := strings.TrimSpace(string(listOutput))
	if trimmed == "" {
		return nil, nil
	}
	commits := strings.Split(trimmed, "\n")
	return commits, nil
}

func ListMergeCommitHash(gitDir string, from string, to string) ([]string, error) {
	gitArgs := []string{}
	if gitDir != "" {
		gitArgs = append(gitArgs, "--git-dir", gitDir)
	}
	listOutput, err := seedshell.PureOutput("git", append(gitArgs, "rev-list", "--merges", "--ancestry-path", from+".."+to)...)
	if err != nil {
		return nil, seederr.WrapErrorf("failed to get merge commit hash for range %v..%v: %w", from, to, err)
	}
	trimmed := strings.TrimSpace(string(listOutput))
	if trimmed == "" {
		return nil, nil
	}
	commits := strings.Split(trimmed, "\n")
	return commits, nil
}

func GetCommitMetadata(gitDir string, commit string) (*CommitMetadata, error) {
	gitArgs := []string{}
	if gitDir != "" {
		gitArgs = append(gitArgs, "--git-dir", gitDir)
	}
	format := strings.Join([]string{
		"%H",                      // 0: hash
		"%T",                      // 1: tree
		"%P",                      // 2: parents (space-separated)
		"%an",                     // 3: author
		"%ae",                     // 4: author email
		"%aI",                     // 5: author time
		"%cn",                     // 6: committer
		"%ce",                     // 7: committer email
		"%cI",                     // 8: committer time
		"%s",                      // 9: subject
		"%b",                      // 10: body
		"%(trailers:only,unfold)", // 11: trailers
	}, "%x00")
	commitOutput, err := seedshell.PureOutput("git", append(gitArgs, "log", "-1", "--format="+format, commit)...)
	if err != nil {
		return nil, seederr.Wrap(err)
	}
	fields := strings.Split(string(commitOutput), "\x00")
	if len(fields) != 12 {
		return nil, seederr.WrapErrorf("unexpected format output for commit %v: got %d fields", commit, len(fields))
	}
	authorTime, err := time.Parse(time.RFC3339, fields[5])
	if err != nil {
		return nil, seederr.Wrap(err)
	}
	committerTime, err := time.Parse(time.RFC3339, fields[8])
	if err != nil {
		return nil, seederr.Wrap(err)
	}
	parents := []string{}
	if fields[2] != "" {
		parents = strings.Split(fields[2], " ")
	}
	trailers := []Trailer{}
	rawTrailers := strings.TrimSpace(fields[11])
	if rawTrailers != "" {
		for _, line := range strings.Split(rawTrailers, "\n") {
			key, value, found := strings.Cut(line, ": ")
			if !found {
				continue
			}
			trailers = append(trailers, Trailer{
				Key:   strings.TrimSpace(key),
				Value: strings.TrimSpace(value),
			})
		}
	}
	result := &CommitMetadata{
		Hash:           fields[0],
		Tree:           fields[1],
		Parents:        parents,
		Author:         fields[3],
		AuthorEmail:    fields[4],
		AuthorTime:     authorTime,
		Committer:      fields[6],
		CommitterEmail: fields[7],
		CommitterTime:  committerTime,
		Subject:        fields[9],
		Body:           fields[10],
		Trailers:       trailers,
	}
	seedlog.Debugf("Commit %v: %#v", result.Hash, result)
	return result, nil
}

func AmendHeadCommit(worktreePath string, trailerKey string, text string) error {
	gitArgs := []string{}
	if worktreePath != "" {
		gitArgs = append(gitArgs, "-C", worktreePath)
	}
	gitArgs = append(gitArgs, "commit", "--amend", "--no-edit", "--trailer", trailerKey+": "+text)
	err := seedshell.ImpureRun("git", gitArgs...)
	if err != nil {
		return seederr.WrapErrorf("failed to amend head commit with trailer %v: %w", trailerKey, err)
	}
	return nil
}
