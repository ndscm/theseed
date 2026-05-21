import { defineConfig } from "i18next-cli"

export default defineConfig({
  locales: [
    "en", // en must be defined as the first language
    "es",
  ],
  extract: {
    input: "**/*.{ts,tsx}",
    output: "locales/{{language}}/{{namespace}}.json",
  },
})
