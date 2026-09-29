import { type DirConfig } from "@//seed/devprod/ndscm/config/DIR"

export default {
  granularity: {
    style: "linux",
    prefix: "seed: stuff: webapp: ",
  },
  build: {
    dist: {
      target: "dist",
      watchRepo: ["node_modules"],
      watch: "\\.(ts|tsx|js|jsx|json|css|html)$",
      run: 'bazel run @pnpm//:pnpm -- --dir "$(pwd)" build',
    },
  },
} satisfies DirConfig
