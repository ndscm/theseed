package runner

import (
	"bytes"
	"fmt"
	"hash/crc32"
	"os"
	"path/filepath"
	"slices"
	"sync"
	"time"

	"github.com/ndscm/theseed/seed/infra/error/go/seederr"
)

type Crc32cChecksum struct {
	Crc32c uint32
	Path   string
}

func (c Crc32cChecksum) String() string {
	return fmt.Sprintf("%08x", c.Crc32c) + "  " + c.Path
}

func Crc32c(worktreePath string, repoPath string) ([]byte, error) {
	targetPath := filepath.Join(worktreePath, repoPath)
	targetStat, err := os.Stat(targetPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, seederr.Wrap(err)
	}

	checksums := []Crc32cChecksum{}
	crc32cTable := crc32.MakeTable(crc32.Castagnoli)
	if targetStat.IsDir() {
		err := filepath.WalkDir(targetPath, func(path string, d os.DirEntry, err error) error {
			if err != nil {
				return seederr.Wrap(err)
			}
			if d.IsDir() {
				return nil
			}
			data := []byte{}
			if d.Type()&os.ModeSymlink != 0 {
				linkTarget, err := os.Readlink(path)
				if err != nil {
					return seederr.Wrap(err)
				}
				data = []byte(linkTarget)
			} else {
				fileData, err := os.ReadFile(path)
				if err != nil {
					return seederr.Wrap(err)
				}
				data = fileData
			}
			crc := crc32.Checksum(data, crc32cTable)
			relPath, err := filepath.Rel(targetPath, path)
			if err != nil {
				return seederr.Wrap(err)
			}
			checksums = append(checksums, Crc32cChecksum{
				Crc32c: crc,
				Path:   relPath,
			})
			return nil
		})
		if err != nil {
			return nil, seederr.Wrap(err)
		}
	} else {
		fileData, err := os.ReadFile(targetPath)
		if err != nil {
			return nil, seederr.Wrap(err)
		}
		crc := crc32.Checksum(fileData, crc32cTable)
		checksums = append(checksums, Crc32cChecksum{
			Crc32c: crc,
			Path:   "-",
		})
	}

	slices.SortFunc(checksums, func(a Crc32cChecksum, b Crc32cChecksum) int {
		if a.Path < b.Path {
			return -1
		} else if a.Path > b.Path {
			return 1
		}
		return 0
	})

	result := bytes.Buffer{}
	for _, c := range checksums {
		result.WriteString(c.String())
		result.WriteByte('\n')
	}
	return result.Bytes(), nil
}

type PathChecksums struct {
	HasOriginal       bool
	OriginalTime      time.Time
	OriginalChecksums []byte

	NewTime      time.Time
	NewChecksums []byte
}

type WatcherStamper struct {
	worktreePath string

	watcherDigest string

	checksumsMutex sync.Mutex
	checksums      map[string]*PathChecksums
}

func (ws *WatcherStamper) checkChanged(repoPath string) (bool, error) {
	watcherStampHome := filepath.Join(ws.worktreePath, ".cache", "ndscm", "stamp", ws.watcherDigest)
	targetPath := filepath.Join(ws.worktreePath, repoPath)
	checksumsPath := filepath.Join(watcherStampHome, repoPath, "checksums.crc32c")

	targetStat, err := os.Stat(targetPath)
	if err != nil {
		if os.IsNotExist(err) {
			_, err := os.Stat(checksumsPath)
			if err != nil {
				if !os.IsNotExist(err) {
					return false, seederr.Wrap(err)
				}
			} else {
				err := os.Remove(checksumsPath)
				if err != nil {
					return false, seederr.Wrap(err)
				}
			}
			return true, nil
		}
		return false, seederr.Wrap(err)
	}

	pathChecksums, ok := ws.checksums[repoPath]
	if !ok {
		checksumsStat, err := os.Stat(checksumsPath)
		if err != nil {
			if !os.IsNotExist(err) {
				return false, seederr.Wrap(err)
			}
			pathChecksums = &PathChecksums{
				HasOriginal: false,
			}
			ws.checksums[repoPath] = pathChecksums
		} else {
			originalChecksums, err := os.ReadFile(checksumsPath)
			if err != nil {
				return false, seederr.Wrap(err)
			}
			pathChecksums = &PathChecksums{
				HasOriginal:       true,
				OriginalTime:      checksumsStat.ModTime(),
				OriginalChecksums: originalChecksums,
			}
			ws.checksums[repoPath] = pathChecksums
		}
	}

	if pathChecksums.HasOriginal && pathChecksums.OriginalTime.After(targetStat.ModTime()) {
		return false, nil
	}

	if pathChecksums.NewTime.IsZero() || pathChecksums.NewTime.Before(targetStat.ModTime()) {
		newChecksums, err := Crc32c(ws.worktreePath, repoPath)
		if err != nil {
			return false, seederr.Wrap(err)
		}
		pathChecksums.NewTime = time.Now()
		pathChecksums.NewChecksums = newChecksums
		ws.checksums[repoPath] = pathChecksums
	}

	if bytes.Equal(pathChecksums.OriginalChecksums, pathChecksums.NewChecksums) {
		return false, nil
	}
	return true, nil
}

