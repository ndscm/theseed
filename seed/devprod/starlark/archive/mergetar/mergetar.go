package main

import (
	"archive/tar"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/ndscm/theseed/seed/infra/error/go/seederr"
	"github.com/ndscm/theseed/seed/infra/flag/go/seedflag"
	"github.com/ndscm/theseed/seed/infra/init/go/seedinit"
	"github.com/ndscm/theseed/seed/infra/log/go/seedlog"
)

var flagOut = seedflag.DefineString("out", "", "Output tar file path")
var flagSrcs = seedflag.DefineString("srcs", "", "Path to JSON file describing srcs map")

func appendTar(tarWriter *tar.Writer, srcTar string, dstPrefix string) error {
	f, err := os.Open(srcTar)
	if err != nil {
		return seederr.Wrap(err)
	}
	defer f.Close()

	tarReader := tar.NewReader(f)
	for {
		hdr, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return seederr.Wrap(err)
		}

		if dstPrefix != "" && dstPrefix != "." {
			hdr.Name = filepath.Join(dstPrefix, hdr.Name)
			if hdr.Typeflag == tar.TypeDir {
				hdr.Name += "/"
			}
		}

		err = tarWriter.WriteHeader(hdr)
		if err != nil {
			return seederr.Wrap(err)
		}
		if hdr.Typeflag == tar.TypeReg {
			_, err = io.Copy(tarWriter, tarReader)
			if err != nil {
				return seederr.Wrap(err)
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
	if len(args) != 0 {
		return seederr.WrapErrorf("usage: mergetar --out <out.tar> --srcs <srcs.json>")
	}
	outTar := flagOut.Get()
	srcsMapPath := flagSrcs.Get()
	if !strings.HasSuffix(outTar, ".tar") {
		return seederr.WrapErrorf("expected .tar output, got: %q", outTar)
	}
	raw, err := os.ReadFile(srcsMapPath)
	if err != nil {
		return seederr.Wrap(err)
	}
	srcs := map[string]string{}
	err = json.Unmarshal(raw, &srcs)
	if err != nil {
		return seederr.Wrap(err)
	}
	seedlog.Infof("Merging tars. srcs=%v", srcs)
	outFile, err := os.Create(outTar)
	if err != nil {
		return seederr.Wrap(err)
	}
	defer outFile.Close()
	tarWriter := tar.NewWriter(outFile)
	defer tarWriter.Close()

	// Sort srcTars for deterministic output across builds.
	srcTars := []string{}
	for srcTar := range srcs {
		srcTars = append(srcTars, srcTar)
	}
	sort.Strings(srcTars)

	for _, srcTar := range srcTars {
		dstPrefix := srcs[srcTar]
		err := appendTar(tarWriter, srcTar, dstPrefix)
		if err != nil {
			return seederr.Wrap(err)
		}
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
