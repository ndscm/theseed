import { type DirConfig } from "@//seed/devprod/ndscm/config/DIR"

export default {
  granularity: {
    style: "linux",
    prefix: "seed: gotcha: tsx: ",
  },
  bootstrap: {
    dist: {
      target: "dist",
      watch: "^src/",
      bazel: {
        build: ":bootstrap",
        run: "{{BAZEL_RUN}}",
      },
    },
  },
} satisfies DirConfig
