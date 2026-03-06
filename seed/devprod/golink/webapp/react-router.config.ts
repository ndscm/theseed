import { type Config } from "@react-router/dev/config"

const config: Config = {
  buildDirectory: "dist",
  ssr: false,
  async prerender() {
    return ["/"]
  },
}

export default config
