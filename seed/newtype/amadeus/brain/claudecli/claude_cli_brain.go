package claudecli

import (
	"context"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"

	"github.com/ndscm/theseed/seed/infra/error/go/seederr"
	"github.com/ndscm/theseed/seed/infra/flag/go/seedflag"
	"github.com/ndscm/theseed/seed/newtype/amadeus/brain"
	"github.com/ndscm/theseed/seed/newtype/gajetto/proto/brainpb"
)

var flagClaudeCliTopicHome = seedflag.DefineString("claude_cli_topic_home", "~/topic/", "Path to topic home")

type claudeCliBrain struct {
	topicHome string

	topicsMutex sync.Mutex
	topics      map[string]*topicRunner
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
	b.topics = map[string]*topicRunner{}
	return nil
}

// validTopicName restricts topic names to lowercase alphanumerics with
// internal dashes or underscores. This prevents path traversal via
// filepath.Join(topicHome, topic) since "..", "/", and "\" are rejected.
var validTopicName = regexp.MustCompile(`^[a-z0-9]+([_-][a-z0-9]+)*$`)

// getTopic returns the topicRunner for topic. If create is false and the
// topic has never been started, it returns (nil, nil) so callers like
// Hibernate can no-op instead of spawning a claude subprocess just to
// tear it back down.
func (b *claudeCliBrain) getTopic(topic string, create bool) (*topicRunner, error) {
	if !validTopicName.MatchString(topic) {
		return nil, seederr.WrapErrorf("invalid topic name: %q", topic)
	}

	b.topicsMutex.Lock()
	defer b.topicsMutex.Unlock()

	t, ok := b.topics[topic]
	if ok {
		return t, nil
	}
	if !create {
		return nil, nil
	}

	topicDir := filepath.Join(b.topicHome, topic)
	err := os.MkdirAll(topicDir, 0755)
	if err != nil {
		return nil, seederr.Wrap(err)
	}
	tr, err := newTopicRunner(topic, topicDir)
	if err != nil {
		return nil, seederr.Wrap(err)
	}
	b.topics[topic] = tr
	return tr, nil
}

func (b *claudeCliBrain) RegisterStepHandler(
	topic string, handler brain.BrainStepHandler,
) error {
	t, err := b.getTopic(topic, true)
	if err != nil {
		return seederr.Wrap(err)
	}
	err = t.RegisterStepHandler(handler)
	if err != nil {
		return seederr.Wrap(err)
	}
	return nil
}

func (b *claudeCliBrain) Input(ctx context.Context, topic string, input *brainpb.BrainInput) error {
	t, err := b.getTopic(topic, true)
	if err != nil {
		return seederr.Wrap(err)
	}
	return t.Input(ctx, input)
}

func (b *claudeCliBrain) Hibernate(topic string, wait bool) error {
	t, err := b.getTopic(topic, false)
	if err != nil {
		return seederr.Wrap(err)
	}
	if t == nil {
		return nil
	}
	err = t.Hibernate(wait)
	if err != nil {
		return seederr.Wrap(err)
	}
	return nil
}
