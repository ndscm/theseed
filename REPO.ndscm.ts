import { type RepoConfig } from "@//seed/devprod/ndscm/config/REPO"

export default {
  domain: "ndscm.com",
  ndscm: {
    version: "26.4.14",
  },
  commit: {
    rewrite: [
      {
        pattern: "^f$",
        replace: "fixup",
      },
      {
        pattern: "^d$",
        replace: "debug",
      },
      {
        pattern: "^w$",
        replace: "wip",
      },
      {
        pattern: "^fixup",
        finish: true,
      },
      {
        pattern: "^debug",
        finish: true,
      },
      {
        pattern: "^wip",
        finish: true,
      },
      {
        pattern: "^(seed: )?(.*)$",
        replace: "seed: $2",
      },
    ],
  },
} satisfies RepoConfig
