// Command normalize-xattrs rewrites namespaced v3 file-capability xattrs to the
// portable v2 form inside a `podman save` (docker-archive) tar, in place.
//
// build_freedom_image runs `podman build --userns auto` so postinst scripts can
// chown to high UIDs. Under that user namespace the kernel stores a file
// capability (e.g. cap_net_raw on /usr/bin/ping) as a VFS_CAP_REVISION_3 xattr
// whose rootid is the build's mapped root. A consumer that loads the image under
// a different, single-uid user namespace -- e.g. a nested rootless `podman load`
// -- cannot map that rootid, so applying the xattr fails with
// `lsetxattr ...: invalid argument`. Rewriting the cap to the non-namespaced
// VFS_CAP_REVISION_2 form (rootid dropped, implicitly 0) keeps the capability but
// makes it portable to any namespace.
//
// The cap lives in one layer, so converting it changes that layer's tar bytes and
// thus its diff_id. The tool repairs the archive to stay internally consistent:
// it renames the layer blob, rewrites the diff_id in the image config (which
// renames the config blob) and points manifest.json at the new names.
package main

import (
	"archive/tar"
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json/v2"
	"io"
	"os"
	"strings"

	"github.com/ndscm/theseed/seed/infra/error/go/seederr"
	"github.com/ndscm/theseed/seed/infra/init/go/seedinit"
	"github.com/ndscm/theseed/seed/infra/log/go/seedlog"
)

// capKey is the PAX extended-header keyword libarchive/GNU tar use to carry a
// file's security.capability xattr.
const capKey = "SCHILY.xattr.security.capability"

const (
	// vfsCapRevisionMask isolates the revision from the flag bits in magic_etc.
	vfsCapRevisionMask = 0xFF000000
	vfsCapRevision2    = 0x02000000
	vfsCapRevision3    = 0x03000000
	// A v3 cap is magic_etc + two 32-bit permitted/inheritable words + rootid;
	// v2 is the same without the trailing rootid.
	vfsCapV3Len = 24
	vfsCapV2Len = 20
)

// entry is one member of the outer docker-archive tar, held in memory so the
// archive can be rewritten with renamed blobs.
type entry struct {
	hdr  *tar.Header
	data []byte
}

// manifestItem is the subset of a docker-archive manifest.json entry we read to
// find the config and layer blob names.
type manifestItem struct {
	Config string   `json:"Config"`
	Layers []string `json:"Layers"`
}

// capV3ToV2 returns the portable v2 encoding of a v3 capability value, or ("",
// false) if raw is not a v3 cap.
func capV3ToV2(raw string) (string, bool) {
	if len(raw) != vfsCapV3Len {
		return "", false
	}
	b := []byte(raw)
	magic := uint32(b[0]) | uint32(b[1])<<8 | uint32(b[2])<<16 | uint32(b[3])<<24
	if magic&vfsCapRevisionMask != vfsCapRevision3 {
		return "", false
	}
	newMagic := (magic &^ uint32(vfsCapRevisionMask)) | vfsCapRevision2
	v2 := make([]byte, vfsCapV2Len)
	v2[0] = byte(newMagic)
	v2[1] = byte(newMagic >> 8)
	v2[2] = byte(newMagic >> 16)
	v2[3] = byte(newMagic >> 24)
	copy(v2[4:], b[4:vfsCapV2Len])
	return string(v2), true
}

// convertLayer rewrites every v3 cap in a layer tar to v2, returning the new tar
// bytes and whether anything changed.
func convertLayer(data []byte) ([]byte, bool, error) {
	tr := tar.NewReader(bytes.NewReader(data))
	buf := bytes.Buffer{}
	tw := tar.NewWriter(&buf)
	changed := false
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, false, seederr.Wrap(err)
		}
		if v2, ok := capV3ToV2(hdr.PAXRecords[capKey]); ok {
			hdr.PAXRecords[capKey] = v2
			// Keep the deprecated Xattrs view consistent so the writer can't
			// re-emit the stale v3 value.
			if hdr.Xattrs != nil {
				hdr.Xattrs["security.capability"] = v2
			}
			changed = true
		}
		err = tw.WriteHeader(hdr)
		if err != nil {
			return nil, false, seederr.Wrap(err)
		}
		if hdr.Typeflag == tar.TypeReg {
			_, err = io.Copy(tw, tr)
			if err != nil {
				return nil, false, seederr.Wrap(err)
			}
		}
	}
	err := tw.Close()
	if err != nil {
		return nil, false, seederr.Wrap(err)
	}
	return buf.Bytes(), changed, nil
}

func sha256Hex(b []byte) string {
	sum := sha256.Sum256(b)
	return hex.EncodeToString(sum[:])
}

func readArchive(path string) ([]entry, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, seederr.Wrap(err)
	}
	defer f.Close()

	entries := []entry{}
	tr := tar.NewReader(f)
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, seederr.Wrap(err)
		}
		data := []byte{}
		if hdr.Typeflag == tar.TypeReg {
			data, err = io.ReadAll(tr)
			if err != nil {
				return nil, seederr.Wrap(err)
			}
		}
		entries = append(entries, entry{hdr: hdr, data: data})
	}
	return entries, nil
}

