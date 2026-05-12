import { type DirConfig } from "../../ndscm/config/DIR"

export default {
  bootstrap: {
    ent: {
      target: "ent",
      watch: "^schema/",
      needBazelBuild: ":bootstrap",
      run: "{{BAZEL_RUN}}",
    },
  },
} satisfies DirConfig
