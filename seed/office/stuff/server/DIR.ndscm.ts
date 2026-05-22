import { type DirConfig } from "@//seed/devprod/ndscm/config/DIR"

export default {
  bootstrap: {
    webapp: {
      target: "webapp",
      // Only watch .go files — watching the full source tree triggers too many rebuilds.
      watch: "\\.go$",
      bazel: {
        build: ":bootstrap",
        run: "{{BAZEL_RUN}}",
      },
    },
  },
} satisfies DirConfig
