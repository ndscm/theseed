package githook

import (
	_ "embed"
)

//go:embed commit-msg
var hookCommitMsgChangeUuid string

func HookCommitMsgChangeUuid() string {
	return hookCommitMsgChangeUuid
}
