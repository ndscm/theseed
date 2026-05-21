import { type Config } from "@react-router/dev/config"

const DEFAULT_LANGUAGE = process.env.DEFAULT_LANGUAGE || "en"
const BUILD_LANGUAGE = process.env.BUILD_LANGUAGE || DEFAULT_LANGUAGE

const config: Config = {
  buildDirectory:
    "dist" + (BUILD_LANGUAGE == DEFAULT_LANGUAGE ? "" : `/${BUILD_LANGUAGE}`),
  ssr: false,
  async prerender() {
    return ["/"]
  },
}

export default config
