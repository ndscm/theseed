package sqlsessiondb

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
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/ndscm/theseed/seed/cloud/sqlsession/database/ent"
	"github.com/ndscm/theseed/seed/infra/error/go/seederr"
	"github.com/ndscm/theseed/seed/infra/flag/go/seedflag"
	"github.com/ndscm/theseed/seed/infra/log/go/seedlog"
)

var flagSqlSessionDatabase = seedflag.DefineString("sqlsession_database", "postgres://127.0.0.1:5432/session", "Session database connection URL")
var flagSqlSessionDatabaseDebug = seedflag.DefineBool("sqlsession_database_debug", false, "Enable debug mode for session database")
var flagSqlSessionDatabaseLogin = seedflag.DefineString("sqlsession_database_login", "session", "Session database login user")
var flagSqlSessionDatabaseSecretFile = seedflag.DefineString("sqlsession_database_secret_file", "", "Path to session database password file")

const (
	dbMaxOpenConns    = 25
	dbMaxIdleConns    = 5
	dbConnMaxLifetime = 30 * time.Minute
	dbConnMaxIdleTime = 5 * time.Minute
)

func Open(ctx context.Context) (*ent.Client, error) {
	connectUrl, err := url.Parse(flagSqlSessionDatabase.Get())
	if err != nil {
		return nil, seederr.Wrap(err)
	}
	seedlog.Debugf("Connecting to database: %s", connectUrl.Host+connectUrl.Path)
	databaseSecretPath := flagSqlSessionDatabaseSecretFile.Get()
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
	connectUrl.User = url.UserPassword(flagSqlSessionDatabaseLogin.Get(), string(databaseSecret))
	db, err := sql.Open("pgx", connectUrl.String())
	if err != nil {
		return nil, seederr.Wrap(err)
	}

	db.SetMaxOpenConns(dbMaxOpenConns)
	db.SetMaxIdleConns(dbMaxIdleConns)
	db.SetConnMaxLifetime(dbConnMaxLifetime)
	db.SetConnMaxIdleTime(dbConnMaxIdleTime)

	client := ent.NewClient(ent.Driver(entsql.OpenDB(dialect.Postgres, db)))
	seedlog.Infof("Connected to database: %s", connectUrl.Host+connectUrl.Path)
	if flagSqlSessionDatabaseDebug.Get() {
		client = client.Debug()
	}
	return client, nil
}
