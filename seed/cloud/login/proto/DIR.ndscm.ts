import { type DirConfig } from "../../../devprod/ndscm/config/DIR"

export default {
  bootstrap: {
    proto: {
      target: [
        //
        "login_pb.d.ts",
        "login_pb.js",
        "loginpb",
        "loginpbconnect",
      ],
      watch: "\\.proto$",
      run: "./build.sh",
    },
  },
} satisfies DirConfig
