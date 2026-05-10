import { type DirConfig } from "../../../devprod/ndscm/config/DIR"

export default {
  bootstrap: {
    ent: {
      target: "ent",
      watch: "^schema/",
      run: "bazel run :bootstrap",
    },
  },
} satisfies DirConfig
