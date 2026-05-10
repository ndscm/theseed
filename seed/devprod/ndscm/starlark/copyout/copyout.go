package main

import (
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/ndscm/theseed/seed/devprod/starlark/inbazel/go/inbazel"
	"github.com/ndscm/theseed/seed/infra/error/go/seederr"
	"github.com/ndscm/theseed/seed/infra/init/go/seedinit"
	"github.com/ndscm/theseed/seed/infra/log/go/seedlog"
)

// DstInfo describes a copy destination. If Dir is false, Srcs must contain
// exactly one file path which is copied to the destination. If Dir is true,
// all Src directories are merged and Src files are placed into the destination.
type DstInfo struct {
	// Dir controls whether the destination is a directory or a single file.
	Dir bool `json:"dir"`

	// Srcs lists source paths (absolute or relative to cwd).
	Srcs []string `json:"srcs"`
}

type DstMap map[string]DstInfo

func CopyOut(dstBaseDir string, dstMap DstMap, sandbox bool) error {
	for dstRelPath, dstInfo := range dstMap {
		if !dstInfo.Dir {
			if len(dstInfo.Srcs) == 0 {
				return seederr.WrapErrorf("missing src for file dst: %q", dstRelPath)
			}
			if len(dstInfo.Srcs) > 1 {
				return seederr.WrapErrorf("multiple srcs for file dst %q: %v", dstRelPath, dstInfo.Srcs)
			}
			src := dstInfo.Srcs[0]
			dst := filepath.Join(dstBaseDir, dstRelPath)
			err := os.MkdirAll(filepath.Dir(dst), 0755)
			if err != nil {
				return seederr.Wrap(err)
			}
			resolved := src
			if sandbox {
				resolved, err = inbazel.Unlink(src)
				if err != nil {
					return seederr.Wrap(err)
				}
			}
			err = inbazel.CopyFile(resolved, dst)
			if err != nil {
				return seederr.Wrap(err)
			}
		} else {
			dst := filepath.Join(dstBaseDir, dstRelPath)
			err := os.MkdirAll(dst, 0755)
			if err != nil {
				return seederr.Wrap(err)
			}
			for _, src := range dstInfo.Srcs {
				info, err := os.Stat(src)
				if err != nil {
					return seederr.Wrap(err)
				}
				if info.IsDir() {
					err = inbazel.CopyTree(src, dst, sandbox)
					if err != nil {
						return seederr.Wrap(err)
					}
				} else {
					resolved := src
					if sandbox {
						resolved, err = inbazel.Unlink(src)
						if err != nil {
							return seederr.Wrap(err)
						}
					}
					err = inbazel.CopyFile(resolved, filepath.Join(dst, filepath.Base(src)))
					if err != nil {
						return seederr.Wrap(err)
					}
				}
			}
		}
	}
	return nil
}

func run() error {
	args, err := seedinit.Initialize(
		seedinit.WithAnywhereFlag(true),
	)
	if err != nil {
		return seederr.Wrap(err)
	}
	if len(args) != 2 {
		return seederr.WrapErrorf("usage: copyout <dst_base_dir> <dst_map.json>")
	}
	dstBaseDir := args[0]
	dstMapPath := args[1]
	raw, err := os.ReadFile(dstMapPath)
	if err != nil {
		return seederr.Wrap(err)
	}
	dstMap := DstMap{}
	err = json.Unmarshal(raw, &dstMap)
	if err != nil {
		return seederr.Wrap(err)
	}
	sandbox, err := inbazel.CheckSandbox()
	if err != nil {
		return seederr.Wrap(err)
	}
	err = CopyOut(dstBaseDir, dstMap, sandbox)
	if err != nil {
		return seederr.Wrap(err)
	}
	return nil
}

func main() {
	err := run()
	if err != nil {
		seedlog.Errorf("%v", err)
		os.Exit(1)
	}
}
