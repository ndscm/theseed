import { reactRouter } from "@react-router/dev/vite"
import { defineConfig } from "vite"
import tsconfigPaths from "vite-tsconfig-paths"

export default defineConfig({
  plugins: [reactRouter(), tsconfigPaths()],
  build: {
    assetsDir: ".assets",
  },
  server: {
    proxy: {
      "/seed.cloud.login.proto.LoginService/": {
        target: process.env.STUFF_SERVICE_SERVER || "",
      },
      "/seed.office.stuff.proto.StuffService/": {
        target: process.env.STUFF_SERVICE_SERVER || "",
      },
    },
  },
})
