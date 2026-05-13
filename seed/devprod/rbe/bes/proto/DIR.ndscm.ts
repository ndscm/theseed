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
      run: "bazel run :bootstrap",
    },
  },
} satisfies DirConfig
