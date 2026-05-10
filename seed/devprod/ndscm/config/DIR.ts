/**
 * Per-directory ndscm configuration, defined in `DIR.ndscm.ts`.
 *
 * A `DirConfig` declares **what to run** when files change inside a managed
 * directory. Each top-level key is a lifecycle phase; each phase contains
 * named tasks with a `watch` pattern, an optional `target`, and a `run`
 * command:
 *
 * - `watch` — regex matched against changed file paths to decide whether
 *   the task should fire.
 * - `target` — output file(s) the task produces. Used to skip the task
 *   when outputs are already up-to-date. Absent for format and test tasks.
 * - `run` — shell command(s) to execute. Format tasks receive the changed
 *   file path via `{{TARGET}}`.
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
 * | **vendor**      | gitignored   | Fetch deps for bazel hermetic builds.                              |
 * | **bootstrap**   | gitignored   | Set up local dev env (editors, language servers, non-bazel tools). |
 * | **tidy**        | git-tracked  | Regenerate checked-in files (`go.mod`, `BUILD.bazel`).             |
 * | **lock**        | git-tracked  | Regenerate checked-in lock files (`go.sum`, `MODULE.bazel.lock`).  |
 * | **build:bazel** | *(none)*     | Compile bazel artifacts. Depends on vendor only.                   |
 * | **build:other** | gitignored   | Compile other artifacts. Can also depend on bootstrap.             |
 * | **test:bazel**  | *(none)*     | Run bazel tests. Depends on vendor only.                           |
 * | **test:other**  | *(none)*     | Run other tests. Also depends on bootstrap.                        |
 * | **test:tidy**   | *(none)*     | Run tidy and lock tests.                                           |
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

      /**
       * Bazel target(s) that will be built before the formatter runs.
       *
       * All targets across the phase are batched into a single
       * `bazel build` invocation. After the build, each target's
       * output paths are resolved and exposed as template variables
       * in `run`:
       *
       * - `{{BAZEL_RUN}}` / `{{BAZEL_RUN[i]}}` — the default
       *   executable output of the first / i-th target.
       * - `{{BAZEL_BUILD}}` / `{{BAZEL_BUILD[i]}}` — the default
       *   artifact (file) output of the first / i-th target.
       */
      needBazelBuild?: string | string[]

      /**
       * Shell command executed via `bash -c`.
       *
       * Template variables:
       * - `{{TARGET}}` — absolute path of the file being formatted.
       * - `{{BAZEL_RUN}}` / `{{BAZEL_RUN[i]}}` — executable from
       *   `needBazelBuild` (see above).
       * - `{{BAZEL_BUILD}}` / `{{BAZEL_BUILD[i]}}` — artifact from
       *   `needBazelBuild` (see above).
       */
      run: string | string[]
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
      run: string | string[]
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
       * changes. Make checks file mtime, so point at concrete files, not
       * directories.
       */
      watchRepo?: string | string[]

      /** Regex matched against changed file paths. */
      watch: string | string[]

      /** Shell command to run. */
      run: string | string[]
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
       * changes. Make checks file mtime, so point at concrete files, not
       * directories.
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
       * changes. Make checks file mtime, so point at concrete files, not
       * directories.
       */
      watchRepo?: string | string[]

      /** Regex matched against changed file paths. */
      watch: string | string[]

      /** Shell command(s) to run. An array is executed sequentially. */
      run: string | string[]
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
       * changes. Make checks file mtime, so point at concrete files, not
       * directories.
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
}
