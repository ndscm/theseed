package main

import (
	"context"
	"encoding/json"
	"os"

	"github.com/ndscm/theseed/seed/infra/error/go/seederr"
	"github.com/ndscm/theseed/seed/infra/flag/go/seedflag"
	"github.com/ndscm/theseed/seed/infra/init/go/initctx"
	"github.com/ndscm/theseed/seed/infra/init/go/seedinit"
	"github.com/ndscm/theseed/seed/infra/log/go/seedlog"
	"github.com/ndscm/theseed/seed/newtype/gajetto/proto/brainpb"
	"github.com/ndscm/theseed/seed/newtype/hooin/dictate/client/go/dictateclient"
)

var flagPerson = seedflag.DefineString("person", "", "Target person identifier")
var flagInputType = seedflag.DefineString("type", "", "BrainInput type discriminator")
var flagInputTopic = seedflag.DefineString("topic", "", "BrainInput topic")
var flagInputText = seedflag.DefineString("text", "", "BrainInput text content")
var flagStream = seedflag.DefineBool("stream", false, "Stream every step")

func sendUnary(ctx context.Context, client *dictateclient.HooinDictateClient, input *brainpb.BrainInput) error {
	step, err := client.SendBrainInput(ctx, flagPerson.Get(), input)
	if err != nil {
		return seederr.Wrap(err)
	}
	data := step.GetData()
	pretty, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		seedlog.Errorf("Failed to marshal step data: %v", err)
		pretty = []byte("<failed to marshal data>")
	}
	seedlog.Infof("[%s] Result: %s", step.Topic, pretty)
	return nil
}

func sendStream(ctx context.Context, client *dictateclient.HooinDictateClient, input *brainpb.BrainInput) error {
	stream, err := client.SendBrainInputStreamBrainStep(ctx, flagPerson.Get(), input)
	if err != nil {
		return seederr.Wrap(err)
	}
	defer stream.Close()

	for stream.Receive() {
		step := stream.Msg()
		data := step.GetData()
		pretty, err := json.MarshalIndent(data, "", "  ")
		if err != nil {
			seedlog.Errorf("Failed to marshal step data: %v", err)
			pretty = []byte("<failed to marshal data>")
		}
		seedlog.Infof("[%s] Step (%s): %s", step.Topic, step.Type, pretty)
	}
	if err := stream.Err(); err != nil {
		return seederr.Wrap(err)
	}
	return nil
}

func run() error {
	_, err := seedinit.Initialize(
		seedinit.WithFallbackEnvPrefix("SEED_"),
	)
	if err != nil {
		return seederr.Wrap(err)
	}
	ctx := initctx.Background()
	client := dictateclient.NewHooinDictateClient("")

	input := &brainpb.BrainInput{
		Type:  flagInputType.Get(),
		Topic: flagInputTopic.Get(),
		Text:  flagInputText.Get(),
	}

	if flagStream.Get() {
		return sendStream(ctx, client, input)
	}
	return sendUnary(ctx, client, input)
}

func main() {
	err := run()
	if err != nil {
		seedlog.Errorf("%v", err)
		os.Exit(1)
	}
}
