# Seed, The

> Seed, The Start of a New Miracle

Theseed is an open-source monorepo boilerplate for starting a new project or a
new company. It is powered by the ndscm team with years of rich experience in
development productivity (devprod/engprod). It is built to adopt the latest
industry standards with no historical baggage. Choose ndscm/theseed if you:

- Want strong devprod support with minimal time investment and zero cost
- Don't have time to evaluate every technology choice individually
- Want common office systems bundled into a single deployment
- Are tired of granting permissions one by one to your harness team
- (For OCD) Want to keep tech debt as low as possible

Theseed gives you:

- A batteries-included monorepo boilerplate with a common tech stack managed by
  Bazel.
- An opinionated tech stack for developing and deploying services on FaaS
  infrastructure to keep deployment costs low.
- A collection of SOP documents in the software factory for onboarding harness
  coding factory workers.

Although theseed can be used standalone, we recommend trying our cloud hosting
platform first. [ndhub (Incoming)](https://ndhub.com) provides a cloud-hosted
harness team so you can skip the technical details and build and deploy
full-scale cloud apps with natural language in seconds.

## Tech Radar

Theseed ships with support for the following tech stacks.

### Tier 1

Tier 1 is the tech stack that ndscm uses to build apps day to day. It is also
the default choice for harness coders in the ndscm ecosystem:

- **Build system:** Bazel
- **Language:** Go (services and CLI)
- **Language:** Python (scripts)
- **Language:** TypeScript (web and Node scripts)
- **Language:** CC (performance-sensitive libraries)
- **Schema:** Ent
- **Schema:** Protobuf
- **RPC:** gRPC and gRPC-Web with ConnectRPC
- **Container:** Podman
- **FaaS:** Google Cloud Run
- **Database:** PostgreSQL with Google Cloud SQL
- **Cache:** Redis
- **DNS:** Cloudflare
- **Agent:** Claude CLI
- **Model:** Opus
- **Model:** Sonnet
- **Webapp:** React with React Router
- **UI:** MUI
- **I18n:** I18next
- **Mobile (Android):** Embedded Webapp
- **(Incoming) Mobile (iOS):** Embedded Webapp
- **(Incoming) UI:** daisyUI

Theseed also includes deployment scripts for these self-hosted apps:

- **Auth:** Keycloak
- **Auth IdP:** Google Cloud Identity
- **(Incoming) Workflow:** Jenkins
- **(Incoming) Drive:** Nextcloud
- **(Incoming) Mail:** Dovecot, Postfix, Rspamd, Roundcube
- **(Incoming) Site-to-site:** OpenWrt with L2TP

### Tier 2

Tier 2 is the tech stack that our sponsored partners use:

- **(Incoming) Language:** Java
- **(Incoming) FaaS:** AWS Lambda
- **(Incoming) VM:** AWS EC2
- **(Incoming) Model:** Codex

## ndhub

ndhub is our first-party cloud hosting platform. We aim to keep pricing fair
with a target gross margin of 20%. In practice that means:

- A cloud provider charges us $1 for storage — we charge $1.20
- A cloud provider charges us $1 for network — we charge $1.20
- A cloud provider charges us $1 for model usage — we charge $1.20
- The cloud provider charges more, we charge more
- The cloud provider charges less, we charge less

The surcharge covers:

- Service maintenance
- New-user promotions
- Price-fluctuation buffer
- Operational overhead (free metadata storage)
- Development costs (servers, agent usage, etc.)
- Team salaries

We hope you enjoy our work — and sponsor us by using ndhub ❤️

## License

Theseed project is licensed under the [MIT License](LICENSE).

By submitting a contribution, you agree to license your work under the same
License and must include a `Signed-off-by:` trailer in each commit to certify
your right to do so.

Directories (mainly in `*/vendor/*`) contain code sourced from other projects.
They must be under a permissive license and are distributed under their original
licenses. Each vendor directory contains its own license file.

In all cases the license that governs a file is the nearest `LICENSE` file found
by walking up the directory tree from that file.

## Acknowledgments

In addition to the tech stack listed above, theseed directly uses the following
tools and libraries:

- Abseil
- Aspect
- Atlas
- Black
- Boost
- CEL
- Clang
- Copilot
- Debian
- FastAPI
- FastMCP
- Fedora
- Gazelle
- Git
- GitHub
- Hedron
- Kafka
- Keycloak
- libpq
- LLVM
- Nextcloud
- Notistack
- Oh My Zsh
- OpenID
- pgx
- Playwright
- pnpm
- Prettier
- PyAutoGUI
- Pydantic
- PyPI
- PyTorch
- Rocky Linux
- SQLite
- Ubuntu
- uv
- Vite
- VS Code
- Zsh

Our design is also inspired by the following projects:

- .NET
- Angular
- Apache
- Babel
- Bootstrap
- Docker
- Envoy
- FUSE
- Gerrit
- Jetty
- Mailu
- Make
- MariaSQL
- MySQL
- Nginx
- Phabricator
- Review Board
- Selenium
- webpack
- Yarnpkg
