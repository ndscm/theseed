package golinkdb

import (
	"context"
	"database/sql"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/ndscm/theseed/seed/infra/flag/go/seedflag"
	"github.com/ndscm/theseed/seed/infra/error/go/seederr"
	"github.com/ndscm/theseed/seed/infra/log/go/seedlog"
	_ "github.com/jackc/pgx/v5/stdlib"
)

var flagGolinkDatabase = seedflag.DefineString("golink_database", "postgres://127.0.0.1:5432/golink", "Database connection URL")
var flagGolinkDatabaseLogin = seedflag.DefineString("golink_database_login", "golink", "Database login user")
var flagGolinkDatabaseSecretFile = seedflag.DefineString("golink_database_secret_file", "${ND_USER_SECRET_HOME}/golink/GOLINK_DATABASE_SECRET", "Path to database password file")

const (
	dbMaxOpenConns    = 25
	dbMaxIdleConns    = 5
	dbConnMaxLifetime = 30 * time.Minute
	dbConnMaxIdleTime = 5 * time.Minute
)

func Open(ctx context.Context) (*sql.DB, error) {
	connectUrl, err := url.Parse(flagGolinkDatabase.Get())
	if err != nil {
		return nil, seederr.Wrap(err)
	}
	databaseSecretPath := flagGolinkDatabaseSecretFile.Get()
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
	connectUrl.User = url.UserPassword(flagGolinkDatabaseLogin.Get(), string(databaseSecret))
	db, err := sql.Open("pgx", connectUrl.String())
	if err != nil {
		return nil, seederr.Wrap(err)
	}

	db.SetMaxOpenConns(dbMaxOpenConns)
	db.SetMaxIdleConns(dbMaxIdleConns)
	db.SetConnMaxLifetime(dbConnMaxLifetime)
	db.SetConnMaxIdleTime(dbConnMaxIdleTime)

	if err := db.PingContext(ctx); err != nil {
		return nil, seederr.Wrap(err)
	}
	seedlog.Infof("Connected to database: %s", connectUrl.Host+connectUrl.Path)
	return db, nil
}
