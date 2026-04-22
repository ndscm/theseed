package main

import (
	"os"

	"github.com/ndscm/theseed/seed/infra/error/go/seederr"
	"github.com/ndscm/theseed/seed/infra/flag/go/seedflag"
	"github.com/ndscm/theseed/seed/infra/init/go/initctx"
	"github.com/ndscm/theseed/seed/infra/init/go/seedinit"
	"github.com/ndscm/theseed/seed/infra/log/go/seedlog"
	"github.com/ndscm/theseed/seed/newtype/amadeus/wake/client/go/wakeclient"
)

var flagHooinDirectServer = seedflag.DefineString("hooin_direct_server", "", "Hooin instance address")
var flagToken = seedflag.DefineString("token", "", "Commute token for authentication")

func run() error {
	_, err := seedinit.Initialize(
		seedinit.WithFallbackEnvPrefix("SEED_"),
	)
	if err != nil {
		return seederr.Wrap(err)
	}
	ctx := initctx.Background()
	client := wakeclient.NewAmadeusWakeClient("")
	err = client.Wake(ctx, flagHooinDirectServer.Get(), flagToken.Get())
	if err != nil {
		return seederr.Wrap(err)
	}
	return nil
}

func main() {
	err := run()
	if err != nil {
		seedlog.Errorf("%v", err)
		os.Exit(1)
	}
}
