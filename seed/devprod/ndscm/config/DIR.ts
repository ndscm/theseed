/**
 * Per-directory ndscm configuration, defined in `DIR.ndscm.ts`.
 *
 * A `DirConfig` declares **what to run** when files change inside a managed
 * directory. Each top-level key is a lifecycle phase; each phase contains
 * named tasks with a `watch` pattern, an optional `target`, and one or more
 * execution methods (`run`, `bazel`, `buck`):
 *
 * - `watch` — regex matched against changed file paths to decide whether
 *   the task should fire.
 * - `target` — output file(s) the task produces. Used to skip the task
 *   when outputs are already up-to-date. Absent for format and test tasks.
 * - `run` — shell command(s) to execute.
 * - `bazel` — bazel targets to build and run. Only triggered when
 *   `build_system` includes `bazel`. `build` target(s) are batched into
 *   a single `bazel build` invocation per phase; `run` command(s) are
 *   executed via `bazel run`.
 * - `buck` — buck targets to run. Only triggered when `build_system`
 *   includes `buck`. Optimized build ground is not supported for buck.
 *
 * ### Lifecycle phases
 *
 * ```
 * checkout ───┬──► format (individual)
 *             │
 *             └──► vendor ───┬──► build:bazel ──► test:bazel
 *                            │
 *                            └──► bootstrap ───┬──► build:other ──► test:other
 *                                              │
 *                                              └──► tidy + lock ──► test:tidy
 * ```
 *
 * | Phase           | Outputs      | Purpose                                                            |
 * | --------------- | ------------ | ------------------------------------------------------------------ |
 * | **format**      | *(in-place)* | Auto-format individual files. Standalone — no heavy setup.         |
 * | **vendor**      | gitignored   | Fetch deps for hermetic builds (bazel, buck).                      |
 * | **bootstrap**   | gitignored   | Set up local dev env (editors, language servers, non-bazel tools). |
 * | **tidy**        | git-tracked  | Regenerate checked-in files (`go.mod`, `BUILD.bazel`).             |
 * | **lock**        | git-tracked  | Regenerate checked-in lock files (`go.sum`, `MODULE.bazel.lock`).  |
 * | **build:bazel** | *(none)*     | Compile bazel artifacts. Depends on vendor only.                   |
 * | **build:other** | gitignored   | Compile other artifacts. Can also depend on bootstrap.             |
 * | **test:bazel**  | *(none)*     | Run bazel tests. Depends on vendor only.                           |
 * | **test:other**  | *(none)*     | Run other tests. Also depends on bootstrap.                        |
 * | **test:tidy**   | *(none)*     | Run tidy and lock tests.                                           |
 *
 * ### Environment variables
 *
 * The following environment variables are set for every `run` command:
 *
 * - `BUILD_WORKSPACE_DIRECTORY` — absolute path of the repo root.
 * - `BUILD_WORKING_DIRECTORY` — absolute path of the task directory.
 *
 * ### Template variables
 *
 * Template variables are substituted into `run` commands where applicable:
 *
 * - `{{TARGET}}` — absolute path of the file being processed. Only
 *   available in format tasks.
 *
 * The following are **bazel-only** template variables, available when
 * `build_system` includes `bazel`:
 *
 * - `{{BAZEL_RUN}}` / `{{BAZEL_RUN[i]}}` — the executable output of the
 *   first / i-th `bazel.build` target. When bazel ground is disabled,
 *   expands to `bazel run <target> -- ` instead.
 * - `{{BAZEL_EXECUTABLE}}` / `{{BAZEL_EXECUTABLE[i]}}` — the executable
 *   path of the first / i-th `bazel.build` target. Empty when bazel
 *   ground is disabled.
 * - `{{BAZEL_BUILD}}` / `{{BAZEL_BUILD[i]}}` — the artifact output paths
 *   (space-joined) of the first / i-th `bazel.build` target. Only
 *   available when bazel ground is enabled.
 *
 * @example
 * ```ts
 * export default {
 *   format: {
 *     ts: {
 *       watch: "\\.(ts|tsx|js|jsx)$",
 *       run: 'prettier --write "{{TARGET}}"',
 *     },
 *   },
 *   tidy: {
 *     go: {
 *       target: "go.mod",
 *       watch: "\\.go$",
 *       run: "go mod tidy",
 *     },
 *   },
 *   lock: {
 *     go: {
 *       target: "go.sum",
 *       watch: "^go.mod$",
 *       run: "go mod tidy",
 *     },
 *   },
 * } satisfies DirConfig
 * ```
 */
