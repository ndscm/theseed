import { type DirConfig } from "../../ndscm/config/DIR"

export default {
  bootstrap: {
    proto: {
      target: [
        //
        "golink_error_pb.d.ts",
        "golink_error_pb.js",
        "golink_pb.d.ts",
        "golink_pb.js",
        "golinkerrorpb",
        "golinkpb",
        "golinkpbconnect",
      ],
      watch: "\\.proto$",
      run: "bazel run :bootstrap",
    },
  },
} satisfies DirConfig
