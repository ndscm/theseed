import { type DirConfig } from "//seed/devprod/ndscm/config/DIR"

export default {
  build: {
    dist: {
      target: "dist",
      watchRepo: ["node_modules"],
      watch: "\\.(ts|tsx|js|jsx|json|css|html)$",
      run: 'bazel run @pnpm//:pnpm -- --dir "$(pwd)" build',
    },
  },
} satisfies DirConfig
