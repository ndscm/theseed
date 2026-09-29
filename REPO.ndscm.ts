import { type RepoConfig } from "@//seed/devprod/ndscm/config/REPO"

export default {
  domain: "ndscm.com",
  collaboration: {
    tracker: {
      system: "github",
    },
    review: {
      system: "github",
    },
  },
} satisfies RepoConfig
