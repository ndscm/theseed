// Command vscode-web-demo-server serves the VS Code workbench (vscode-web) in the
// browser.
package main

import (
	"embed"
	"io"
	"io/fs"
	"net/http"
	"os"
	"path"

	"github.com/ndscm/theseed/seed/infra/error/go/seederr"
	"github.com/ndscm/theseed/seed/infra/flag/go/seedflag"
	"github.com/ndscm/theseed/seed/infra/init/go/seedinit"
	"github.com/ndscm/theseed/seed/infra/log/go/seedlog"
)

// flagPort is the port the workbench is served on.
var flagPort = seedflag.DefineString("port", "2633", "Server port")

//go:embed all:node_modules
var embedFs embed.FS

// indexHtml is the bootstrap page for the workbench, in place of the index.html the
// dist does not ship.
//
// create() takes IWorkbenchConstructionOptions. Empty is legal: with no
// workspaceProvider the workbench opens an empty window, which is all a bare demo needs.
const indexHtml = `<!DOCTYPE html>
<html>
  <head>
    <meta charset="utf-8" />
    <meta name="viewport" content="width=device-width, initial-scale=1.0, maximum-scale=1.0, minimum-scale=1.0, user-scalable=no" />
    <title>Visual Studio Code</title>
    <link rel="icon" href="/favicon.ico" type="image/x-icon" />
    <link rel="stylesheet" href="/out/vs/workbench/workbench.web.main.internal.css" />
    <style>
      html, body { height: 100%; width: 100%; margin: 0; overflow: hidden; }
    </style>
    <script>
      globalThis._VSCODE_FILE_ROOT = new URL("/out/", window.location.origin).toString();
    </script>
  </head>
  <body aria-label=""></body>
  <script type="module" src="/out/nls.messages.js"></script>
  <script type="module">
    import { create } from "/out/vs/workbench/workbench.web.main.internal.js";
    create(document.body, {});
  </script>
</html>
`

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

func run() error {
	_, err := seedinit.Initialize()
	if err != nil {
		return seederr.Wrap(err)
	}

	webapp, err := fs.Sub(embedFs, "node_modules/vscode-web")
	if err != nil {
		return seederr.Wrap(err)
	}

	files := http.FileServerFS(webapp)
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" || r.URL.Path == "/index.html" {
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			io.WriteString(w, indexHtml)
			return
		}

		contentType, ok := contentTypes[path.Ext(r.URL.Path)]
		if ok {
			w.Header().Set("Content-Type", contentType)
		}
		files.ServeHTTP(w, r)
	})

	port := flagPort.Get()
	seedlog.Infof("Starting vscode web demo server on :%v", port)
	err = http.ListenAndServe(":"+port, handler)
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
