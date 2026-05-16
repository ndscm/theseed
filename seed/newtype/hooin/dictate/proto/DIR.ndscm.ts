import { type DirConfig } from "//seed/devprod/ndscm/config/DIR"

export default {
  bootstrap: {
    proto: {
      target: ["dictatepb", "dictatepbconnect"],
      watch: "\\.proto$",
      needBazelBuild: ":bootstrap",
      run: "{{BAZEL_RUN}}",
    },
  },
} satisfies DirConfig
