# [WIP] Kirintor Code Review System

Project Kirintor is a high-performance code review system for ndscm managed
monorepos.

## Alternatives

### GitHub Pull Requests

GitHub's pull request system was built to streamline merges. Getting it to fit
the ndscm workflow takes a lot of configuration, and all that setup is
error-prone.

Until Kirintor shipped, the ndscm team relied on GitHub as its primary code
review platform. That came with several limitations:

- No support for patch-based review (diff ranges). You can only review the whole
  PR diff.
- Reviews only move forward. There's no way to review a rebase.
- A tight rate limit chokes on the volume of CI events we generate.
- A hard cap of 100 commits per PR breaks our melt workflow.
- Pricing is per-seat, which gets far too expensive for our agent team.
