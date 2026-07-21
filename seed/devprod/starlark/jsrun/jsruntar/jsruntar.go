package main

import (
	"archive/tar"
	"encoding/json/v2"
	"io"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/ndscm/theseed/seed/infra/error/go/seederr"
	"github.com/ndscm/theseed/seed/infra/flag/go/seedflag"
	"github.com/ndscm/theseed/seed/infra/init/go/seedinit"
	"github.com/ndscm/theseed/seed/infra/log/go/seedlog"
)

var flagOutDir = seedflag.DefineString("out_dir", "", "Output directory path")
var flagOutTar = seedflag.DefineString("out_tar", "", "Output tar file path")
var flagCopies = seedflag.DefineString("copies", "", "JSON file mapping source paths to destination paths")
var flagTool = seedflag.DefineString("tool", "", "Path to the tool executable")

func copyFile(src, dst string) error {
	err := os.MkdirAll(filepath.Dir(dst), 0755)
	if err != nil {
		return seederr.Wrap(err)
	}
	srcFile, err := os.Open(src)
	if err != nil {
		return seederr.Wrap(err)
	}
	defer srcFile.Close()
	info, err := srcFile.Stat()
	if err != nil {
		return seederr.Wrap(err)
	}
	dstFile, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, info.Mode())
	if err != nil {
		return seederr.Wrap(err)
	}
	defer dstFile.Close()
	_, err = io.Copy(dstFile, srcFile)
	if err != nil {
		return seederr.Wrap(err)
	}
	return nil
}

func tarDir(outDir, outTar string) error {
	f, err := os.Create(outTar)
	if err != nil {
		return seederr.Wrap(err)
	}
	defer f.Close()
	tarWriter := tar.NewWriter(f)
	defer tarWriter.Close()
	walkErr := filepath.WalkDir(outDir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return seederr.Wrap(err)
		}
		rel, err := filepath.Rel(outDir, path)
		if err != nil {
			return seederr.Wrap(err)
		}
		if rel == "." {
			return nil
		}
		info, err := d.Info()
		if err != nil {
			return seederr.Wrap(err)
		}
		var link string
		if info.Mode()&os.ModeSymlink != 0 {
			link, err = os.Readlink(path)
			if err != nil {
				return seederr.Wrap(err)
			}
		}
		hdr, err := tar.FileInfoHeader(info, link)
		if err != nil {
			return seederr.Wrap(err)
		}
		hdr.Name = rel
		if info.IsDir() {
			hdr.Name += "/"
		}
		err = tarWriter.WriteHeader(hdr)
		if err != nil {
			return seederr.Wrap(err)
		}
		if info.Mode().IsRegular() {
			file, err := os.Open(path)
			if err != nil {
				return seederr.Wrap(err)
			}
			_, err = io.Copy(tarWriter, file)
			file.Close()
			if err != nil {
				return seederr.Wrap(err)
			}
		}
		return nil
	})
	if walkErr != nil {
		return seederr.Wrap(walkErr)
	}
	return nil
}

func run() error {
	toolArgs, err := seedinit.Initialize(
		seedinit.WithAnywhereFlag(true),
	)
	if err != nil {
		return seederr.Wrap(err)
	}
	outDir := flagOutDir.Get()
	outTar := flagOutTar.Get()
	copiesPath := flagCopies.Get()
	toolPath := flagTool.Get()

	raw, err := os.ReadFile(copiesPath)
	if err != nil {
		return seederr.Wrap(err)
	}
	copies := map[string]string{}
	err = json.Unmarshal(raw, &copies)
	if err != nil {
		return seederr.Wrap(err)
	}
	seedlog.Infof("Copying %d source files to output tree", len(copies))
	for src, dst := range copies {
		err := copyFile(src, dst)
		if err != nil {
			return seederr.Wrap(err)
		}
	}

	seedlog.Infof("Running tool: %s %v", toolPath, toolArgs)
	cmd := exec.Command(toolPath, toolArgs...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	if err != nil {
		return seederr.Wrap(err)
	}

	seedlog.Infof("Archiving %s to %s", outDir, outTar)
	err = tarDir(outDir, outTar)
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
