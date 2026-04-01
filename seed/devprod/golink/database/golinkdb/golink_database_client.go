package golinkdb

import (
	"context"
	"database/sql"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"
	"github.com/ndscm/theseed/seed/devprod/golink/database/ent"
	"github.com/ndscm/theseed/seed/devprod/golink/proto/golinkerrorpb"
	"github.com/ndscm/theseed/seed/infra/flag/go/seedflag"
	"github.com/ndscm/theseed/seed/infra/error/go/seederr"
	"github.com/ndscm/theseed/seed/infra/log/go/seedlog"
	_ "github.com/jackc/pgx/v5/stdlib"
)

var flagGolinkDatabase = seedflag.DefineString("golink_database", "postgres://127.0.0.1:5432/golink", "Database connection URL")
var flagGolinkDatabaseDebug = seedflag.DefineBool("golink_database_debug", false, "Enable debug mode for database")
var flagGolinkDatabaseLogin = seedflag.DefineString("golink_database_login", "golink", "Database login user")
var flagGolinkDatabaseSecretFile = seedflag.DefineString("golink_database_secret_file", "${ND_USER_SECRET_HOME}/golink/GOLINK_DATABASE_SECRET", "Path to database password file")

const (
	dbMaxOpenConns    = 25
	dbMaxIdleConns    = 5
	dbConnMaxLifetime = 30 * time.Minute
	dbConnMaxIdleTime = 5 * time.Minute
)

func Open(ctx context.Context) (*ent.Client, error) {
	connectUrl, err := url.Parse(flagGolinkDatabase.Get())
	if err != nil {
		return nil, seederr.Code(golinkerrorpb.Code_INTERNAL_INVALID_DATABASE_URL, err)
	}
	seedlog.Debugf("Connecting to database: %s", connectUrl.Host+connectUrl.Path)
	databaseSecretPath := flagGolinkDatabaseSecretFile.Get()
	if strings.HasPrefix(databaseSecretPath, "~/") {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return nil, seederr.Code(golinkerrorpb.Code_INTERNAL_INVALID_USER_HOME, err)
		}
		databaseSecretPath = filepath.Join(homeDir, databaseSecretPath[2:])
	}
	databaseSecret, err := os.ReadFile(databaseSecretPath)
	if err != nil {
		return nil, seederr.Code(golinkerrorpb.Code_INTERNAL_INVALID_DATABASE_SECRET_FILE, err)
	}
	databaseSecret = []byte(strings.TrimSpace(string(databaseSecret)))
	connectUrl.User = url.UserPassword(flagGolinkDatabaseLogin.Get(), string(databaseSecret))
	db, err := sql.Open("pgx", connectUrl.String())
	if err != nil {
		return nil, seederr.Code(golinkerrorpb.Code_INTERNAL_OPEN_DATABASE_FAILED, err)
	}

	db.SetMaxOpenConns(dbMaxOpenConns)
	db.SetMaxIdleConns(dbMaxIdleConns)
	db.SetConnMaxLifetime(dbConnMaxLifetime)
	db.SetConnMaxIdleTime(dbConnMaxIdleTime)

	client := ent.NewClient(ent.Driver(entsql.OpenDB(dialect.Postgres, db)))
	seedlog.Infof("Connected to database: %s", connectUrl.Host+connectUrl.Path)
	if flagGolinkDatabaseDebug.Get() {
		client = client.Debug()
	}
	return client, nil
}
