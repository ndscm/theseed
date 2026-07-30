# Code Quality

<!-- SOP guidelines don't apply to this doc -->

Writing code that simply runs without bugs falls far short of good code in the
theseed ecosystem. Beyond executability, we also measure code quality by three
abilities: readability, testability, and reviewability.

## Readability

Good code should always be self-describing. Self-describing code dramatically
cuts down on the need for comments. Good readability saves the author and future
readers the time of reconstructing the original reasoning behind the code. It
also lowers the real cost of maintaining the code and avoids the eventual
rewrite of an entire module that unreadable implementation invites.

## Testability

Testing has always been the responsibility of the original author, because only
the author understands every use case and every side effect worth considering.
Testability is a first-class concern from the moment the code is written.
Tossing testing over the wall to some kind of "test" team is not how we work in
the theseed ecosystem.

Testing against private functions is discouraged. When a developer feels the
need to test a private function, it usually signals that the struct isn't well
designed for its callers. Tests do more than build confidence that the code is
bug-free; they also document the original design intent of the module for its
users. Working through the friction of writing tests gives the developer a
better understanding of how to use the module.

Testing against mocks is discouraged in the theseed ecosystem. Some argue that
mocks isolate a test's scope and cut down on unrelated failures. But testing
against mocks burns test runs on scenarios that don't exist and hides real
failures from the layers above. Tests in a monorepo don't serve just one team;
they give the teams that depend on you the confidence to make changes. So
overusing mocks is not recommended. The one case where mocks are appropriate is
when the interface is extremely stable and the real logic is extremely slow
(e.g. a database).

## Reviewability

It's not enough for the code to end up in a good final state — its history
should be in good shape too. That means keeping each commit as small as
possible: do exactly one thing per commit. Small commits help reviewers
understand a change faster on the way to the initial merge into main, and they
save time on `blame` when a bug surfaces in the prod tier.

ndscm is designed to manage patches for better reviewability in the theseed
ecosystem. The commit metadata that ndscm manages (e.g. Break, Migrate,
Side-effect-of-change-uuid) drives the other tools that review the code (e.g.
the CI system). Keeping commit granularity small is not as hard as developers
expect.
