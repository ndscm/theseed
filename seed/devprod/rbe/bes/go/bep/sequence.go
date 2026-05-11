package bep

import (
	"bytes"
	"encoding/binary"

	"github.com/ndscm/theseed/seed/devprod/rbe/bes/proto/buildeventstreampb"
	"github.com/ndscm/theseed/seed/infra/error/go/seederr"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

type deterministicEventId string

var deterministicMarshaller = proto.MarshalOptions{Deterministic: true}

func deterministic(id *buildeventstreampb.BuildEventId) (deterministicEventId, error) {
	b, err := deterministicMarshaller.Marshal(id)
	if err != nil {
		return "", seederr.Wrap(err)
	}
	return deterministicEventId(b), nil
}

type BuildEventSequence struct {
	eventMap map[deterministicEventId]*buildeventstreampb.BuildEvent

	eventSequence []*buildeventstreampb.BuildEventId
}

func (evs *BuildEventSequence) Get(id *buildeventstreampb.BuildEventId) (*buildeventstreampb.BuildEvent, error) {
	deterministicId, err := deterministic(id)
	if err != nil {
		return nil, seederr.Wrap(err)
	}
	event, ok := evs.eventMap[deterministicId]
	if !ok {
		return nil, seederr.WrapErrorf("event not found for id: %v", id)
	}
	return event, nil
}

func (evs *BuildEventSequence) Search(matcher func(*buildeventstreampb.BuildEventId) bool) []*buildeventstreampb.BuildEvent {
	results := []*buildeventstreampb.BuildEvent{}
	for _, id := range evs.eventSequence {
		if matcher(id) {
			deterministicId, err := deterministic(id)
			if err != nil {
				continue
			}
			results = append(results, evs.eventMap[deterministicId])
		}
	}
	return results
}

func (evs *BuildEventSequence) Len() int {
	return len(evs.eventSequence)
}

func (evs *BuildEventSequence) DumpJson(pretty bool) ([]byte, error) {
	marshaller := protojson.MarshalOptions{}
	if pretty {
		marshaller.Multiline = true
		marshaller.Indent = "  "
	}

	jsonEvents := [][]byte{}
	for _, id := range evs.eventSequence {
		deterministicId, err := deterministic(id)
		if err != nil {
			return nil, seederr.Wrap(err)
		}
		event := evs.eventMap[deterministicId]
		jsonEvent, err := marshaller.Marshal(event)
		if err != nil {
			return nil, seederr.Wrap(err)
		}
		jsonEvents = append(jsonEvents, jsonEvent)
	}

	result := []byte("[")
	if pretty {
		result = append(result, '\n')
		result = append(result, bytes.Join(jsonEvents, []byte(",\n"))...)
		result = append(result, '\n')
	} else {
		result = append(result, bytes.Join(jsonEvents, []byte(","))...)
	}
	result = append(result, ']')
	if pretty {
		result = append(result, '\n')
	}
	return result, nil
}

func ParseBuildEventProtos(data []byte) (*BuildEventSequence, error) {
	events := &BuildEventSequence{
		eventMap: map[deterministicEventId]*buildeventstreampb.BuildEvent{},
	}
	for len(data) > 0 {
		msgLen, n := binary.Uvarint(data)
		if n <= 0 {
			return nil, seederr.WrapErrorf("failed to decode varint at offset %d", len(data))
		}
		data = data[n:]
		if uint64(len(data)) < msgLen {
			return nil, seederr.WrapErrorf("truncated message: need %d bytes, have %d", msgLen, len(data))
		}
		event := &buildeventstreampb.BuildEvent{}
		err := proto.Unmarshal(data[:msgLen], event)
		if err != nil {
			return nil, seederr.Wrap(err)
		}
		deterministicId, err := deterministic(event.GetId())
		if err != nil {
			return nil, seederr.Wrap(err)
		}
		_, ok := events.eventMap[deterministicId]
		if ok {
			return nil, seederr.WrapErrorf("duplicate event id found: %v", event.GetId())
		}
		events.eventSequence = append(events.eventSequence, event.GetId())
		events.eventMap[deterministicId] = event
		data = data[msgLen:]
	}
	return events, nil
}
