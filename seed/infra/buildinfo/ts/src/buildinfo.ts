import BazelStableStatus from "./stable-status"
import BazelVolatileStatus from "./volatile-status"

type BuildInfo = {
  // bazel status
  buildHost: string
  buildUser: string

  // git status
  buildBranch: string
  buildBrief: string
  buildDirty: string
  buildGitCommit: string
  buildTag: string

  // volatile status
  buildTime: string
}

const parseBazelStatus = (statusText: string): Record<string, string> => {
  const status: Record<string, string> = {}
  const lines = statusText.split("\n")
  for (const rawLine of lines) {
    const line = rawLine.trim()
    if (line === "" || line.startsWith("#")) {
      continue
    }
    const separatorIndex = line.indexOf(" ")
    if (separatorIndex === -1) {
      status[line] = ""
      continue
    }
    const key = line.slice(0, separatorIndex)
    const value = line.slice(separatorIndex + 1)
    status[key] = value
  }
  return status
}

const parseBuildTime = (buildTimestamp: string | undefined): string => {
  if (buildTimestamp === undefined) {
    return ""
  }
  const buildTimeSeconds = Number.parseInt(buildTimestamp, 10)
  if (Number.isNaN(buildTimeSeconds)) {
    return ""
  }
  return new Date(buildTimeSeconds * 1000).toISOString()
}

export const Get = (): BuildInfo | null => {
  if (!BazelStableStatus && !BazelVolatileStatus) {
    return null
  }

  const stableStatus = parseBazelStatus(BazelStableStatus)
  const volatileStatus = parseBazelStatus(BazelVolatileStatus)

  const buildInfo: BuildInfo = {
    buildHost: stableStatus["BUILD_HOST"] ?? "",
    buildUser: stableStatus["BUILD_USER"] ?? "",

    buildBranch: stableStatus["STABLE_BUILD_BRANCH"] ?? "",
    buildBrief: stableStatus["STABLE_BUILD_BRIEF"] ?? "",
    buildDirty: stableStatus["STABLE_BUILD_DIRTY"] ?? "",
    buildGitCommit: stableStatus["STABLE_GIT_COMMIT"] ?? "",
    buildTag: stableStatus["STABLE_BUILD_TAG"] ?? "",

    buildTime: parseBuildTime(volatileStatus["BUILD_TIMESTAMP"]),
  }
  return buildInfo
}

export default {
  Get,
}
