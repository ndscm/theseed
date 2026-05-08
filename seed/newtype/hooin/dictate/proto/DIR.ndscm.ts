import { type DirConfig } from "../../../../devprod/ndscm/config/DIR"

export default {
  bootstrap: {
    proto: {
      target: ["dictatepb", "dictatepbconnect"],
      watch: "\\.proto$",
      run: "./build.sh",
    },
  },
} satisfies DirConfig
