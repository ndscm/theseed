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
  resolve: {
    alias: {
      "opentype.js/dist/opentype.module.js": "opentype.js/dist/opentype.mjs",
      "opentype.js/dist/opentype.module": "opentype.js/dist/opentype.mjs",
    },
  },
  ssr: {
    noExternal: ["@univerjs/*"],
  },
  optimizeDeps: {
    include: ["@univerjs/*"],
  },
  server: {
    proxy: {
      "/seed.cloud.login.proto.LoginService/": {
        target: process.env.LOGIN_SERVICE_SERVER || "",
      },
      "/seed.office.stuff.proto.StuffService/": {
        target: process.env.STUFF_SERVICE_SERVER || "",
      },
    },
  },
})
