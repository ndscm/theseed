// Command xslashtar runs `xslash-helper create` into a scratch directory and
// tars the result - the generated .xslash entry and its side-by-side symlinks -
// preserving the symlinks (which plain Bazel outputs can't represent). Staging
// and archiving in-process keeps the build action off the host shell.
package main

import (
	"archive/tar"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/ndscm/theseed/seed/infra/error/go/seederr"
	"github.com/ndscm/theseed/seed/infra/flag/go/seedflag"
	"github.com/ndscm/theseed/seed/infra/init/go/seedinit"
	"github.com/ndscm/theseed/seed/infra/log/go/seedlog"
)

var flagXslashHelper = seedflag.DefineString("xslash_helper", "", "Path to the xslash-helper executable")
var flagOutEntry = seedflag.DefineString("out_entry", "", "Name of the generated entry file inside the tar")
var flagOutTar = seedflag.DefineString("out_tar", "", "Output tar file path")

// See: https://github.com/bazelbuild/rules_pkg/blob/1.2.0/pkg/private/tar/tar_writer.py#L34
var portableTime = time.Date(2000, time.January, 1, 0, 0, 0, 0, time.UTC)

func tarDir(dir, outTar string) error {
	f, err := os.Create(outTar)
	if err != nil {
		return seederr.Wrap(err)
	}
	defer f.Close()
	tarWriter := tar.NewWriter(f)
	defer tarWriter.Close()
	walkErr := filepath.WalkDir(dir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return seederr.Wrap(err)
		}
		rel, err := filepath.Rel(dir, path)
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
		hdr.ModTime = portableTime
		hdr.AccessTime = time.Time{}
		hdr.ChangeTime = time.Time{}
		hdr.Uid = 0
		hdr.Gid = 0
		hdr.Uname = ""
		hdr.Gname = ""
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
	args, err := seedinit.Initialize(&seedinit.InitializeOptions{
		AnywhereFlag: true,
	})
	if err != nil {
		return seederr.Wrap(err)
	}
	outTar := flagOutTar.Get()
	xslashHelper := flagXslashHelper.Get()
	outEntry := flagOutEntry.Get()

	if len(args) > 0 && args[0] == "--" {
		args = args[1:]
	}

	stage, err := os.MkdirTemp("", "xslashtar")
	if err != nil {
		return seederr.Wrap(err)
	}
	defer os.RemoveAll(stage)

	args = append(args, "--out", filepath.Join(stage, outEntry))
	seedlog.Infof("Running tool: %s %v", xslashHelper, args)
	cmd := exec.Command(xslashHelper, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	if err != nil {
		return seederr.Wrap(err)
	}

	seedlog.Infof("Archiving %s to %s", stage, outTar)
	err = tarDir(stage, outTar)
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
