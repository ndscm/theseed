package seedcors

import (
	"net/http"
	"strings"

	"github.com/ndscm/theseed/seed/infra/flag/go/seedflag"
	"github.com/ndscm/theseed/seed/infra/log/go/seedlog"
	"github.com/rs/cors"
)

var flagCorsOrigins = seedflag.DefineString("cors_origins", "https://www.ndscm.com", "")
var flagCorsMethods = seedflag.DefineString("cors_methods", "GET,POST,PUT,DELETE,OPTIONS", "")

func InterceptCorsMiddleware(next http.Handler) http.Handler {
	corsOrigins := strings.Split(flagCorsOrigins.Get(), ",")
	seedlog.Infof("Cors origins: %v", corsOrigins)
	corsMethods := strings.Split(flagCorsMethods.Get(), ",")
	seedlog.Infof("Cors methods: %v", corsMethods)
	c := cors.New(cors.Options{
		AllowedOrigins:   corsOrigins,
		AllowedMethods:   corsMethods,
		AllowedHeaders:   []string{"*"},
		AllowCredentials: true,
		MaxAge:           60,
		Debug:            true,
	})
	return c.Handler(next)
}
