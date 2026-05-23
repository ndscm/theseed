import { type DirConfig } from "@//seed/devprod/ndscm/config/DIR"

export default {
  bootstrap: {
    buildinfo: {
      // Generates stub status txt files to keep gopls happy.
      // It won't match the actual values from a real bazel build.
      target: [
        // sort
        "stable-status.txt",
        "volatile-status.txt",
      ],
      watch: "^BUILD.bazel$",
      bazel: {
        build: ":bootstrap",
        run: "{{BAZEL_RUN}}",
      },
    },
  },
} satisfies DirConfig
