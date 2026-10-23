import { type UpstreamConfig } from "@//seed/devprod/ndscm/config/UPSTREAM"

export default {
  xslash: {
    converge: "melt",
    scm: "git",
    repo: "git@github.com:ndscm/xslash.git",
    tracking: "main",
  },
} satisfies UpstreamConfig
