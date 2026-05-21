package reactrouter

import (
	"io/fs"
	"net/http"

	"github.com/ndscm/theseed/seed/infra/error/go/seederr"
	"github.com/ndscm/theseed/seed/infra/spa/go/seedspa"
)

func I18nSpaServer(webapp fs.FS) (http.Handler, []string, error) {
	extraLanguages := []string{}
	fallbacks := map[string]string{"": "/__spa-fallback.html"}
	langEntries, err := fs.ReadDir(webapp, ".")
	if err != nil {
		return nil, nil, seederr.Wrap(err)
	}
	for _, langEntry := range langEntries {
		if !langEntry.IsDir() {
			continue
		}
		lang := langEntry.Name()
		fallback := "/" + lang + "/__spa-fallback.html"
		_, err := webapp.Open(lang + "/__spa-fallback.html")
		if err != nil {
			continue
		}
		extraLanguages = append(extraLanguages, lang)
		fallbacks["/"+lang+"/"] = fallback
	}

	handler := seedspa.SpaServer(http.FS(webapp), fallbacks)
	return handler, extraLanguages, nil
}
