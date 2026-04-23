package main

import (
	"os"
	"strings"

	"github.com/ndscm/theseed/seed/infra/error/go/seederr"
	"github.com/ndscm/theseed/seed/infra/flag/go/seedflag"
	"github.com/ndscm/theseed/seed/infra/init/go/initctx"
	"github.com/ndscm/theseed/seed/infra/init/go/seedinit"
	"github.com/ndscm/theseed/seed/infra/log/go/seedlog"
	"github.com/ndscm/theseed/seed/newtype/hooin/dictate/client/go/dictateclient"
	"github.com/ndscm/theseed/seed/newtype/hooin/dictate/proto/dictatepb"
	"google.golang.org/grpc/codes"
)

// flagTopics is a comma-separated list of `person:topic` pairs.
// An omitted topic matches any topic for that person.
var flagTopics = seedflag.DefineString("topics", ":", "Comma-separated list of person:topic pairs, leave empty for wildcard (e.g. `:,christina:,:develop`)")

func parsePersonTopics(spec string) ([]*dictatepb.PersonTopic, error) {
	spec = strings.TrimSpace(spec)
	if spec == "" {
		return nil, seederr.CodeErrorf(codes.InvalidArgument, "--topics is required")
	}
	personTopics := []*dictatepb.PersonTopic{}
	for _, entry := range strings.Split(spec, ",") {
		personId, topic, ok := strings.Cut(entry, ":")
		if !ok {
			return nil, seederr.CodeErrorf(codes.InvalidArgument, "invalid person_topic entry: %q", entry)
		}
		personTopics = append(personTopics, &dictatepb.PersonTopic{
			PersonId: strings.TrimSpace(personId),
			Topic:    strings.TrimSpace(topic),
		})
	}
	if len(personTopics) == 0 {
		return nil, seederr.CodeErrorf(codes.InvalidArgument, "--topics is required")
	}
	return personTopics, nil
}

func run() error {
	_, err := seedinit.Initialize(
		seedinit.WithFallbackEnvPrefix("SEED_"),
	)
	if err != nil {
		return seederr.Wrap(err)
	}
	personTopics, err := parsePersonTopics(flagTopics.Get())
	if err != nil {
		return seederr.Wrap(err)
	}

	ctx := initctx.Background()
	client := dictateclient.NewHooinDictateClient("")

	stream, err := client.SubscribeBrainStep(ctx, personTopics)
	if err != nil {
		return seederr.Wrap(err)
	}
	defer stream.Close()

	for stream.Receive() {
		seedlog.Infof("BrainStep: %v", stream.Msg())
	}
	if err := stream.Err(); err != nil {
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
