# FaaS-Oriented Development: Make your service FaaS friendly

Most services in theseed ecosystem should be designed to be friendly to
Function-as-a-Service infrastructure (e.g., Cloud Run, AWS Lambda).
FaaS-friendliness is an architectural health check: a service that survives
scale-to-zero, cold starts, and infrastructure swaps is also faster and cheaper
on bare metal.

## Rules

1. **Maximize scalability.** Write handlers as if an arbitrary number of copies
   of the process may run concurrently and any individual copy may disappear at
   any time. Keep request handling stateless and push shared state out to
   external systems instead of in-process memory or local disk.
2. **Minimize binary size.** Treat every new dependency as weight the binary has
   to carry forever. Reach for the standard library first and pull in third-
   party modules only when they pay for themselves; favor small, focused
   libraries over frameworks that drag in transitive trees.
3. **Shorten the bootstrap process.** Prefer components that are cheap to
   construct. The server must still be fully ready by the time it reports ready,
   so the way to keep startup fast is to avoid pulling in subsystems whose setup
   is inherently heavy, not to hide that work behind lazy paths.
4. **Design for portability.** Build against standard HTTP or gRPC interfaces
   and read configuration from flags and environment variables, so the same
   binary runs unchanged on FaaS, a container orchestrator, or a bare-metal
   host.

## Pros

- **Lower cost.** Scale-to-zero idle behavior and small, lightweight instances
  mean the service only consumes resources proportional to actual traffic.
- **Infrastructure portability.** The same binary moves between FaaS, container
  orchestrators, and bare metal without code changes.
- **Fast cold starts.** A small dependency footprint and lightweight components
  keep startup latency low without resorting to lazy initialization tricks.
- **Predictable scaling.** Stateless handlers let the platform add or remove
  instances freely without coordinating in-process state.

## Cons

- **Dependency discipline.** Treating every import as permanent weight slows
  down feature work that would otherwise grab a convenient framework.
- **Externalized state cost.** Pushing state to a database or cache adds a
  network hop that an in-process cache would avoid.
- **Component selection pressure.** Ruling out subsystems with heavy setup
  narrows the menu of libraries and services the codebase can adopt.
