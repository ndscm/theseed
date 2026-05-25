import { type DirConfig } from "@//seed/devprod/ndscm/config/DIR"

export default {
  bootstrap: {
    proto: {
      target: [
        // sort
        "certificatepb",
        "certificatepbconnect",
      ],
      watch: "\\.proto$",
      bazel: {
        build: ":bootstrap",
        run: "{{BAZEL_RUN}}",
      },
    },
  },
} satisfies DirConfig
