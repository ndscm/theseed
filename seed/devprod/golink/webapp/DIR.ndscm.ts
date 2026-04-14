import { type DirConfig } from "../../ndscm/config/DIR"

export default {
  bootstrap: {
    node_modules: {
      target: "node_modules",
      watch: "^package.json$",
      run: 'bazel run @pnpm//:pnpm -- --dir "$(pwd)" install',
    },
  },
  build: {
    dist: {
      target: "dist",
      watch: "\\.(ts|tsx|js|jsx|json|css|html)$",
      run: 'bazel run @pnpm//:pnpm -- --dir "$(pwd)" build',
    },
  },
} satisfies DirConfig
