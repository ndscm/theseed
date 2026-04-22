package claudepayload

import (
	"encoding/json"

	"github.com/ndscm/theseed/seed/infra/error/go/seederr"
	"google.golang.org/protobuf/types/known/structpb"
)

// Types in this file mirror the line envelopes of Claude CLI stream-json
// mode: `claude --input-format stream-json --output-format stream-json`.

type StreamInputEnvelope struct {
	Type string `json:"type"`
}

type StreamInputMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type StreamInputUser struct {
	StreamInputEnvelope
	Message *StreamInputMessage `json:"message,omitempty"`
}

type StreamOutputEnvelope struct {
	Type      string `json:"type"`
	Subtype   string `json:"subtype,omitempty"`
	SessionId string `json:"session_id,omitempty"`
}

func (e StreamOutputEnvelope) Data() (*structpb.Struct, error) {
	b, err := json.Marshal(e)
	if err != nil {
		return nil, seederr.Wrap(err)
	}
	raw := map[string]any{}
	err = json.Unmarshal(b, &raw)
	if err != nil {
		return nil, seederr.Wrap(err)
	}
	s, err := structpb.NewStruct(raw)
	if err != nil {
		return nil, seederr.Wrap(err)
	}
	return s, nil
}

type StreamOutputSystem struct {
	StreamOutputEnvelope
	Cwd            string   `json:"cwd,omitempty"`
	Model          string   `json:"model,omitempty"`
	Tools          []string `json:"tools,omitempty"`
	PermissionMode string   `json:"permissionMode,omitempty"`
	ApiKeySource   string   `json:"apiKeySource,omitempty"`
}

func (s StreamOutputSystem) Data() (*structpb.Struct, error) {
	b, err := json.Marshal(s)
	if err != nil {
		return nil, seederr.Wrap(err)
	}
	raw := map[string]any{}
	err = json.Unmarshal(b, &raw)
	if err != nil {
		return nil, seederr.Wrap(err)
	}
	out, err := structpb.NewStruct(raw)
	if err != nil {
		return nil, seederr.Wrap(err)
	}
	return out, nil
}

type StreamContentBlock struct {
	Type      string          `json:"type"`
	Text      string          `json:"text,omitempty"`
	Id        string          `json:"id,omitempty"`
	Name      string          `json:"name,omitempty"`
	Input     json.RawMessage `json:"input,omitempty"`
	ToolUseId string          `json:"tool_use_id,omitempty"`
	Content   json.RawMessage `json:"content,omitempty"`
	IsError   bool            `json:"is_error,omitempty"`
	Thinking  string          `json:"thinking,omitempty"`
	Signature string          `json:"signature,omitempty"`
}

type StreamOutputMessage struct {
	Id           string               `json:"id,omitempty"`
	Type         string               `json:"type,omitempty"`
	Role         string               `json:"role,omitempty"`
	Model        string               `json:"model,omitempty"`
	Content      []StreamContentBlock `json:"content,omitempty"`
	StopReason   string               `json:"stop_reason,omitempty"`
	StopSequence string               `json:"stop_sequence,omitempty"`
	Usage        *StreamOutputUsage   `json:"usage,omitempty"`
}

type StreamOutputAssistant struct {
	StreamOutputEnvelope
	Message         *StreamOutputMessage `json:"message,omitempty"`
	ParentToolUseId string               `json:"parent_tool_use_id,omitempty"`
}

func (a StreamOutputAssistant) Data() (*structpb.Struct, error) {
	b, err := json.Marshal(a)
	if err != nil {
		return nil, seederr.Wrap(err)
	}
	raw := map[string]any{}
	err = json.Unmarshal(b, &raw)
	if err != nil {
		return nil, seederr.Wrap(err)
	}
	s, err := structpb.NewStruct(raw)
	if err != nil {
		return nil, seederr.Wrap(err)
	}
	return s, nil
}

type StreamOutputUser struct {
	StreamOutputEnvelope
	Message         *StreamOutputMessage `json:"message,omitempty"`
	ParentToolUseId string               `json:"parent_tool_use_id,omitempty"`
}

