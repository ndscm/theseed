package service

import (
	"context"
	"database/sql"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/ndscm/theseed/seed/devprod/golink/database/golinkdb"
	"github.com/ndscm/theseed/seed/infra/log/go/seedlog"
)

type GolinkHandler struct {
	Webapp http.Handler
}

func (h *GolinkHandler) serveWebapp(w http.ResponseWriter, r *http.Request) {
	if h.Webapp != nil {
		h.Webapp.ServeHTTP(w, r)
		return
	}
	http.NotFound(w, r)
}

// ServeHTTP implements http.Handler and handles redirect requests.
func (h *GolinkHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	path := strings.TrimPrefix(r.URL.Path, "/")
	if path == "" || path == "favicon.ico" || strings.HasPrefix(path, ".") {
		h.serveWebapp(w, r)
		return
	}

	parts := strings.SplitN(path, "/", 2)
	key := normalizeKey(parts[0])
	rest := ""
	if len(parts) > 1 {
		rest = parts[1]
	}
	if key == "" {
		h.serveWebapp(w, r)
		return
	}

	db, err := golinkdb.Open(ctx)
	if err != nil {
		seedlog.Errorf("Failed to open db: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	defer func() {
		err := db.Close()
		if err != nil {
			seedlog.Errorf("close db failed: %v", err)
		}
	}()

	row, err := golinkdb.SelectLinkByKey(ctx, db, key)
	if errors.Is(err, sql.ErrNoRows) {
		h.serveWebapp(w, r)
		return
	}
	if err != nil {
		seedlog.Errorf("Failed to query link: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Build final URL
	location := row.Target
	if rest != "" {
		if strings.HasSuffix(row.Target, "/") || strings.HasSuffix(row.Target, "=") {
			location = row.Target + rest
		} else {
			location = row.Target + "/" + rest
		}
	}

	http.Redirect(w, r, location, http.StatusTemporaryRedirect)

	// Increment hit count asynchronously
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
		defer cancel()
		db, err := golinkdb.Open(ctx)
		if err != nil {
			seedlog.Errorf("Failed to open db for hit count: %v", err)
			return
		}
		defer func() {
			err := db.Close()
			if err != nil {
				seedlog.Errorf("close db failed: %v", err)
			}
		}()
		err = golinkdb.IncrementHitCount(ctx, db, key)
		if err != nil {
			seedlog.Errorf("Failed to increment hit count: %v", err)
		}
	}()
}
