/**
 * Repository-wide ndscm configuration, defined in `REPO.ndscm.ts` at the
 * repository root. Each feature or development branch can override this
 * config with its own `REPO.ndscm.ts`, allowing branches to carry
 * branch-specific upstream definitions or tooling versions.
 */
export type RepoConfig = {
  /**
   * Email domain for the repository. Access control identifies users and
   * services by email address. When an email shares this domain, the domain
   * part can be omitted - for example, `"alice"` is equivalent to
   * `"alice@example.com"` if the repo domain is `"example.com"`.
   */
  domain?: string

  /**
   * Path to the primary directory branch developers should focus their work in.
   */
  home?: string

  /**
   * Advisory guidance on how freely to spend when weighing technical choices.
   */
  budget?: {
    /**
     * Overall spending level for the branch.
     *
     * ### `"low"`
     *
     * - Best for individual developers working with agents.
     * - Always deploy as serverless functions.
     * - By default, no long-running instances or databases.
     * - No idle costs.
     * - Deploy to prod only, with no staging environment.
     * - Write unit tests only for critical logic.
     * - Write hermetic integration tests only for common workflows.
     *
     * ### `"medium"`
     *
     * - Best for small tech teams.
     * - Always deploy as serverless functions.
     * - By default, allow a single long-running gateway instance.
     * - By default, allow a single SQL database server.
     * - Deploy to prod and staging environments.
     * - Write unit tests for common workflows.
     * - Write hermetic integration tests for every workflow.
     *
     * ### `"high"`
     *
     * - Best for medium tech teams.
     * - Prefer serverless, but allow long-running instances for performance.
     * - Allow different database engines where they bring performance.
     * - Include staging and dev environments.
     * - Isolate the development team from the deployment team.
     * - Meet Cybersecurity Framework standards.
     * - Write unit tests for every scenario.
     * - Write hermetic integration tests for every scenario.
     * - Write flaky staging tests.
     *
     * ### `"max"`
     *
     * - Best for large tech teams.
     * - Pick whatever performs best.
     * - Special process required for accessing customer data.
     * - Monitor regressions and downtime.
     * - Run Chaos Monkey in the production environment.
     *
     * ### `"ultra"`
     *
     * - No budget limit.
     * - Pick the best-performing options.
     * - Pick the best devprod options.
     */
    level?: "low" | "medium" | "high" | "max" | "ultra"

    /**
     * When enabled, relaxes the long-running service limits.
     *
     * - `"low"`: Permits a single gateway service.
     * - `"medium"`: Permits additional long-running services such as a
     *   message queue, mail, or drive.
     * - `"high"` and above: No effect.
     */
    relaxLongRunning?: boolean

    /**
     * When enabled, relaxes the long-running database limits.
     *
     * - `"low"`: Permits a single SQL database service.
     * - `"medium"`: Permits additional long-running database engines such
     *   as Redis or MongoDB.
     * - `"high"`: Permits multiple instances of the same database engine.
     * - `"max"` and above: No effect.
     */
    relaxDatabase?: boolean

    /**
     * When enabled, runs non-user-facing, non-production services such as CI
     * agents and AI agents on on-premises servers to cut costs.
     */
    onPremises?: boolean
  }

  /**
   * Collaboration surfaces for the repository: where issues are tracked and
   * where pending code changes are discussed. Each surface can be backed by
   * a different system, letting a repo mix ndscm-native tooling with GitHub.
   */
  collaboration?: {
    /**
     * The ticket tracker for issues, bugs, and feature requests.
     */
    tracker?: {
      /**
       * The tracker system.
       *
       * - `"ndscm"`: Stores tickets on a branch.
       * - `"github"`: Uses GitHub Issues.
       */
      system?: "ndscm" | "github"

      /**
       * Branch that stores the tickets. Only used by the `"ndscm"` system.
       * When empty, tickets are stored on the current branch.
       */
      branch?: string
    }

    /**
     * The discussion board for pending code changes.
     */
    review?: {
      /**
       * Backend hosting the review board.
       *
       * - `"ndscm"`: Stores patches and discussions on a branch.
       * - `"github"`: Uses GitHub pull requests.
       */
      system?: "ndscm" | "github"

      /**
       * Branch that stores the patches and discussions. Only used by the
       * `"ndscm"` system. When empty, they are stored on the current branch.
       */
      branch?: string
    }
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
       * the same repository as the local branch - i.e., this entry
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
       * - `"rebase"`: Rebase local patches onto the upstream tip.
       *   Local patches are regenerated; the upstream chain stays
       *   unchanged. Use this for feature and development branches.
       *
       * - `"melt"`: Rebase upstream patches onto the local tip.
       *   Upstream patches are regenerated; the local chain stays
       *   unchanged. Use this for long-lived integration branches
       *   that absorb upstream changes without rewriting their own
       *   history.
       */
      converge: "rebase" | "melt"
    }
  }
}
