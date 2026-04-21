# Project Hooin

Hooin is the online agent registry and routing hub for the seed agent system.

Where **Amadeus** manages the lifecycle of a single agent container, Hooin sits
above the fleet. It is the meeting point where external producers submit
`BrainInput`s, where agent containers commute in to pick up the inputs addressed
to them, and where agents report the `BrainStep`s they emit back for fan-out to
observers.

## System architecture

```
  ┌──────────────┐                      ┌──────────────────┐
  │  Management  │        Wake          │     Amadeus      │
  │    System    │ ───────────────────▶ │   (in agent      │
  │              │      Hibernate       │    container)    │
  └──────────────┘                      └──┬────────────┬──┘
                                           ▲            │
                          Commute          │            │  ReportBrainStep
                       stream BrainInput   │            │  BrainStep
                                           ▼            ▼
  ┌──────────────┐    SendBrainInput    ┌──┴────────────┴──┐
  │              │ ───────────────────▶ │                  │
  │      UI      │ ◀── BrainStep ────── │      Hooin       │
  │              │  SubscribeBrainStep  │                  │
  │              │ ◀── stream ───────── │                  │
  └──────────────┘                      └──────────────────┘
```

Every agent container runs an **Amadeus** instance. On `Wake`, Amadeus opens a
`Commute` stream to Hooin using its registered token, and Hooin starts pushing
`BrainInput`s addressed to that agent's person down the stream. As the agent
processes inputs, it publishes each emitted `BrainStep` back to Hooin via
`ReportBrainStep`, which fans it out to matching subscribers.

## Concepts

### Person

A `person_id` identifies an agent instance. Hooin routes inputs and steps by
`person_id` and maintains the mapping from `person_id` to the currently
commuting agent container.

### PersonTopic

`PersonTopic { person_id, topic }` is the addressing tuple for subscription
fan-out. A single person can emit on multiple topics, and observers can
subscribe to any combination. An empty `person_id` matches every person; an
empty `topic` matches every topic for the selected person(s).

### Token

`CommuteRequest.token` and `ReportBrainStepRequest.token` authenticate the
agent. Tokens are resolved by the
[gajetto `Team`](../gajetto/README.md#team-and-identity) loaded at server
startup and tie an authenticating agent to exactly one `person_id`.

## RPC surface

Hooin ships two services on the same port (default `4664`), multiplexed via
`seedgrpc.GrpcMux`. Both are defined in
[`commute/proto/commute.proto`](commute/proto/commute.proto) and
[`dictate/proto/dictate.proto`](dictate/proto/dictate.proto). Brain message
types come from
[`//seed/newtype/gajetto/proto:brain.proto`](../gajetto/proto/brain.proto).

### `HooinCommuteService` — agent side

#### `Commute(CommuteRequest) → stream BrainInput`

The agent-facing stream. An Amadeus instance authenticates with its token and
receives every `BrainInput` routed to its person for the life of the session.

#### `ReportBrainStep(ReportBrainStepRequest) → Empty`

Unary publish. The commuting agent calls this for each `BrainStep` it emits.
Hooin authenticates the token, resolves it to a `person_id`, and fans the step
out to every `SubscribeBrainStep` / `SendBrainInputStreamBrainStep` subscriber
whose `PersonTopic` matches.

### `HooinDictateService` — external side

#### `SendBrainInput(SendBrainInputRequest) → BrainStep`

Unary submit. Routes a single `BrainInput` to the agent currently on duty for
`person_id` and returns the terminal `result` `BrainStep` once processing
finishes. Callers that want the full stream should use
`SendBrainInputStreamBrainStep` or subscribe separately via
`SubscribeBrainStep`.

#### `SendBrainInputStreamBrainStep(SendBrainInputRequest) → stream BrainStep`

Submit and stream. Delivers one `BrainInput` and streams every `BrainStep`
produced while processing it.

#### `SubscribeBrainStep(SubscribeBrainStepRequest) → stream BrainStep`

Long-lived subscription. Streams `BrainStep`s reported for any input processed
against the requested `PersonTopic`s, regardless of who submitted the input.

## Design notes

- **Submit and observe are separate.** `SendBrainInput*` is for producers that
  want per-input response semantics; `SubscribeBrainStep` is for dashboards,
  recorders, and auditors that care about the stream of events, not individual
  call results.
- **Steps come from agents, not Hooin.** Hooin doesn't synthesize `BrainStep`s —
  agents report them via `ReportBrainStep`, and Hooin is strictly a fan-out hub.
  This keeps the authoritative step source co-located with the processing logic.
- **One duty slot per person.** Hooin claims the duty slot atomically on
  `Commute` and rejects overlapping reconnects rather than silently replacing
  the live session.
- **gRPC-Web friendly.** All streaming RPCs are server-streaming only, so they
  work over gRPC, Connect, and gRPC-Web (browsers included).
- **Routing is Hooin's job.** Producers address inputs by `person_id` and don't
  need to know which container services that person. Agent binding happens
  implicitly at `Commute` time.
- **Token → person, not token → session.** A token identifies a registered
  person. Commute and ReportBrainStep both authenticate with the same token, so
  a reporting agent is always the one currently on duty for that person.

## Clients and tooling

Go clients live under
[`commute/client/go/commuteclient`](commute/client/go/commuteclient) and
[`dictate/client/go/dictateclient`](dictate/client/go/dictateclient). They
default to `http://127.0.0.1:4664` via the `--hooin_commute_service_server` and
`--hooin_dictate_service_server` flags.

Two command-line utilities live under [`cmd/`](cmd):

- [`cmd/send`](cmd/send/send.go) — submit a single `BrainInput` (unary or
  streaming via `--stream`).
- [`cmd/subscribe`](cmd/subscribe/subscribe.go) — open a `SubscribeBrainStep` on
  a comma-separated list of `person:topic` pairs.
