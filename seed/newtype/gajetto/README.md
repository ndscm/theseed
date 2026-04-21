# Project Gajetto

Gajetto is the shared schema and identity library for the seed agent system.

It owns no RPC service of its own. Instead, it defines the vocabulary that every
other service — Amadeus, Hooin, and downstream consumers — speaks: the brain
message types that flow between them, and the team/person model used to
authenticate agents.

## Scope

- Brain schemas in [`proto/brain.proto`](proto/brain.proto), compiled to Go
  under [`proto/brainpb/`](proto/brainpb/).
- Team and person abstractions in [`team/`](team/), including the default static
  loader [`team/staticteam/`](team/staticteam/).
- No RPC surface: Gajetto is imported, not dialed.

## Brain messages

### `BrainInput`

A `BrainInput` is one self-contained multimodal message fed **into** an agent's
brain.

| Field       | Purpose                                                                  |
| ----------- | ------------------------------------------------------------------------ |
| `uuid`      | Stable identifier for dedup and tracing.                                 |
| `timestamp` | When the input was produced at the source.                               |
| `type`      | Application-defined discriminator.                                       |
| `topic`     | Logical channel; also selects which brain on the agent handles it.       |
| `wait`      | If set, the caller expects the service to block through full processing. |
| `metadata`  | Free-form key/value pairs. `first-person` is reserved.                   |
| `text`      | Textual payload.                                                         |
| `files`     | Generic attachments, including images and videos that are _not_ prompts. |
| `visual`    | Visuals consumed directly on the brain's multimodal path.                |
| `audio`     | Audio consumed directly on the brain's multimodal path.                  |

`files` is the attachment bag; `visual` and `audio` mean "this is part of the
multimodal prompt." General media belongs in `files`.

### `BrainStep`

A `BrainStep` is one event emitted **by** the brain while it is processing — a
tool call, a partial completion, a reasoning checkpoint, and so on. Steps are
application-defined via `type`, scoped by `topic`, and carry arbitrary key/value
`data`.

Steps flow out of the agent through Hooin:

- `HooinDictateService.SendBrainInput` returns the terminal `result` step for a
  submitted input.
- `HooinDictateService.SendBrainInputStreamBrainStep` streams every step for one
  input until the `result` closes the stream.
- `HooinDictateService.SubscribeBrainStep` streams steps for any input matching
  the requested `PersonTopic`s.

Amadeus does not carry `BrainStep`s on its RPC surface.

### Attachment types

`FileInput`, `VisualInput`, and `AudioInput` share the same shape: `uuid`,
`timestamp`, `path`, `mime`, `metadata`, and `data`.

## Team and identity

Agents authenticate to Hooin with a token that resolves to a `person_id`. The
resolution is performed against a [`team.Team`](team/team.go) implementation:

```go
type Team interface {
    GetHandle() string
    GetDisplayName() string
    GetMember(personId string) (Person, bool)
    Auth(token string) (personId string, err error)
}
```

### `staticteam`

The default loader reads a JSON file whose path is controlled by
`--static_team_file` (default `/etc/gajetto/team.json`). A leading `~/` is
expanded to the user's home. The file shape:

```json
{
  "handle": "example",
  "displayName": "Example Team",
  "members": {
    "alice": { "handle": "alice", "displayName": "Alice", "token": "…" },
    "bob": { "handle": "bob", "token": "…" }
  }
}
```

The map key for each member is the `person_id` returned by `Auth`.

## Design principles

- **Transport-agnostic.** Gajetto types are pure data; any service can carry
  them.
- **Forward-compatible.** `type`, `topic`, and `metadata` are free-form so new
  behaviors can ship without schema churn.
- **Explicit MIME, explicit UUIDs.** No inference and no implicit generation at
  the schema level — callers own both.
- **No upward dependencies.** Gajetto depends only on well-known Google protobuf
  types (`Empty`, `Timestamp`) and must not import Amadeus or Hooin.