func (ws *WatcherStamper) CheckChanged(repoPath string) (bool, error) {
	ws.checksumsMutex.Lock()
	defer ws.checksumsMutex.Unlock()
	return ws.checkChanged(repoPath)
}

func (ws *WatcherStamper) stamp(repoPath string) (bool, error) {
	changed, err := ws.checkChanged(repoPath)
	if err != nil {
		return false, seederr.Wrap(err)
	}
	if changed {
		pathChecksums := ws.checksums[repoPath]
		if pathChecksums == nil {
			return true, nil
		}
		watcherStampHome := filepath.Join(ws.worktreePath, ".cache", "ndscm", "stamp", ws.watcherDigest)
		checksumsPath := filepath.Join(watcherStampHome, repoPath, "checksums.crc32c")
		err = os.MkdirAll(filepath.Dir(checksumsPath), 0755)
		if err != nil {
			return false, seederr.Wrap(err)
		}
		err = os.WriteFile(checksumsPath, pathChecksums.NewChecksums, 0644)
		if err != nil {
			return false, seederr.Wrap(err)
		}
		pathChecksums.HasOriginal = true
		pathChecksums.OriginalTime = time.Now()
		pathChecksums.OriginalChecksums = pathChecksums.NewChecksums
	}
	return changed, nil
}

func (ws *WatcherStamper) Stamp(repoPath string) (bool, error) {
	ws.checksumsMutex.Lock()
	defer ws.checksumsMutex.Unlock()
	return ws.stamp(repoPath)
}

type RepoStamper struct {
	worktreePath string

	watchersMutex sync.Mutex
	watchers      map[string]*WatcherStamper
}

func (rs *RepoStamper) digest(watcher *Watcher) (string, error) {
	watcherSha256, watcherBytes, err := watcher.Sha256()
	if err != nil {
		return "", seederr.Wrap(err)
	}
	watcherDigest := fmt.Sprintf("%x", watcherSha256)
	watcherStampHome := filepath.Join(rs.worktreePath, ".cache", "ndscm", "stamp", watcherDigest)
	watcherPath := filepath.Join(watcherStampHome, "watcher.json")
	_, err = os.Stat(watcherPath)
	if err != nil {
		if !os.IsNotExist(err) {
			return "", seederr.Wrap(err)
		}
		err := os.MkdirAll(filepath.Dir(watcherPath), 0755)
		if err != nil {
			return "", seederr.Wrap(err)
		}
		err = os.WriteFile(watcherPath, watcherBytes, 0644)
		if err != nil {
			return "", seederr.Wrap(err)
		}
	}
	return watcherDigest, nil
}

func (rs *RepoStamper) watcherStamper(watcher *Watcher) (*WatcherStamper, error) {
	rs.watchersMutex.Lock()
	defer rs.watchersMutex.Unlock()

	watcherDigest, err := rs.digest(watcher)
	if err != nil {
		return nil, seederr.Wrap(err)
	}

	ws, ok := rs.watchers[watcherDigest]
	if !ok {
		ws = &WatcherStamper{
			worktreePath:  rs.worktreePath,
			watcherDigest: watcherDigest,
			checksums:     map[string]*PathChecksums{},
		}
		rs.watchers[watcherDigest] = ws
	}
	return ws, nil
}

func (rs *RepoStamper) CheckChanged(watcher *Watcher, repoPath string) (bool, error) {
	ws, err := rs.watcherStamper(watcher)
	if err != nil {
		return false, seederr.Wrap(err)
	}
	return ws.CheckChanged(repoPath)
}

func (rs *RepoStamper) Stamp(watcher *Watcher, repoPath string) (bool, error) {
	ws, err := rs.watcherStamper(watcher)
	if err != nil {
		return false, seederr.Wrap(err)
	}
	return ws.Stamp(repoPath)
}

func (rs *RepoStamper) StampAll(repoAnalysis *RepoAnalysis) (map[string]bool, error) {
	dirtySet := map[string]bool{}
	for _, repoPhase := range repoAnalysis.Phases {
		for _, watcher := range repoPhase.Watchers {
			for _, targetRepoPath := range watcher.Targets {
				changed, err := rs.Stamp(&watcher, targetRepoPath)
				if err != nil {
					return nil, seederr.Wrap(err)
				}
				if changed {
					dirtySet[targetRepoPath] = true
				}
			}
			for _, watchRepoPath := range watcher.Watch {
				changed, err := rs.Stamp(&watcher, watchRepoPath)
				if err != nil {
					return nil, seederr.Wrap(err)
				}
				if changed {
					dirtySet[watchRepoPath] = true
				}
			}
		}
	}
	return dirtySet, nil
}

func NewRepoStamper(worktreePath string) *RepoStamper {
	rs := &RepoStamper{
		worktreePath: worktreePath,
		watchers:     map[string]*WatcherStamper{},
	}
	return rs
}
