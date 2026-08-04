package steinsdb

import (
	"context"
	"database/sql"
	"net/url"
	"time"

	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/ndscm/theseed/seed/infra/error/go/seederr"
	"github.com/ndscm/theseed/seed/infra/flag/go/seedflag"
	"github.com/ndscm/theseed/seed/infra/log/go/seedlog"
	"github.com/ndscm/theseed/seed/newtype/steins/database/ent"
)

var flagSteinsDatabase = seedflag.DefineString(
	"steins_database", "postgres://127.0.0.1:5432/steins",
	"Database connection URL",
)
var flagSteinsDatabaseDebug = seedflag.DefineBool(
	"steins_database_debug", false,
	"Enable debug mode for database",
)
var flagSteinsDatabaseLogin = seedflag.DefineString(
	"steins_database_login", "steins",
	"Database login user",
)
var flagSteinsDatabaseSecret = seedflag.DefineSecret(
	"steins_database_secret",
	"Steins database password",
)

const dbMaxOpenConns = 25
const dbMaxIdleConns = 5
const dbConnMaxLifetime = 30 * time.Minute
const dbConnMaxIdleTime = 5 * time.Minute

func Open(ctx context.Context) (*ent.Client, error) {
	connectUrl, err := url.Parse(flagSteinsDatabase.Get())
	if err != nil {
		return nil, seederr.Wrap(err)
	}
	seedlog.Debugf("Connecting to database: %s", connectUrl.Host+connectUrl.Path)
	databaseLogin := flagSteinsDatabaseLogin.Get()
	databaseSecret, err := flagSteinsDatabaseSecret.LoadString()
	if err != nil {
		return nil, seederr.Wrap(err)
	}
	connectUrl.User = url.UserPassword(databaseLogin, databaseSecret)
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
	if flagSteinsDatabaseDebug.Get() {
		client = client.Debug()
	}
	return client, nil
}
