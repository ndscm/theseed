import { type DirConfig } from "../../../../../devprod/ndscm/config/DIR"

export default {
  bootstrap: {
    dist: {
      target: "dist",
      watchRepo: ["node_modules", "seed/office/stuff/proto/stuff_pb.js"],
      watch: "^src/",
      run: 'bazel run @pnpm//:pnpm -- --dir "$(pwd)" build',
    },
  },
} satisfies DirConfig
