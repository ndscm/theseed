package seedspa

import (
	"io"
	"net/http"
	"os"
	"path"
	"sort"
	"strings"

	"github.com/ndscm/theseed/seed/infra/error/go/seederr"
)

type resizedFileInfo struct {
	os.FileInfo
	size int64
}

func (rfi *resizedFileInfo) Size() int64 {
	return rfi.size
}

var _ os.FileInfo = (*resizedFileInfo)(nil)

type memoryFile struct {
	*strings.Reader
	stat os.FileInfo
}

func (f *memoryFile) Close() error {
	return nil
}

func (f *memoryFile) Readdir(int) ([]os.FileInfo, error) {
	return nil, nil
}

func (f *memoryFile) Stat() (os.FileInfo, error) {
	return f.stat, nil
}

var _ http.File = (*memoryFile)(nil)

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

	headInjection string
}

func (sfs *spaFileSystem) Open(name string) (http.File, error) {
	// Open succeeds for both files and directories. For directories,
	// http.FileServer will follow up with Open("dir/index.html").
	file, err := sfs.webapp.Open(name)
	if err != nil {
		if !os.IsNotExist(err) {
			return nil, seederr.Wrap(err)
		}
		if containsExtension(name) {
			// raise not found error for static asset requests
			return nil, err
		}
	}

	// raise dir for next round
	if file != nil {
		stat, err := file.Stat()
		if err != nil {
			return nil, seederr.Wrap(err)
		}
		if stat.IsDir() {
			return file, nil
		}
	}

	// try fallback
	if file == nil {
		for _, prefix := range sfs.prefixes {
			if strings.HasPrefix(name, prefix) {
				fallbackFile, err := sfs.webapp.Open(sfs.fallbacks[prefix])
				if err != nil {
					return nil, seederr.Wrap(err)
				}
				file = fallbackFile
				break
			}
		}
	}
	if file == nil {
		return nil, os.ErrNotExist
	}

	// inject head content if needed
	if sfs.headInjection != "" && (!containsExtension(name) || strings.HasSuffix(name, ".html")) {
		stat, err := file.Stat()
		if err != nil {
			return nil, seederr.Wrap(err)
		}
		htmlContent, err := io.ReadAll(file)
		if err != nil {
			return nil, seederr.Wrap(err)
		}
		err = file.Close()
		if err != nil {
			return nil, seederr.Wrap(err)
		}
		injectedHtmlContent := strings.Replace(
			string(htmlContent),
			`</head>`,
			sfs.headInjection+`</head>`,
			1,
		)
		file = &memoryFile{
			Reader: strings.NewReader(injectedHtmlContent),
			stat: &resizedFileInfo{
				FileInfo: stat,
				size:     int64(len(injectedHtmlContent)),
			},
		}
	}
	return file, nil
}

type spaServerConfig struct {
	headInjection string
}

type SpaServerOption func(*spaServerConfig)

func WithHeadInjection(injection string) SpaServerOption {
	return func(config *spaServerConfig) {
		config.headInjection += injection
	}
}

// SpaServer returns an [http.Handler] that serves static files from webapp,
// with SPA-style fallback routing. Keys in fallbacks are path prefixes
// (e.g. "/es/") and values are the files to serve when no real file matches
// under that prefix (e.g. "/es/fallback.html").
//
// Every prefix directory must also contain an index.html. When a request
// targets the directory itself (e.g. "/es"), [http.FileServer] resolves it to
// "/es/index.html" before the fallback logic ever runs.
func SpaServer(webapp http.FileSystem, fallbacks map[string]string, options ...SpaServerOption) http.Handler {
	config := &spaServerConfig{}
	for _, option := range options {
		option(config)
	}

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

		headInjection: config.headInjection,
	})
	return handler
}
