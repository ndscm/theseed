import { type DirConfig } from "../../../devprod/ndscm/config/DIR"

export default {
  bootstrap: {
    proto: {
      target: [
        "stuffpb",
        "stuffpbconnect",
        "stufferrorpb",
        "stuff_pb.js",
        "stuff_pb.d.ts",
        "stuff_error_pb.js",
        "stuff_error_pb.d.ts",
      ],
      watch: "\\.proto$",
      needBazelBuild: ":bootstrap",
      run: "{{BAZEL_RUN}}",
    },
  },
} satisfies DirConfig
