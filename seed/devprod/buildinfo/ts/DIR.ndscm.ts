import { type DirConfig } from "../../ndscm/config/DIR"

export default {
  bootstrap: {
    dist: {
      target: "dist",
      watchRepo: ["node_modules"],
      watch: "^src/",
      run: "./build.sh",
    },
  },
} satisfies DirConfig
