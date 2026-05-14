import { type DirConfig } from "../../../../devprod/ndscm/config/DIR"

export default {
  bootstrap: {
    proto: {
      target: [
        //
        "commutepb",
        "commutepbconnect",
      ],
      watch: "\\.proto$",
      needBazelBuild: ":bootstrap",
      run: "{{BAZEL_RUN}}",
    },
  },
} satisfies DirConfig
