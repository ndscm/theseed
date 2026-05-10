import { type DirConfig } from "../../../../devprod/ndscm/config/DIR"

export default {
  bootstrap: {
    proto: {
      target: [
        //
        "wakepb",
        "wakepbconnect",
      ],
      watch: "\\.proto$",
      run: "bazel run :bootstrap",
    },
  },
} satisfies DirConfig
