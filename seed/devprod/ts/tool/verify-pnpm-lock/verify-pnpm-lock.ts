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
    const importer = lock.importers[packagePath]
    if (!importer) {
      return
    }
    const dependencies = {
      ...importer.dependencies,
      ...importer.devDependencies,
    }
    debug(`PackagePath: ${packagePath}`)
    Object.keys(dependencies).forEach((dependency) => {
      const dep = dependencies[dependency]
      if (!dep) {
        return
      }
      const specifier = dep.specifier
      const versionMap = versions[dependency] ?? {}
      versions[dependency] = versionMap
      const packages = versionMap[specifier] ?? []
      versionMap[specifier] = packages
      packages.push(packagePath)
      debug(`  ${dependency}: ${specifier}`)
    })
  })
  let ok = true
  Object.keys(versions).forEach((dependency) => {
    const versionEntry = versions[dependency]
    if (!versionEntry) {
      return
    }
    const specifiers = Object.keys(versionEntry).filter(
      (specifier) =>
        !APPROVED_LEGACY_SPECIFIERS[dependency]?.includes(specifier),
    )
    if (specifiers.length > 1) {
      console.error(
        `Dependency ${dependency} has multiple specifiers:`,
        versionEntry,
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
