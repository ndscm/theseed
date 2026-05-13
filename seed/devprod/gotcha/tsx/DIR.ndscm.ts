import { type DirConfig } from "../../ndscm/config/DIR"

export default {
  bootstrap: {
    dist: {
      target: "dist",
      watchRepo: ["node_modules"],
      watch: "^src/",
      run: 'bazel run @pnpm//:pnpm -- --dir "$(pwd)" build',
    },
  },
} satisfies DirConfig
