import { type DirConfig } from "//seed/devprod/ndscm/config/DIR"

export default {
  bootstrap: {
    dist: {
      target: "dist",
      watchRepo: ["node_modules"],
      watch: "^src/",
      bazel: {
        build: ":bootstrap",
        run: "{{BAZEL_RUN}}",
      },
    },
  },
} satisfies DirConfig
