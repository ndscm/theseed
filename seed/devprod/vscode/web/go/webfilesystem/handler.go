// Package webfilesystem serves the web-file-system VS Code extension.
//
// The extension is what gives an embedded workbench a file system: VS Code takes
// one only from an extension, and this is the extension a page that embeds a
// workbench loads. It is served as the two files VS Code asks for — the manifest,
// and the one script the manifest names — from wherever this handler is mounted,
// which is the location the page hands the workbench as an additional builtin
// extension.
//
// What the file system does is neither here nor in the extension: the extension
// forwards every question to the page that embedded the workbench. The page's
// half of it is //seed/devprod/vscode/web/ts/web-file-system.
package webfilesystem

import (
	"embed"
	"io/fs"
	"net/http"
	"path"
	"strings"

	"github.com/ndscm/theseed/seed/infra/error/go/seederr"
)

//go:embed all:node_modules
var embedFs embed.FS

// contentTypes covers what an extension is made of. The standard library derives
// the type from the system mime table, which does not carry all of it.
var contentTypes = map[string]string{
	".js":   "text/javascript",
	".json": "application/json",
	".map":  "application/json",
}

type WebFileSystemHandler struct {
	files http.Handler
}

func (h *WebFileSystemHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	contentType, ok := contentTypes[path.Ext(r.URL.Path)]
	if ok {
		w.Header().Set("Content-Type", contentType)
	}
	h.files.ServeHTTP(w, r)
}

var _ http.Handler = (*WebFileSystemHandler)(nil)

// CreateAssetProvider creates a handler serving the extension, mounted under the
// given path prefix, e.g. "/vscode-web-extension/web-file-system", and returns
// the pattern to mount it at along with it.
//
// The prefix is what the page tells the workbench the extension is at, and VS
// Code finds it there by asking for the `package.json` beneath it.
func CreateAssetProvider(prefix string) (string, *WebFileSystemHandler, error) {
	prefix = strings.TrimRight(prefix, "/") + "/vscode-extension/web-file-system"
	extension, err := fs.Sub(embedFs, "node_modules/web-file-system")
	if err != nil {
		return "", nil, seederr.Wrap(err)
	}

	h := &WebFileSystemHandler{
		files: http.StripPrefix(prefix, http.FileServerFS(extension)),
	}
	return prefix + "/", h, nil
}
