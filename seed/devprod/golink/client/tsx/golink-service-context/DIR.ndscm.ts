import { type DirConfig } from "../../../../ndscm/config/DIR"

export default {
  bootstrap: {
    node_modules: {
      target: "node_modules",
      watch: "^package.json$",
      run: 'bazel run @pnpm//:pnpm -- --dir "$(pwd)" install',
    },
    dist: {
      target: "dist",
      watchRepo: "seed/devprod/golink/proto/golink_pb.js",
      watch: "^src/",
      run: 'bazel run @pnpm//:pnpm -- --dir "$(pwd)" build',
    },
  },
} satisfies DirConfig
