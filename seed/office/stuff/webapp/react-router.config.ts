import fs from "node:fs/promises"
import path from "node:path"

import { type Config } from "@react-router/dev/config"

const DEFAULT_LANGUAGE = process.env.DEFAULT_LANGUAGE || "en"
const BUILD_LANGUAGE = process.env.BUILD_LANGUAGE || DEFAULT_LANGUAGE

export default {
  basename: BUILD_LANGUAGE == DEFAULT_LANGUAGE ? "/" : `/${BUILD_LANGUAGE}`,
  buildDirectory: "dist/" + BUILD_LANGUAGE,
  ssr: false,
  prerender: async () => {
    return [
      // sort
      "/",
    ]
  },

  /**
   * Hack: Flatten the localized build output to align React Router with Vite's asset pipeline.
   *
   * Problem:
   * When `basename` is defined (e.g., "/es/"), React Router's SSG engine strictly honors it
   * and outputs pre-rendered HTML into a nested directory (e.g., `dist/es/client/es/`).
   * However, Vite ignores the basename for physical file generation. It outputs the `assets/`
   * directory, `public/` files (like favicon.ico), and `__spa-fallback.html` directly to
   * the root (`dist/es/client/`). This results in a fractured build where HTML and assets
   * are physically misaligned.
   *
   * Solution:
   * We hook into the end of the build to flatten the output directory. By pulling the
   * pre-rendered HTML files up into the client root, everything sits perfectly alongside
   * Vite's native assets.
   *
   * The Go backend can then safely mount the flattened `dist/es/client` directory
   * directly to the `/es/` URL route.
   */
  buildEnd: async ({ reactRouterConfig }) => {
    if (BUILD_LANGUAGE == DEFAULT_LANGUAGE) {
      return
    }
    const clientDir = path.join(reactRouterConfig.buildDirectory, "client")
    const langDir = path.join(clientDir, BUILD_LANGUAGE)
    const items = await fs.readdir(langDir).catch(() => [])
    for (const item of items) {
      await fs.rename(path.join(langDir, item), path.join(clientDir, item))
    }
    if (items.length > 0) {
      await fs.rmdir(langDir)
    }
  },
} satisfies Config