func (u StreamOutputUser) Data() (*structpb.Struct, error) {
	b, err := json.Marshal(u)
	if err != nil {
		return nil, seederr.Wrap(err)
	}
	raw := map[string]any{}
	err = json.Unmarshal(b, &raw)
	if err != nil {
		return nil, seederr.Wrap(err)
	}
	s, err := structpb.NewStruct(raw)
	if err != nil {
		return nil, seederr.Wrap(err)
	}
	return s, nil
}

type StreamOutputUsage struct {
	InputTokens              int64           `json:"input_tokens,omitempty"`
	CacheCreationInputTokens int64           `json:"cache_creation_input_tokens,omitempty"`
	CacheReadInputTokens     int64           `json:"cache_read_input_tokens,omitempty"`
	OutputTokens             int64           `json:"output_tokens,omitempty"`
	ServerToolUse            json.RawMessage `json:"server_tool_use,omitempty"`
	ServiceTier              string          `json:"service_tier,omitempty"`
}

type StreamOutputResult struct {
	StreamOutputEnvelope
	IsError       bool               `json:"is_error"`
	DurationMs    int64              `json:"duration_ms,omitempty"`
	DurationApiMs int64              `json:"duration_api_ms,omitempty"`
	NumTurns      int                `json:"num_turns,omitempty"`
	Result        string             `json:"result,omitempty"`
	TotalCostUsd  float64            `json:"total_cost_usd,omitempty"`
	Usage         *StreamOutputUsage `json:"usage,omitempty"`
}

func (r StreamOutputResult) Data() (*structpb.Struct, error) {
	b, err := json.Marshal(r)
	if err != nil {
		return nil, seederr.Wrap(err)
	}
	raw := map[string]any{}
	err = json.Unmarshal(b, &raw)
	if err != nil {
		return nil, seederr.Wrap(err)
	}
	s, err := structpb.NewStruct(raw)
	if err != nil {
		return nil, seederr.Wrap(err)
	}
	return s, nil
}

// DecodeStreamOutputData parses a single stream-json output line from the
// Claude CLI and returns its fields as a google.protobuf.Struct.
//
// The line is decoded into the typed envelope matching its `type` field, and
// that envelope's Data() is returned. Unknown types fall back to the bare
// StreamOutputEnvelope.
func DecodeStreamOutputData(line []byte) (string, *structpb.Struct, error) {
	envelope := StreamOutputEnvelope{}
	err := json.Unmarshal(line, &envelope)
	if err != nil {
		return "", nil, seederr.Wrap(err)
	}
	switch envelope.Type {
	case "system":
		data := StreamOutputSystem{}
		err := json.Unmarshal(line, &data)
		if err != nil {
			return "", nil, seederr.Wrap(err)
		}
		s, err := data.Data()
		if err != nil {
			return "", nil, seederr.Wrap(err)
		}
		return data.Type, s, nil
	case "assistant":
		data := StreamOutputAssistant{}
		err := json.Unmarshal(line, &data)
		if err != nil {
			return "", nil, seederr.Wrap(err)
		}
		s, err := data.Data()
		if err != nil {
			return "", nil, seederr.Wrap(err)
		}
		return data.Type, s, nil
	case "user":
		data := StreamOutputUser{}
		err := json.Unmarshal(line, &data)
		if err != nil {
			return "", nil, seederr.Wrap(err)
		}
		s, err := data.Data()
		if err != nil {
			return "", nil, seederr.Wrap(err)
		}
		return data.Type, s, nil
	case "result":
		data := StreamOutputResult{}
		err := json.Unmarshal(line, &data)
		if err != nil {
			return "", nil, seederr.Wrap(err)
		}
		s, err := data.Data()
		if err != nil {
			return "", nil, seederr.Wrap(err)
		}
		return data.Type, s, nil
	default:
		s, err := envelope.Data()
		if err != nil {
			return "", nil, seederr.Wrap(err)
		}
		return envelope.Type, s, nil
	}
}
