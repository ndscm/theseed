import { type DirConfig } from "//seed/devprod/ndscm/config/DIR"

export default {
  bootstrap: {
    dist: {
      target: "dist",
      watchRepo: ["node_modules", "seed/devprod/golink/proto/golink_pb.js"],
      watch: "^src/",
      bazel: {
        build: "@pnpm//:pnpm",
        run: 'BAZEL_BINDIR="$(dirname "{{BAZEL_EXECUTABLE}}")" {{BAZEL_RUN}} --dir "$(pwd)" build',
      },
    },
  },
} satisfies DirConfig
