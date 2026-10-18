/**
 * Upstream sources this directory can sync from, defined in `UPSTREAM.ndscm.ts`.
 *
 * An `UpstreamConfig` is keyed by upstream name (for example `"theseed"`).
 * Each entry declares **where** a source lives and **how** its commits are
 * folded into the local branch:
 *
 * - `converge` — how the local and upstream commit chains are reconciled.
 * - `scm` — source control system backing the upstream.
 * - `repo` / `local` / `tracking` — which ref to track, and where it lives.
 * - `include` / `exclude` / `rename` — path modifiers applied in order to
 *   every incoming commit.
 *
 * This config exists to share upstream definitions across the team and
 * across branches in a single, reviewable file. It is not consumed by
 * local git operations: ndscm does not use it to drive `git fetch`,
 * `git pull`, `git rebase`, etc. Local git state (remotes, tracking
 * branches, refspecs) is managed separately by git itself.
 *
 * @example
 * ```ts
 * export default {
 *   theseed: {
 *     converge: "melt",
 *     scm: "git",
 *     repo: "https://github.com/ndscm/theseed.git",
 *     tracking: "main",
 *   },
 * } satisfies UpstreamConfig
 * ```
 */
export type UpstreamConfig = {
  [name: string]: {
    /**
     * How the local and upstream commit chains are reconciled when
     * syncing. Only `"melt"` is supported in the directory upstream config.
     *
     * - `"melt"` — Rebase the upstream patches onto the local tip,
     *   regenerating them; the local chain is left unchanged. Use this
     *   for long-lived integration branches that absorb upstream changes
     *   without rewriting their own history.
     */
    converge: "melt"

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
     * URL of the upstream repository. When omitted, the upstream lives in
     * the same repository as the local branch — i.e., this entry
     * describes a tracking relationship between two refs in the current
     * repo rather than a remote source.
     */
    repo?: string

    /**
     * Whether `tracking` names a branch in the local repository rather
     * than one on `repo`. Defaults to `false`, in which case `tracking` is
     * resolved on the upstream side.
     */
    local?: boolean

    /**
     * Name of the branch to track. Resolved on the upstream `repo` by
     * default, or in the local repository when `local` is `true`.
     */
    tracking?: string

    /**
     * Path filters controlling which files from incoming upstream commits
     * are kept. Each pattern is a regex matched against file paths.
     *
     * The modifiers are applied one by one, in order, to every incoming
     * commit: {@link include} first, then {@link exclude}, then
     * {@link rename}.
     *
     * Only paths matching at least one `include` pattern are carried over.
     * When omitted, all paths are included.
     */
    include?: string[]

    /**
     * Path filters removing files from incoming upstream commits. Each
     * pattern is a regex matched against file paths.
     *
     * Applied after {@link include} and before {@link rename}: any path
     * matching an `exclude` pattern is dropped, even if it was included.
     */
    exclude?: string[]

    /**
     * Path rewrites applied to incoming upstream commits. Each rule is a
     * regex {@link match}/{@link replace} pair.
     *
     * Applied last, after {@link include} and {@link exclude}, to the paths
     * that survive filtering. Rules are applied in order.
     */
    rename?: {
      /** Regex matched against the file path. */
      match: string

      /** Replacement for the matched portion. Supports regex backreferences. */
      replace: string
    }[]
  }
}
