import { type DirConfig } from "@//seed/devprod/ndscm/config/DIR"

export default {
  bootstrap: {
    proto: {
      target: [
        "kurisupb",
        "kurisupbconnect",
        "kurisuerrorpb",
        "kurisu_pb.js",
        "kurisu_pb.d.ts",
        "kurisu_error_pb.js",
        "kurisu_error_pb.d.ts",
      ],
      watch: "\\.proto$",
      bazel: {
        build: ":bootstrap",
        run: "{{BAZEL_RUN}}",
      },
    },
  },
} satisfies DirConfig
