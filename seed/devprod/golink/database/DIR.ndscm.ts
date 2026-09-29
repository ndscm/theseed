import { type DirConfig } from "@//seed/devprod/ndscm/config/DIR"

export default {
  granularity: {
    style: "linux",
    prefix: "seed: golink: database: ",
  },
  bootstrap: {
    ent: {
      target: "ent",
      watch: "^schema/",
      bazel: {
        build: ":bootstrap",
        run: "{{BAZEL_RUN}}",
      },
    },
  },
} satisfies DirConfig
