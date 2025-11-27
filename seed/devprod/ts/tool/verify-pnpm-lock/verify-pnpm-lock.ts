import fs from "node:fs/promises"
import process from "node:process"

import DebugModule from "debug"
import JsYaml from "js-yaml"

const debug = DebugModule("verify-pnpm-lock")

const APPROVED_LEGACY_SPECIFIERS: { [dependency: string]: string[] } = {
  "@types/react": ["^18.3.27"], // For taro miniapp
  "react-dom": ["^18.3.1"], // For taro miniapp
  react: ["^18.3.1"], // For taro miniapp
  vite: ["^5.4.21"], // For taro miniapp
}

type PnpmLock = {
  importers: {
    [packagePath: string]: {
      dependencies?: {
        [dependency: string]: {
          specifier: string
          version: string
        }
      }
      devDependencies?: {
        [dependency: string]: {
          specifier: string
          version: string
        }
      }
    }
  }
}

export const VerifyPnpmLock = async (args: { lockPath: string }) => {
  const { lockPath } = args
  const lockYaml = await fs.readFile(lockPath, "utf8")

  const versions: {
    [dependency: string]: { [specifier: string]: string[] }
  } = {}
  const lock = JsYaml.load(lockYaml) as PnpmLock
  Object.keys(lock.importers).forEach((packagePath) => {
    const dependencies = {
      ...lock.importers[packagePath].dependencies,
      ...lock.importers[packagePath].devDependencies,
    }
    debug(`PackagePath: ${packagePath}`)
    Object.keys(dependencies).forEach((dependency) => {
      const specifier = dependencies[dependency].specifier
      versions[dependency] = versions[dependency] || {}
      versions[dependency][specifier] = versions[dependency][specifier] || []
      versions[dependency][specifier].push(packagePath)
      debug(`  ${dependency}: ${specifier}`)
    })
  })
  let ok = true
  Object.keys(versions).forEach((dependency) => {
    const specifiers = Object.keys(versions[dependency]).filter(
      (specifier) =>
        !APPROVED_LEGACY_SPECIFIERS[dependency]?.includes(specifier),
    )
    if (specifiers.length > 1) {
      console.error(
        `Dependency ${dependency} has multiple specifiers:`,
        versions[dependency],
      )
      ok = false
    }
  })
  if (!ok) {
    throw new Error("Multiple specifiers found")
  }
}

const main = async () => {
  const pnpmLock = process.argv[2] || "../../../../../pnpm-lock.yaml"
  await VerifyPnpmLock({ lockPath: pnpmLock })
}

if (import.meta.filename == process.argv[1]) {
  main()
}
