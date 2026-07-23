// Types in this file mirror the line envelopes of Claude CLI stream-json
// mode: `claude --input-format stream-json --output-format stream-json`.
//
// They describe the shape of the `BrainStep.data` google.protobuf.Struct
// (see seed/newtype/gajetto/proto/brain.proto) that the Go claudepayload
// package produces (see payload.go), so the frontend can parse it. Field names
// match the JSON tags emitted by that package.

export interface StreamInputEnvelope {
  type: string
}

export interface StreamInputMessage {
  role: string
  content: string
}

export interface StreamInputUser extends StreamInputEnvelope {
  message?: StreamInputMessage
}

export interface StreamOutputEnvelope {
  type: string
  subtype?: string
  session_id?: string
}

export interface StreamOutputSystem extends StreamOutputEnvelope {
  type: "system"
  cwd?: string
  model?: string
  tools?: string[]
  permissionMode?: string
  apiKeySource?: string
}

export interface StreamContentBlock {
  type: string
  text?: string
  id?: string
  name?: string
  input?: unknown
  tool_use_id?: string
  content?: unknown
  is_error?: boolean
  thinking?: string
  signature?: string
}

export interface StreamOutputMessage {
  id?: string
  type?: string
  role?: string
  model?: string
  content?: StreamContentBlock[]
  stop_reason?: string
  stop_sequence?: string
  usage?: StreamOutputUsage
}

export interface StreamOutputAssistant extends StreamOutputEnvelope {
  type: "assistant"
  message?: StreamOutputMessage
  parent_tool_use_id?: string
}

export interface StreamOutputUser extends StreamOutputEnvelope {
  type: "user"
  message?: StreamOutputMessage
  parent_tool_use_id?: string
}

export interface StreamOutputUsage {
  input_tokens?: number
  cache_creation_input_tokens?: number
  cache_read_input_tokens?: number
  output_tokens?: number
  server_tool_use?: unknown
  service_tier?: string
}

export interface StreamOutputResult extends StreamOutputEnvelope {
  type: "result"
  is_error: boolean
  duration_ms?: number
  duration_api_ms?: number
  num_turns?: number
  result?: string
  total_cost_usd?: number
  usage?: StreamOutputUsage
}

// StreamOutput is the discriminated union of the recognized stream-json output
// lines, keyed by a literal `type`. It deliberately omits the bare
// StreamOutputEnvelope: a member with an open `type: string` would overlap every
// literal and defeat narrowing on `.type`.
export type StreamOutput =
  | StreamOutputSystem
  | StreamOutputAssistant
  | StreamOutputUser
  | StreamOutputResult

// decodeStreamInput narrows a parsed `BrainStep.data` struct into the typed
// input envelope matching its `type` field, or undefined for an unrecognized
// type.
//
// It is the frontend counterpart of the StreamInputUser payload the Go
// claudepayload package echoes back as a "claudecli-input" step (see
// writeInput in topic_runner.go). Only the "user" variant exists today, so
// this mirrors DecodeStreamOutput but over the single-member input union.
export const DecodeStreamInput = (
  data: unknown,
): StreamInputUser | undefined => {
  if (data == null || typeof data !== "object") {
    return undefined
  }
  const envelope = data as Partial<StreamInputEnvelope>
  if (typeof envelope.type !== "string") {
    return undefined
  }
  switch (envelope.type) {
    case "user":
      return envelope as StreamInputUser
    default:
      return undefined
  }
}

// decodeStreamOutput narrows a parsed `BrainStep.data` struct into the typed
// envelope matching its `type` field, or undefined for an unrecognized type.
//
// It is the frontend counterpart of DecodeStreamOutputData in payload.go: the
// struct has already been decoded from JSON, so this only selects the variant.
// Unlike the Go, which returns a bare envelope for unknown types, this returns
// undefined so callers narrow cleanly against the four known variants.
export const DecodeStreamOutput = (data: unknown): StreamOutput | undefined => {
  // `data` is BrainStep.data, an optional google.protobuf.Struct, so protobuf-es
  // hands us undefined when a step carries no payload. The cast below is
  // compile-time only, so reading `.type` off a non-object throws at runtime and
  // skips the `default: undefined` path callers rely on.
  if (data == null || typeof data !== "object") {
    return undefined
  }
  const envelope = data as Partial<StreamOutputEnvelope>
  if (typeof envelope.type !== "string") {
    return undefined
  }
  switch (envelope.type) {
    case "system":
      return envelope as StreamOutputSystem
    case "assistant":
      return envelope as StreamOutputAssistant
    case "user":
      return envelope as StreamOutputUser
    case "result":
      return envelope as StreamOutputResult
    default:
      return undefined
  }
}

export default {
  DecodeStreamInput,
  DecodeStreamOutput,
}
