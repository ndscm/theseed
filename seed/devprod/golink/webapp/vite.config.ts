import { reactRouter } from "@react-router/dev/vite"
import { defineConfig } from "vite"
import tsconfigPaths from "vite-tsconfig-paths"

const DEFAULT_LANGUAGE = process.env.DEFAULT_LANGUAGE || "en"
const BUILD_LANGUAGE = process.env.BUILD_LANGUAGE || DEFAULT_LANGUAGE

export default defineConfig({
  base: BUILD_LANGUAGE == DEFAULT_LANGUAGE ? "/" : `/${BUILD_LANGUAGE}/`,
  plugins: [reactRouter(), tsconfigPaths()],
  define: {
    BUILD_LANGUAGE: JSON.stringify(BUILD_LANGUAGE),
  },
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
