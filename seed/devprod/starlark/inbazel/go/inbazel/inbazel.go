package inbazel

import (
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/ndscm/theseed/seed/infra/error/go/seederr"
)

// CheckSandbox checks if we are running in a bazel linux sandbox
// by looking for a read-only root mount.
func CheckSandbox() (bool, error) {
	mounts, err := os.ReadFile("/proc/self/mounts")
	if err != nil {
		return false, seederr.Wrap(err)
	}
	for _, line := range strings.Split(string(mounts), "\n") {
		fields := strings.Fields(line)
		if len(fields) >= 4 && fields[1] == "/" {
			for _, mode := range strings.Split(fields[3], ",") {
				if mode == "ro" {
					return true, nil
				}
			}
			return false, nil
		}
	}
	return false, nil
}

// Unlink resolves a bazel-wrapped symlink. When sandbox is true, it
// reads the symlink target and joins it with the parent directory.
func Unlink(src string) (string, error) {
	srcStat, err := os.Lstat(src)
	if err != nil {
		return "", seederr.Wrap(err)
	}
	if srcStat.Mode()&os.ModeSymlink == 0 {
		return "", seederr.WrapErrorf("expected symlink, got regular file (not in bazel sandbox?): %s", src)
	}
	rawPath, err := os.Readlink(src)
	if err != nil {
		return "", seederr.Wrap(err)
	}
	srcPath := filepath.Join(filepath.Dir(src), rawPath)
	return srcPath, nil
}

func CopyFile(src string, dst string) error {
	srcReader, err := os.Open(src)
	if err != nil {
		return seederr.Wrap(err)
	}
	defer srcReader.Close()
	srcStat, err := srcReader.Stat()
	if err != nil {
		return seederr.Wrap(err)
	}
	err = os.Remove(dst)
	if err != nil && !os.IsNotExist(err) {
		return seederr.Wrap(err)
	}
	dstWriter, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, srcStat.Mode())
	if err != nil {
		return seederr.Wrap(err)
	}
	defer dstWriter.Close()
	_, err = io.Copy(dstWriter, srcReader)
	if err != nil {
		return seederr.Wrap(err)
	}
	return nil
}

// CopyTree recursively copies a directory tree from src to dst,
// resolving bazel symlinks when sandbox is true.
func CopyTree(src string, dst string, sandbox bool) error {
	entries, err := os.ReadDir(src)
	if err != nil {
		return seederr.Wrap(err)
	}
	for _, entryStat := range entries {
		s := filepath.Join(src, entryStat.Name())
		d := filepath.Join(dst, entryStat.Name())
		if entryStat.IsDir() {
			err = os.MkdirAll(d, 0755)
			if err != nil {
				return seederr.Wrap(err)
			}
			err = CopyTree(s, d, sandbox)
			if err != nil {
				return seederr.Wrap(err)
			}
			continue
		}
		resolved := s
		if sandbox {
			resolved, err = Unlink(s)
			if err != nil {
				return seederr.Wrap(err)
			}
		}
		err = CopyFile(resolved, d)
		if err != nil {
			return seederr.Wrap(err)
		}
	}
	return nil
}

// CopySrcs copies a list of source files/directories into dstBaseDir.
// stripComponents controls how much of each src path prefix is removed:
//   - >= 0: strip the first N path components (like tar --strip-components).
//   - < 0: flatten — discard the entire src path and place files directly
//     under dstBaseDir using only their basename.
func CopySrcs(srcs []string, dstBaseDir string, stripComponents int, sandbox bool) error {
	for _, src := range srcs {
		dst := src
		if stripComponents < 0 {
			dst = ""
		} else {
			parts := strings.SplitN(dst, string(filepath.Separator), stripComponents+1)
			dst = parts[len(parts)-1]
		}
		realDst := filepath.Join(dstBaseDir, dst)
		err := os.MkdirAll(filepath.Dir(realDst), 0755)
		if err != nil {
			return seederr.Wrap(err)
		}
		srcStat, err := os.Stat(src)
		if err != nil {
			return seederr.Wrap(err)
		}
		if srcStat.IsDir() {
			err = CopyTree(src, realDst, sandbox)
			if err != nil {
				return seederr.Wrap(err)
			}
			continue
		}
		resolved := src
		if sandbox {
			resolved, err = Unlink(src)
			if err != nil {
				return seederr.Wrap(err)
			}
		}
		err = CopyFile(resolved, realDst)
		if err != nil {
			return seederr.Wrap(err)
		}
	}
	return nil
}
