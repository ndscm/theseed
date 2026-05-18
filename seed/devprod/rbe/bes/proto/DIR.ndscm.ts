import { type DirConfig } from "//seed/devprod/ndscm/config/DIR"

export default {
  bootstrap: {
    proto: {
      target: [
        "actioncachepb",
        "buildeventstreampb",
        "commandlinepb",
        "failuredetailspb",
        "invocationpolicypb",
        "optionfilterspb",
        "packageloadmetricspb",
        "strategypolicypb",
      ],
      watch: "^BUILD.bazel$",
      bazel: {
        build: ":bootstrap",
        run: "{{BAZEL_RUN}}",
      },
    },
  },
} satisfies DirConfig
