package seedspa

import (
	"net/http"
	"path"
	"sort"
	"strings"
)

// containsExtension checks whether the final segment of filePath contains a
// dot. A dot signals a static asset (e.g. "app.js"), so the SPA fallback is
// skipped. If a dynamic route has a literal dot in its last segment, the dot
// must be URL-escaped (e.g. "v1%2E0"); otherwise it will be treated as an
// asset request and won't fall back.
func containsExtension(filePath string) bool {
	fileName := path.Base(filePath)
	return strings.Contains(fileName, ".")
}

type spaFileSystem struct {
	webapp http.FileSystem

	prefixes  []string
	fallbacks map[string]string
}

func (sfs *spaFileSystem) Open(name string) (http.File, error) {
	// Open succeeds for both files and directories. For directories,
	// http.FileServer will follow up with Open("dir/index.html").
	file, err := sfs.webapp.Open(name)
	if err == nil {
		return file, nil
	}
	if !containsExtension(name) {
		for _, prefix := range sfs.prefixes {
			if strings.HasPrefix(name, prefix) {
				spaFile, err := sfs.webapp.Open(sfs.fallbacks[prefix])
				if err != nil {
					return nil, err
				}
				return spaFile, nil
			}
		}
	}
	return nil, err
}

// SpaServer returns an [http.Handler] that serves static files from webapp,
// with SPA-style fallback routing. Keys in fallbacks are path prefixes
// (e.g. "/es/") and values are the files to serve when no real file matches
// under that prefix (e.g. "/es/fallback.html").
//
// Every prefix directory must also contain an index.html. When a request
// targets the directory itself (e.g. "/es"), [http.FileServer] resolves it to
// "/es/index.html" before the fallback logic ever runs.
func SpaServer(webapp http.FileSystem, fallbacks map[string]string) http.Handler {
	prefixes := []string{}
	for p := range fallbacks {
		prefixes = append(prefixes, p)
	}
	// Sort longest first so more specific prefixes match before shorter ones.
	sort.Slice(prefixes,
		func(i int, j int) bool {
			return len(prefixes[i]) > len(prefixes[j])
		})
	handler := http.FileServer(&spaFileSystem{
		webapp:    webapp,
		prefixes:  prefixes,
		fallbacks: fallbacks,
	})
	return handler
}
