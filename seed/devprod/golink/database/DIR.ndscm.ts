import { type DirConfig } from "../../ndscm/config/DIR"

export default {
  bootstrap: {
    ent: {
      target: "ent",
      watch: "^schema/",
      run: "bazel run :bootstrap",
    },
  },
} satisfies DirConfig