func writeArchive(path string, entries []entry) error {
	tmp := path + ".tmp"
	f, err := os.Create(tmp)
	if err != nil {
		return seederr.Wrap(err)
	}
	tw := tar.NewWriter(f)
	for _, e := range entries {
		err = tw.WriteHeader(e.hdr)
		if err != nil {
			return seederr.Wrap(err)
		}
		if e.hdr.Typeflag == tar.TypeReg {
			_, err = tw.Write(e.data)
			if err != nil {
				return seederr.Wrap(err)
			}
		}
	}
	err = tw.Close()
	if err != nil {
		return seederr.Wrap(err)
	}
	err = f.Close()
	if err != nil {
		return seederr.Wrap(err)
	}
	err = os.Rename(tmp, path)
	if err != nil {
		return seederr.Wrap(err)
	}
	return nil
}

// manifestBlobs returns the set of layer blob names and the set of config blob
// names referenced by the archive's manifest.json.
func manifestBlobs(entries []entry) (map[string]bool, map[string]bool, error) {
	raw := []byte{}
	found := false
	for _, e := range entries {
		if e.hdr.Name == "manifest.json" {
			raw = e.data
			found = true
			break
		}
	}
	if !found {
		return nil, nil, seederr.WrapErrorf("manifest.json not found in archive")
	}
	items := []manifestItem{}
	err := json.Unmarshal(raw, &items)
	if err != nil {
		return nil, nil, seederr.Wrap(err)
	}
	layers := map[string]bool{}
	configs := map[string]bool{}
	for _, it := range items {
		configs[it.Config] = true
		for _, l := range it.Layers {
			layers[l] = true
		}
	}
	return layers, configs, nil
}

func normalize(imagePath string) error {
	entries, err := readArchive(imagePath)
	if err != nil {
		return seederr.Wrap(err)
	}
	layerSet, configSet, err := manifestBlobs(entries)
	if err != nil {
		return seederr.Wrap(err)
	}

	// Convert the layers that carry a namespaced cap. Blob names are the content
	// digest, so derive the old digest by hashing rather than trusting the name,
	// then swap that substring in the name (handles `<digest>.tar` and any
	// `blobs/sha256/<digest>` variant alike).
	diffRename := map[string]string{}
	blobRename := map[string]string{}
	for i := range entries {
		e := &entries[i]
		if !layerSet[e.hdr.Name] {
			continue
		}
		newData, changed, err := convertLayer(e.data)
		if err != nil {
			return seederr.Wrap(err)
		}
		if !changed {
			continue
		}
		oldHex := sha256Hex(e.data)
		newHex := sha256Hex(newData)
		diffRename["sha256:"+oldHex] = "sha256:" + newHex
		newName := strings.ReplaceAll(e.hdr.Name, oldHex, newHex)
		blobRename[e.hdr.Name] = newName
		e.data = newData
		e.hdr.Name = newName
		e.hdr.Size = int64(len(newData))
	}
	if len(diffRename) == 0 {
		seedlog.Infof("no namespaced capabilities found; leaving %s unchanged", imagePath)
		return nil
	}

	// Rewrite the config blobs' diff_ids, which changes their digest and so their
	// blob name too.
	for i := range entries {
		e := &entries[i]
		if !configSet[e.hdr.Name] {
			continue
		}
		oldHex := sha256Hex(e.data)
		newData := e.data
		for old, updated := range diffRename {
			newData = bytes.ReplaceAll(newData, []byte(old), []byte(updated))
		}
		newHex := sha256Hex(newData)
		newName := strings.ReplaceAll(e.hdr.Name, oldHex, newHex)
		blobRename[e.hdr.Name] = newName
		e.data = newData
		e.hdr.Name = newName
		e.hdr.Size = int64(len(newData))
	}

	// Point manifest.json at every renamed blob.
	for i := range entries {
		e := &entries[i]
		if e.hdr.Name != "manifest.json" {
			continue
		}
		newData := e.data
		for old, updated := range blobRename {
			newData = bytes.ReplaceAll(newData, []byte(old), []byte(updated))
		}
		e.data = newData
		e.hdr.Size = int64(len(newData))
	}

	seedlog.Infof("normalized %d capability layer(s) in %s", len(diffRename), imagePath)
	err = writeArchive(imagePath, entries)
	if err != nil {
		return seederr.Wrap(err)
	}
	return nil
}

func run() error {
	args, err := seedinit.Initialize()
	if err != nil {
		return seederr.Wrap(err)
	}
	if len(args) != 1 {
		return seederr.WrapErrorf("usage: normalize-xattrs <image.tar>")
	}
	imagePath := args[0]
	if !strings.HasSuffix(imagePath, ".tar") {
		return seederr.WrapErrorf("expected .tar archive, got: %q", imagePath)
	}
	err = normalize(imagePath)
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
