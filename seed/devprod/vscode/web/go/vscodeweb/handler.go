// Package vscodeweb serves the VS Code workbench (vscode-web) assets and the
// bootstrap page an embedding page loads into an iframe to draw one.
//
// The workbench is drawn in an iframe rather than in the embedding page itself:
// VS Code allows one workbench per page and disposing it does not give the page
// another, so a page that showed one and wants a second — a different folder, a
// second visit — would have to reload, and the embedding page is a single-page
// app that cannot. An iframe is a page of its own; a fresh one is a fresh page,
// and removing it takes the workbench, its worker, and its file system with it.
//
// So `embed.html` is served here: the bootstrap the iframe loads. It calls
// create() with the options the embedding page sends it once it is ready — which
// folder, on whose file system, through which extension — because those are the
// embedding page's to say and differ every time it is asked. What the page hands
// the iframe is in //seed/devprod/vscode/web/tsx.
package vscodeweb

import (
	"embed"
	"io"
	"io/fs"
	"net/http"
	"path"
	"strings"

	"github.com/ndscm/theseed/seed/infra/error/go/seederr"
)

//go:embed all:node_modules
var embedFs embed.FS

// embedHTML is the bootstrap the iframe loads. It is a page beside the workbench
// assets, so it reaches them by relative path — `out/...` — and needs to know
// neither the origin it is served from nor the prefix it is mounted under.
//
// It draws nothing until the embedding page tells it to. On load it says it is
// ready and then waits: a `create` message carries the folder to open, whose
// file system it is on, and where each extension it loads is served, and only
// then is the workbench created. Both sides post same-origin — the iframe is
// served from the embedding page's own origin — and each checks that a message
// came from the other before acting on it.
//
// The extension host would rather run on another origin off a CDN, out of the
// product baked into the workbench; emptying webEndpointUrlTemplate is what makes
// VS Code start it in a same-origin worker instead, out of the files beside the
// workbench, which is the only place ours are. That same-origin worker is also
// what lets the embedding page answer the file system: it and the worker share a
// BroadcastChannel across the iframe, being of one origin.
const embedHTML = `<!DOCTYPE html>
<html>
  <head>
    <meta charset="utf-8" />
    <meta name="viewport" content="width=device-width, initial-scale=1.0, maximum-scale=1.0, minimum-scale=1.0, user-scalable=no" />
    <title>Visual Studio Code</title>
    <link rel="stylesheet" href="out/vs/workbench/workbench.web.main.internal.css" />
    <style>
      html, body { height: 100%; width: 100%; margin: 0; overflow: hidden; }
    </style>
    <script>
      globalThis._VSCODE_FILE_ROOT = new URL("out/", document.baseURI).toString();
    </script>
  </head>
  <body aria-label=""></body>
  <script type="module">
    import "./out/nls.messages.js";
    import { create, URI } from "./out/vs/workbench/workbench.web.main.internal.js";

    const parentOrigin = window.location.origin;

    window.addEventListener("message", (event) => {
      if (event.origin !== parentOrigin || event.source !== window.parent) {
        return;
      }
      const options = event.data;
      if (!options || options.type !== "create") {
        return;
      }
      try {
        create(document.body, {
          productConfiguration: { webEndpointUrlTemplate: "" },
          additionalBuiltinExtensions: (options.additionalExtensions || []).map(
            (path) => ({
              scheme: window.location.protocol.replace(":", ""),
              authority: window.location.host,
              path,
            }),
          ),
          workspaceProvider: {
            workspace: {
              folderUri: URI.from({
                scheme: options.workspaceScheme,
                authority: options.workspaceAuthority,
                path: options.workspacePath,
              }),
            },
            open: async () => false,
            trusted: true,
          },
        });
      } catch (error) {
        const message = error && error.message ? error.message : String(error);
        window.parent.postMessage({ type: "error", message }, parentOrigin);
      }
    });

    window.parent.postMessage({ type: "ready" }, parentOrigin);
  </script>
</html>
`

// contentTypes covers the extensions the workbench loads: the standard library derives
// the type from the system mime table, which does not carry all of them.
var contentTypes = map[string]string{
	".css":   "text/css",
	".html":  "text/html",
	".js":    "text/javascript",
	".json":  "application/json",
	".map":   "application/json",
	".mjs":   "text/javascript",
	".png":   "image/png",
	".svg":   "image/svg+xml",
	".ttf":   "font/ttf",
	".wasm":  "application/wasm",
	".woff2": "font/woff2",
}

type VscodeWebHandler struct {
	files http.Handler
}

func (h *VscodeWebHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// The bootstrap is the one thing served here that is not a workbench asset:
	// it is a page beside them, not one of them, so it is written out rather than
	// read off the embedded dist.
	if path.Base(r.URL.Path) == "embed.html" {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		io.WriteString(w, embedHTML)
		return
	}

	contentType, ok := contentTypes[path.Ext(r.URL.Path)]
	if ok {
		w.Header().Set("Content-Type", contentType)
	}
	h.files.ServeHTTP(w, r)
}

var _ http.Handler = (*VscodeWebHandler)(nil)

// CreateAssetProvider creates a handler serving the vscode-web assets, mounted under
// the given path prefix, e.g. "/vscode-web".
func CreateAssetProvider(prefix string) (string, *VscodeWebHandler, error) {
	prefix = strings.TrimRight(prefix, "/") + "/vscode-web"
	webapp, err := fs.Sub(embedFs, "node_modules/vscode-web")
	if err != nil {
		return "", nil, seederr.Wrap(err)
	}

	h := &VscodeWebHandler{
		files: http.StripPrefix(prefix, http.FileServerFS(webapp)),
	}
	return prefix + "/", h, nil
}
