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
      run: "./build.sh",
    },
  },
} satisfies DirConfig
