import fs from "node:fs/promises"
import process from "node:process"

import DebugModule from "debug"
import JsYaml from "js-yaml"

const debug = DebugModule("verify-pnpm-lock")

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
    const specifiers = Object.keys(versions[dependency])
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
