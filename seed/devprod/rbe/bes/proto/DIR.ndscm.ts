import { type DirConfig } from "../../../ndscm/config/DIR"

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
      needBazelBuild: ":bootstrap",
      run: "{{BAZEL_RUN}}",
    },
  },
} satisfies DirConfig
