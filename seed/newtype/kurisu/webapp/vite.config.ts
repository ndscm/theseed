import { reactRouter } from "@react-router/dev/vite"
import tailwindcss from "@tailwindcss/vite"
import { defineConfig } from "vite"
import tsconfigPaths from "vite-tsconfig-paths"

import unsafeDevLogin from "../../../cloud/login/ts/vite-plugin-unsafe-dev-login"

const DEFAULT_LANGUAGE = process.env.DEFAULT_LANGUAGE || "en"
const BUILD_LANGUAGE = process.env.BUILD_LANGUAGE || DEFAULT_LANGUAGE

export default defineConfig({
  base: BUILD_LANGUAGE == DEFAULT_LANGUAGE ? "/" : `/${BUILD_LANGUAGE}/`,
  plugins: [
    // sort
    reactRouter(),
    tailwindcss(),
    tsconfigPaths(),
    unsafeDevLogin({ clientId: "kurisu-webapp-dev" }),
  ],
  define: {
    BUILD_LANGUAGE: JSON.stringify(BUILD_LANGUAGE),
  },
  server: {
    proxy: {
      "/seed.cloud.login.proto.LoginService/": {
        target: process.env.LOGIN_SERVICE_SERVER || "",
      },
      "/seed.newtype.hooin.dictate.proto.HooinDictateService/": {
        target: process.env.HOOIN_DICTATE_SERVICE_SERVER || "",
      },
      "/seed.newtype.hooin.invade.proto.HooinInvadeService/": {
        target: process.env.HOOIN_INVADE_SERVICE_SERVER || "",
      },
      "/seed.newtype.hooin.roster.proto.HooinRosterService/": {
        target: process.env.HOOIN_ROSTER_SERVICE_SERVER || "",
      },
      "/seed.newtype.kurisu.proto.KurisuService/": {
        target: process.env.KURISU_SERVICE_SERVER || "",
      },
    },
  },
})
