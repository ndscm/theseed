import { defineConfig } from "i18next-cli"

export default defineConfig({
  locales: ["en"],
  extract: {
    input: "**/*.{ts,tsx}",
    output: "locales/{{language}}/{{namespace}}.json",
  },
})
