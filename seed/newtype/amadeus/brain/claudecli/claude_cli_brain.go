package claudecli

import (
	"context"
	"os"
	"path/filepath"
	"strings"

	"github.com/ndscm/theseed/seed/infra/error/go/seederr"
	"github.com/ndscm/theseed/seed/infra/flag/go/seedflag"
	"github.com/ndscm/theseed/seed/newtype/amadeus/brain"
	"github.com/ndscm/theseed/seed/newtype/gajetto/proto/brainpb"
)

var flagClaudeCliTopicHome = seedflag.DefineString("claude_cli_topic_home", "~/topic/", "Path to topic home")

type claudeCliBrain struct {
	topicHome string
}

func NewClaudeCliBrain() brain.Brain {
	return &claudeCliBrain{}
}

func (b *claudeCliBrain) Initialize() error {
	topicHome := flagClaudeCliTopicHome.Get()
	if strings.HasPrefix(topicHome, "~/") {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return seederr.Wrap(err)
		}
		topicHome = filepath.Join(homeDir, topicHome[2:])
	}
	b.topicHome = topicHome
	return nil
}

func (b *claudeCliBrain) RegisterStepHandler(
	topic string, handler brain.BrainStepHandler,
) error {
	panic("RegisterStepHandler is not implemented")
}

func (b *claudeCliBrain) Input(ctx context.Context, topic string, input *brainpb.BrainInput) error {
	panic("Input is not implemented")
}

func (b *claudeCliBrain) Hibernate(topic string, wait bool) error {
	panic("Hibernate is not implemented")
}
