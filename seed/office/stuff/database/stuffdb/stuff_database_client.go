package stuffdb

import (
	"context"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"entgo.io/ent/dialect"
	_ "github.com/lib/pq"
	"github.com/ndscm/theseed/seed/infra/error/go/seederr"
	"github.com/ndscm/theseed/seed/infra/flag/go/seedflag"
	"github.com/ndscm/theseed/seed/infra/log/go/seedlog"
	"github.com/ndscm/theseed/seed/office/stuff/database/ent"
)

var flagStuffDatabase = seedflag.DefineString("stuff_database", "postgres://127.0.0.1:5432/stuff", "Database connection URL")
var flagStuffDatabaseDebug = seedflag.DefineBool("stuff_database_debug", false, "Enable debug mode for database")
var flagStuffDatabaseLogin = seedflag.DefineString("stuff_database_login", "stuff", "Database login user")
var flagStuffDatabaseSecretFile = seedflag.DefineString("stuff_database_secret_file", "${ND_USER_SECRET_HOME}/stuff/STUFF_DATABASE_SECRET", "Path to database password file")

func Open(ctx context.Context) (*ent.Client, error) {
	connectUrl, err := url.Parse(flagStuffDatabase.Get())
	if err != nil {
		return nil, seederr.Wrap(err)
	}
	seedlog.Debugf("Connecting to database: %s", connectUrl.Host+connectUrl.Path)
	databaseSecretPath := flagStuffDatabaseSecretFile.Get()
	if strings.HasPrefix(databaseSecretPath, "~/") {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return nil, seederr.Wrap(err)
		}
		databaseSecretPath = filepath.Join(homeDir, databaseSecretPath[2:])
	}
	databaseSecret, err := os.ReadFile(databaseSecretPath)
	if err != nil {
		return nil, seederr.Wrap(err)
	}
	databaseSecret = []byte(strings.TrimSpace(string(databaseSecret)))
	connectUrl.User = url.UserPassword(flagStuffDatabaseLogin.Get(), string(databaseSecret))
	db, err := ent.Open(dialect.Postgres, connectUrl.String())
	if err != nil {
		return nil, seederr.Wrap(err)
	}
	seedlog.Infof("Connected to database: %s", connectUrl.Host+connectUrl.Path)
	if flagStuffDatabaseDebug.Get() {
		db = db.Debug()
	}
	return db, nil
}
