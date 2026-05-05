import { reactRouter } from "@react-router/dev/vite"
import { defineConfig } from "vite"
import tsconfigPaths from "vite-tsconfig-paths"

export default defineConfig({
  plugins: [reactRouter(), tsconfigPaths()],
  resolve: {
    alias: {
      "opentype.js/dist/opentype.module.js": "opentype.js/dist/opentype.mjs",
      "opentype.js/dist/opentype.module": "opentype.js/dist/opentype.mjs",
    },
  },
  build: {
    assetsDir: ".assets",
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
        target: process.env.STUFF_SERVICE_SERVER || "",
      },
      "/seed.office.stuff.proto.StuffService/": {
        target: process.env.STUFF_SERVICE_SERVER || "",
      },
    },
  },
})
