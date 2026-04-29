package seedctx

import (
	"context"
	"os"
	"os/exec"
	"strings"

	"github.com/ndscm/theseed/seed/infra/error/go/seederr"
	"github.com/ndscm/theseed/seed/infra/flag/go/seedflag"
)

var flagPerformer = seedflag.DefineString("performer", "", "Performer identity for client context. If set, --performer_command is ignored.")
var flagPerformerCommand = seedflag.DefineString("performer_command", "",
	`Shell command whose stdout is used as the performer. Used only when --performer is empty.
WARNING: the value is executed via "sh -c"; only set this from a trusted source.
Binaries that opt into seedflag.WithFallbackEnvPrefix make this flag environment-controllable, which would let any writer of that env run arbitrary shell.`,
)

func Performer(ctx context.Context) (string, error) {
	if ctx == nil {
		return "", seederr.WrapErrorf("nil context provided")
	}
	performer, ok := ctx.Value(SeedContextKey("performer")).(string)
	if !ok {
		return "", seederr.WrapErrorf("performer not found in context")
	}
	return performer, nil
}

func WithPerformer(parent context.Context, performer string) context.Context {
	return context.WithValue(parent, SeedContextKey("performer"), performer)
}

func Background() context.Context {
	performer := flagPerformer.Get()
	performerCommand := flagPerformerCommand.Get()
	if performer == "" && performerCommand != "" {
		cmd := exec.Command("sh", "-c", performerCommand)
		cmd.Stderr = os.Stderr
		output, err := cmd.Output()
		if err != nil {
			panic(seederr.WrapErrorf("failed to run performer command %q: %v", performerCommand, err))
		}
		performer = strings.TrimSpace(string(output))
	}
	ctx := context.Background()
	if performer == "" {
		panic(seederr.WrapErrorf("no performer configured: set --performer or --performer_command"))
	}
	ctx = WithPerformer(ctx, performer)
	return ctx
}
