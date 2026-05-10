import { type DirConfig } from "../../../../../devprod/ndscm/config/DIR"

export default {
  bootstrap: {
    node_modules: {
      target: "node_modules",
      watch: "^package.json$",
      run: 'bazel run @pnpm//:pnpm -- --dir "$(pwd)" install',
    },
    dist: {
      target: "dist",
      watchRepo: "seed/office/stuff/proto/stuff_pb.js",
      watch: "^src/",
      run: 'bazel run @pnpm//:pnpm -- --dir "$(pwd)" build',
    },
  },
} satisfies DirConfig
