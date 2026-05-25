package account

import (
	"fmt"
	"os"
	"syscall"

	"github.com/ndscm/theseed/seed/infra/error/go/seederr"
)

type AcmeAccountLocker interface {
	Lock() error
	Unlock() error
}

type LocalAcmeAccountLocker struct {
	acmeId      string
	acmeAccount *AcmeAccount

	lockFile *os.File
}

func (l *LocalAcmeAccountLocker) getLockPath() string {
	// The SFE certificate service must not be scaled across multiple nodes, so
	// a local file lock is sufficient — no need for a distributed lock.
	return fmt.Sprintf("/tmp/%s_%s.lock", l.acmeId, l.acmeAccount.GetEmail())
}

func (l *LocalAcmeAccountLocker) Lock() error {
	f, err := os.OpenFile(l.getLockPath(), os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		return seederr.WrapErrorf("open lock file failed: %w", err)
	}
	err = syscall.Flock(int(f.Fd()), syscall.LOCK_EX)
	if err != nil {
		f.Close()
		return seederr.WrapErrorf("lock acme account failed: %w", err)
	}
	l.lockFile = f
	return nil
}

func (l *LocalAcmeAccountLocker) Unlock() error {
	if l.lockFile == nil {
		return nil
	}
	err := l.lockFile.Close()
	l.lockFile = nil
	if err != nil {
		return seederr.WrapErrorf("unlock acme account failed: %w", err)
	}
	return nil
}

// Ensure LocalAcmeAccountLocker implements AcmeAccountLocker.
var _ AcmeAccountLocker = &LocalAcmeAccountLocker{}

func NewLocalAcmeAccountLocker(acmeId string, acmeAccount *AcmeAccount) *LocalAcmeAccountLocker {
	l := &LocalAcmeAccountLocker{
		acmeId:      acmeId,
		acmeAccount: acmeAccount,
	}
	return l
}
