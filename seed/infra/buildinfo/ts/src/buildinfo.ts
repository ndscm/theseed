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

export const Get = (): BuildInfo | null => {
  if (typeof window === "undefined") {
    // TODO(nagi): support node runtime
    return null
  }
  const buildInfo = (window as any).__SEED_BUILD_INFO__
  if (!buildInfo) {
    return null
  }
  return buildInfo as BuildInfo
}

export default {
  Get,
}
