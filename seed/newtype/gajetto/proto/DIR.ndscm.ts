import { type DirConfig } from "@//seed/devprod/ndscm/config/DIR"

export default {
  bootstrap: {
    proto: {
      target: [
        // sort
        "brain_pb.d.ts",
        "brain_pb.js",
        "brainpb",
      ],
      watch: "\\.proto$",
      bazel: {
        build: ":bootstrap",
        run: "{{BAZEL_RUN}}",
      },
    },
  },
} satisfies DirConfig
