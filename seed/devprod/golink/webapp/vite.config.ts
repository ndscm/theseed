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
      "/seed.devprod.golink.proto.GolinkService/": {
        target: process.env.GOLINK_SERVICE_SERVER || "",
      },
    },
  },
})
