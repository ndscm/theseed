import { reactRouter } from "@react-router/dev/vite"
import { defineConfig } from "vite"
import tsconfigPaths from "vite-tsconfig-paths"

import unsafeDevLogin from "../../../cloud/login/ts/vite-plugin-unsafe-dev-login"

const DEFAULT_LANGUAGE = process.env.DEFAULT_LANGUAGE || "en"
const BUILD_LANGUAGE = process.env.BUILD_LANGUAGE || DEFAULT_LANGUAGE

export default defineConfig({
  base: BUILD_LANGUAGE == DEFAULT_LANGUAGE ? "/" : `/${BUILD_LANGUAGE}/`,
  plugins: [
    //sort
    reactRouter(),
    tsconfigPaths(),
    unsafeDevLogin({ clientId: "webapp-dev" }),
  ],
  define: {
    BUILD_LANGUAGE: JSON.stringify(BUILD_LANGUAGE),
  },
  build: {
    assetsDir: ".assets",
  },
  server: {
    proxy: {
      "/seed.cloud.login.proto.LoginService/": {
        target: process.env.LOGIN_SERVICE_SERVER || "",
      },
      "/seed.devprod.golink.proto.GolinkService/": {
        target: process.env.GOLINK_SERVICE_SERVER || "",
      },
    },
  },
})
