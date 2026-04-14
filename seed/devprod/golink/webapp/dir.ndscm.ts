import { type DirConfig } from "../../ndscm/config/dir"

export default {
  bootstrap: {
    pnpm: {
      target: "node_modules",
      watch: "^package.json$",
      run: "pnpm install",
    },
  },
} satisfies DirConfig
