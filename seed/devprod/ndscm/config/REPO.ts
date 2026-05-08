/**
 * Repository-wide ndscm configuration, defined in `REPO.ndscm.ts` at the
 * repository root. Each feature or development branch can override this
 * config with its own `REPO.ndscm.ts`, allowing branches to carry
 * branch-specific upstream definitions or tooling versions.
 */
export type RepoConfig = {
  /**
   * ndscm tooling metadata. Controls which version of ndscm the repository
   * expects, so that the CLI can warn or adapt when running against a
   * mismatched repo.
   */
  ndscm: {
    /**
     * Version of the ndscm tooling this repository targets.
     */
    version?: string
  }

  /**
   * Upstream sources the repository can sync from. Each entry is keyed by a
   * remote name (for example `"theseed"`) and describes how to fetch and
   * track that source.
   *
   * This config exists to share upstream definitions across the team and
   * across branches in a single, reviewable file. It is not consumed by
   * local git operations: ndscm does not use it to drive `git fetch`,
   * `git pull`, `git rebase`, etc. Local git state (remotes, tracking
   * branches, refspecs) is managed separately by git itself.
   */
  upstream?: {
    [name: string]: {
      /**
       * Source control system used by the upstream. Only `"git"` is
       * supported today.
       *
       * This field is required rather than defaulted so that adding
       * support for additional SCMs in the future does not silently
       * change the meaning of existing configs. Forcing every entry to
       * spell out `scm: "git"` keeps current users from being migrated
       * out from under them when new options appear.
       */
      scm: "git"

      /**
       * Repository URL to fetch from. When omitted, the upstream lives in
       * the same repository as the local branch — i.e., this entry
       * describes a tracking relationship between two refs in the current
       * repo rather than a remote source.
       */
      repo?: string

      /**
       * Whether `tracking` names a local branch instead of a branch on
       * `repo`. Defaults to `false`, in which case `tracking` is resolved
       * against the remote.
       */
      local?: boolean

      /**
       * Name of the branch to track on the upstream side.
       */
      tracking?: string

      /**
       * How the local and upstream chains converge when syncing.
       *
       * - `"rebase"` — Rebase local patches onto the upstream tip.
       *   Local patches are regenerated; the upstream chain stays
       *   unchanged. Use this for feature and development branches.
       *
       * - `"melt"` — Rebase upstream patches onto the local tip.
       *   Upstream patches are regenerated; the local chain stays
       *   unchanged. Use this for long-lived integration branches
       *   that absorb upstream changes without rewriting their own
       *   history.
       */
      converge: "rebase" | "melt"
    }
  }
}
