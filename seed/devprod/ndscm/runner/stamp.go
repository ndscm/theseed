package runner

import (
	"bytes"
	"fmt"
	"hash/crc32"
	"os"
	"path/filepath"
	"slices"

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

func Stamp(worktreePath string, watcher *Watcher, repoPath string) (bool, error) {
	watcherDigest, watcherBytes, err := watcher.Sha256()
	if err != nil {
		return false, seederr.Wrap(err)
	}
	watcherStampHome := filepath.Join(worktreePath, ".cache", "ndscm", "stamp", fmt.Sprintf("%x", watcherDigest))
	watcherPath := filepath.Join(watcherStampHome, "watcher.json")
	_, err = os.Stat(watcherPath)
	if err != nil {
		if !os.IsNotExist(err) {
			return false, seederr.Wrap(err)
		}
		err := os.MkdirAll(filepath.Dir(watcherPath), 0755)
		if err != nil {
			return false, seederr.Wrap(err)
		}
		err = os.WriteFile(watcherPath, watcherBytes, 0644)
		if err != nil {
			return false, seederr.Wrap(err)
		}
	}

	targetPath := filepath.Join(worktreePath, repoPath)
	checksumsPath := filepath.Join(watcherStampHome, repoPath, "checksums.crc32c")
	changed := false

	targetStat, err := os.Stat(targetPath)
	if err != nil {
		if os.IsNotExist(err) {
			_, err := os.Stat(checksumsPath)
			if err == nil {
				err := os.Remove(checksumsPath)
				if err != nil {
					return false, seederr.Wrap(err)
				}
			}
			if err != nil && !os.IsNotExist(err) {
				return false, seederr.Wrap(err)
			}
			return true, nil
		}
		return false, seederr.Wrap(err)
	}

	checksumsStat, err := os.Stat(checksumsPath)
	if err != nil {
		if !os.IsNotExist(err) {
			return false, seederr.Wrap(err)
		}
		changed = true
	}

	if !changed {
		if checksumsStat.ModTime().After(targetStat.ModTime()) {
			return false, nil
		}
	}

	newChecksums, err := Crc32c(worktreePath, repoPath)
	if err != nil {
		return false, seederr.Wrap(err)
	}

	if !changed {
		originalChecksums, err := os.ReadFile(checksumsPath)
		if err != nil {
			return false, seederr.Wrap(err)
		}
		if !bytes.Equal(originalChecksums, newChecksums) {
			changed = true
		}
	}

	if changed {
		err := os.MkdirAll(filepath.Dir(checksumsPath), 0755)
		if err != nil {
			return false, seederr.Wrap(err)
		}
		err = os.WriteFile(checksumsPath, newChecksums, 0644)
		if err != nil {
			return false, seederr.Wrap(err)
		}
	}
	return changed, nil
}

func StampAll(worktreePath string, repoAnalysis *RepoAnalysis) (map[string]bool, error) {
	dirtySet := map[string]bool{}
	for _, repoPhase := range repoAnalysis.Phases {
		for _, watcher := range repoPhase.Watchers {
			for _, targetRepoPath := range watcher.Targets {
				changed, err := Stamp(worktreePath, &watcher, targetRepoPath)
				if err != nil {
					return nil, seederr.Wrap(err)
				}
				if changed {
					dirtySet[targetRepoPath] = true
				}
			}
			for _, watchRepoPath := range watcher.Watch {
				changed, err := Stamp(worktreePath, &watcher, watchRepoPath)
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
