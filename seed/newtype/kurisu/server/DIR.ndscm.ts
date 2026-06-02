import { type DirConfig } from "@//seed/devprod/ndscm/config/DIR"

export default {
  bootstrap: {
    webapp: {
      target: "webapp",
      watch: "\\.go$",
      bazel: {
        build: ":bootstrap",
        run: "{{BAZEL_RUN}}",
      },
    },
  },
} satisfies DirConfig
