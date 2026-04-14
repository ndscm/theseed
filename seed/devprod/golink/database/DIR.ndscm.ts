import { type DirConfig } from "../../ndscm/config/DIR"

export default {
  bootstrap: {
    ent: {
      target: "ent",
      watch: "^schema/",
      run: "./build.sh",
    },
  },
} satisfies DirConfig