export type DirConfig = {
  /**
   * Commit granularity for changes in this directory.
   *
   * When set, changes to this directory are grouped into a single separate
   * commit rather than folded into a shared commit with other directories.
   *
   * @example
   * ```ts
   * granularity: {
   *   convention: "linux",
   *   prefix: "seed: ndscm: ",
   * }
   * ```
   */
  granularity?: {
    /**
     * Commit message style the commit in this directory should follow.
     *
     * - `"linux"` — Linux-kernel style, `<project>: <sub>: verb summary`.
     * - `"conventional"` — Conventional Commits, `type(scope/sub): verb summary`.
     * - `"freedom"` — No enforced convention.
     */
    style: "linux" | "conventional" | "freedom"

    /**
     * Prefix prepended to the commit message for this directory's commit.
     *
     * Only used for the `"linux"` {@link style}, and usually end with `: `
     */
    prefix?: string

    /**
     * Scope inserted into the commit message for this directory's commit.
     *
     * Only used for the `"conventional"` {@link style}.
     */
    scope?: string
  }

  /**
   * Auto-format individual changed files in place.
   *
   * Standalone — must not depend on bootstrap outputs or heavy setup
   * (`node_modules`, `.venv`). Bazel-provided tools are fine.
   *
   * `run` receives the changed file path as `{{TARGET}}`.
   *
   * @example
   * ```ts
   * format: {
   *   ts: {
   *     watch: "\\.(ts|tsx|js|jsx|css|json|yaml|md|html)$",
   *     run: 'prettier --write "{{TARGET}}"',
   *   },
   *   cc: {
   *     watch: "\\.(c|cc|cpp|h|hh|hpp)$",
   *     run: 'clang-format -i "{{TARGET}}"',
   *   },
   * }
   * ```
   */
  format?: {
    [task: string]: {
      /** Regex matched against changed file paths. */
      watch: string | string[]

      /** Shell command executed via `bash -c`. */
      run?: string | string[]

      /** Bazel targets to build and run for this task. */
      bazel?: {
        /** Bazel target(s) to build before the task runs. */
        build: string | string[]
        /** Bazel command(s) to execute via `bazel run`. */
        run: string | string[]
      }

      /** Buck targets to run for this task. */
      buck?: {
        /** Buck command(s) to execute. */
        run: string | string[]
      }
    }
  }

  /**
   * Fetch vendored dependencies for bazel hermetic builds.
   *
   * Outputs are gitignored and regenerated on demand.
   */
  vendor?: {
    [task: string]: {
      /** Gitignored output file(s) or directory. Skipped when up-to-date. */
      target: string | string[]

      /**
       * Literal file paths (relative to repo root) added as extra Make
       * prerequisites. Use this to express cross-project dependencies —
       * e.g., this task must re-run when another project's build output
       * changes.
       */
      watchRepo?: string | string[]

      /** Regex matched against changed file paths. */
      watch: string | string[]

      /** Shell command to run. */
      run?: string | string[]

      /** Bazel targets to build and run for this task. */
      bazel?: {
        /** Bazel target(s) to build before the task runs. */
        build: string | string[]
        /** Bazel command(s) to execute via `bazel run`. */
        run: string | string[]
      }

      /** Buck targets to run for this task. */
      buck?: {
        /** Buck command(s) to execute. */
        run: string | string[]
      }
    }
  }

  /**
   * Set up the local dev environment so editors, language servers, and
   * non-bazel toolchains (`pnpm build`, `go build`) work correctly.
   *
   * Outputs are gitignored. May depend on bazel targets (e.g.
   * code-generated files). Typically triggered by lockfile changes.
   *
   * @example
   * ```ts
   * bootstrap: {
   *   pnpm: {
   *     target: "node_modules",
   *     watch: "^pnpm-lock.yaml$",
   *     run: "pnpm install",
   *   },
   *   uv: {
   *     target: ".venv",
   *     watch: "^uv.lock$",
   *     run: "uv sync",
   *   },
   * }
   * ```
   */
  bootstrap?: {
    [task: string]: {
      /** Gitignored output file(s) or directory. Skipped when up-to-date. */
      target: string | string[]

      /**
       * Literal file paths (relative to repo root) added as extra Make
       * prerequisites. Use this to express cross-project dependencies —
       * e.g., this task must re-run when another project's build output
       * changes.
       */
      watchRepo?: string | string[]

      /** Regex matched against changed file paths. */
      watch: string | string[]

      /** Shell command to run. */
      run?: string | string[]

      /** Bazel targets to build and run for this task. */
      bazel?: {
        /** Bazel target(s) to build before the task runs. */
        build: string | string[]
        /** Bazel command(s) to execute via `bazel run`. */
        run: string | string[]
      }

      /** Buck targets to run for this task. */
      buck?: {
        /** Buck command(s) to execute. */
        run: string | string[]
      }
    }
  }

  /**
   * Regenerate **git-tracked** files (`go.mod`, `BUILD.bazel`) to keep
   * them in sync with their sources.
   *
   * Usually depends on bootstrap being up-to-date. Commands can be chained
   * as an array when multiple steps are needed.
   *
   * @example
   * ```ts
   * tidy: {
   *   go: {
   *     target: "go.mod",
   *     watch: "\\.go$",
   *     run: "go mod tidy",
   *   },
   *   gazelle: {
   *     target: "BUILD.bazel",
   *     watch: "\\.(go|py)$",
   *     run: "bazel run //:gazelle",
   *   },
   * }
   * ```
   */
  tidy?: {
    [task: string]: {
      /**
       * Git-tracked output file(s) regenerated by this command.
       * Other modified files are treated as side effects.
       */
      target: string | string[]

      /**
       * Literal file paths (relative to repo root) added as extra Make
       * prerequisites. Use this to express cross-project dependencies —
       * e.g., this task must re-run when another project's build output
       * changes.
       */
      watchRepo?: string | string[]

      /** Regex matched against changed file paths. */
      watch: string | string[]

      /** Shell command(s) to run. An array is executed sequentially. */
      run: string | string[]
    }
  }

  /**
   * Regenerate **git-tracked** lock files to keep them in sync with
   * their sources.
   *
   * Runs after tidy, since lock files are derived from the config files
   * that tidy produces.
   *
   * @example
   * ```ts
   * lock: {
   *   go: {
   *     target: "go.sum",
   *     watch: "^go.mod$",
   *     run: "go mod tidy",
   *   },
   *   bazel: {
   *     target: "MODULE.bazel.lock",
   *     watch: "(^|/)BUILD.bazel$",
   *     run: "bazel mod tidy",
   *   },
   * }
   * ```
   */
  lock?: {
    [task: string]: {
      /**
       * Git-tracked output file(s) regenerated by this command.
       * Other modified files are treated as side effects.
       */
      target: string | string[]

      /**
       * Literal file paths (relative to repo root) added as extra Make
       * prerequisites. Use this to express cross-project dependencies —
       * e.g., this task must re-run when another project's build output
       * changes.
       */
      watchRepo?: string | string[]

      /** Regex matched against changed file paths. */
      watch: string | string[]

      /** Shell command(s) to run. An array is executed sequentially. */
      run: string | string[]

      /**
       * When `true`, this task's regenerated lock file is committed in a
       * separate commit rather than combined with the directory's other
       * changes.
       */
      granularity?: boolean
    }
  }

  /**
   * Compile or bundle artifacts from source.
   *
   * Two paths: **bazel** builds depend on vendor; **other** builds
   * (non-bazel toolchains) depend on bootstrap.
   */
  build?: {
    [task: string]: {
      /** Output file(s) or directory produced by the build. */
      target: string | string[]

      /**
       * Literal file paths (relative to repo root) added as extra Make
       * prerequisites. Use this to express cross-project dependencies —
       * e.g., this task must re-run when another project's build output
       * changes.
       */
      watchRepo?: string | string[]

      /** Regex matched against changed file paths. */
      watch: string | string[]

      /** Shell command(s) to run. */
      run: string | string[]
    }
  }

  /**
   * Run tests when source files change. No `target` — tests produce no
   * artifacts.
   *
   * Three paths: **bazel** tests follow bazel builds, **other** tests
   * follow other builds, and **tidy** tests follow tidy and lock.
   *
   * @example
   * ```ts
   * test: {
   *   bazel: {
   *     watch: "^.*$",
   *     run: "bazel test //...",
   *   },
   * }
   * ```
   */
  test?: {
    [task: string]: {
      /** Regex matched against changed file paths. */
      watch: string | string[]

      /** Shell command(s) to run. */
      run: string | string[]
    }
  }

  /**
   * Upstream sources this directory can sync from. Each entry is keyed by an
   * upstream name (for example `"theseed"`) and declares where that source
   * lives and which branch to track.
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
    }
  }
}
