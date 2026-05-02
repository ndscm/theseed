# Project Amadeus

Amadeus is the agent lifecycle management service.

One Amadeus instance runs **inside each agent container** and exposes a small
control surface for driving that container's agent: wake it, restart its brain,
and put it back to sleep. `BrainStep`s emitted by the agent while it is awake
are published through the Hooin service the agent commutes to — never through
Amadeus.

A brain is the configured CLI interface for the underlying LLM. The brain is
selected by the `--brain` flag; today only `claudecli` ships. If a single brain
is registered, the flag can be omitted.

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

The Wake / RestartBrain / Hibernate loop is driven against Amadeus by whoever
orchestrates the container (today typically a human via the `wake` / `hibernate`
CLIs). Hooin, once commuted to, owns the channel that carries `BrainInput`s into
the agent. `BrainStep`s flow back out through Hooin to any interested UIs and
subscribers — not through the Amadeus control RPCs.

## RPC surface

All messages are defined in [`wake/proto/wake.proto`](wake/proto/wake.proto).
Brain message types come from
[`//seed/newtype/gajetto/proto:brain.proto`](../gajetto/proto/brain.proto).

The service is served on port `2623` by default (`--port`, word: `AMAD`).

### `Wake(WakeRequest) → Empty`

Starts the agent and returns once it has commuted to the specified Hooin service
and is ready to process inputs.

On invocation, Amadeus:

1. Opens a `HooinCommuteService.Commute` stream to
   `WakeRequest.hooin_direct_server`, authenticating with `WakeRequest.token`.
2. Starts forwarding `BrainInput`s from Hooin into the brain, creating a
   per-topic brain runner on first input for each topic.

`BrainStep`s emitted by the brain are published through Hooin via
`ReportBrainStep`.

### `RestartBrain(RestartBrainRequest) → Empty`

- `hot_upgrade = true`: preserve in-memory state across the restart where
  possible.
- `wait = true`: block until every runner is ready again.

### `Hibernate(HibernateRequest) → Empty`

Stops the agent, cancels the commute stream to Hooin, and returns the container
to a quiescent state. A fresh `Wake` is required to resume work.

- `wait = true`: block until hibernation has completed.

## Brain layer

A **brain** is a pluggable implementation of [`brain.Brain`](brain/brain.go).
Brains self-register at package init; today `claudecli` is the only shipped
implementation.

### `claudecli`

Runs the `claude` CLI in `stream-json` mode per topic. State:

- `--claude_cli_topic_home` (default `~/topic/`) picks the parent directory that
  houses one subdirectory per topic. A leading `~/` is expanded to the user's
  home.
- The first `BrainInput` for a topic spawns a `claude` subprocess in
  `stream-json` mode in that topic's directory.
- The subprocess's stdout is parsed per line into `BrainStep`s and reported back
  to Hooin via `ReportBrainStep`.

The subprocess is started with `--permission-mode=bypassPermissions`. See
[Security boundary](#security-boundary) below for what that implies.

## Clients and tooling

The Go client lives under
[`wake/client/go/wakeclient`](wake/client/go/wakeclient) and defaults to
`http://127.0.0.1:2623` via the `--amadeus_service_server` flag.

Two command-line utilities live under [`cmd/`](cmd):

- [`cmd/wake`](cmd/wake/wake.go) — call `Wake` with `--hooin_direct_server` and
  `--token`.
- [`cmd/hibernate`](cmd/hibernate/hibernate.go) — call `Hibernate`, optionally
  `--wait`.

## Design notes

- **One container, one Amadeus.** Amadeus is the control-plane endpoint for a
  single agent container. Fleet-level orchestration lives above it.
- **Amadeus does not persist.** Tokens, topic bindings, and brain state are all
  transient to the wake session. Durable routing and identity live in Hooin and
  higher-level registries.
- **Control plane only.** Amadeus carries no `BrainInput` or `BrainStep` on its
  RPC surface. All data flows through the commute channel into Hooin, and
  observers consume steps from Hooin.
- **Wake is exclusive; Hibernate is idempotent.** There is never more than one
  live commute session per Amadeus instance, so the control surface is simple
  and race-free at the fleet level.

## Security boundary

The `claudecli` brain runs each `claude` subprocess with
`--permission-mode=bypassPermissions`. Every interactive permission prompt the
CLI would normally raise — file edits, shell commands, network calls — is
suppressed and auto-approved.

This is intentional, and it is the only practical mode for an unattended agent:
there is no human at the keyboard to answer "allow this `rm -rf`?" prompts. The
trade-off is explicit: while the brain is awake, the model has the **full
file-system and shell access of the user the `claude` process runs as**, scoped
to whatever that user can reach inside the container.

The containment story therefore lives one level out, not at the CLI prompt:

- The agent container is the security boundary. It is treated as a hostile
  environment from the host's perspective and runs as a non-root user with no
  privileged capabilities beyond what its tooling explicitly needs.
- Secrets must not be mounted into the container unless the brain is expected to
  use them. Anything readable by the brain user is readable by the model.
- Outbound network reachability from the container is the model's outbound
  network reachability. Restrict it at the network layer, not at the CLI.

Operators wiring up new agent containers are responsible for that outer
boundary. Amadeus and `claudecli` deliberately do not try to re-implement
permission prompting on top of `bypassPermissions`.
