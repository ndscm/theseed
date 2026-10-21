/**
 * A parsed xslash entry: a JSON file describing where to download a prebuilt
 * executable for each platform. The file may begin with a `#!` shebang line,
 * which is stripped before the JSON is parsed.
 */
export type XslashEntry = {
  /** Logical name; used as arg0 on exec and as the lookup key in `xslash.json`. */
  name: string

  /** If true, use this entry directly and skip project-level resolution. */
  final?: boolean

  /**
   * Platform blocks keyed by `<kernel>-<arch>` — kernel one of `linux`,
   * `darwin`, `windows`; arch one of `amd64`, `arm64` (e.g. `linux-amd64`). May
   * be empty for a bare name-only entry that defers to project-level resolution.
   *
   * Each block is either a single artifact (via `hash`/`digest`/`size`/
   * `providers`) or several composed into one directory (via `compose`); the
   * two are mutually exclusive, and `path` is always required. These rules are
   * enforced when the entry is resolved rather than by this type.
   */
  platforms?: {
    [platform: string]: {
      /**
       * Hash algorithm used for `digest`; `blake3` is the only supported value.
       * Required for a single-artifact platform; absent when `compose` is set.
       */
      hash: string

      /**
       * Expected digest of the downloaded artifact, in lowercase hex. Required
       * for a single-artifact platform; absent when `compose` is set.
       */
      digest: string

      /**
       * Expected size of the downloaded artifact, in bytes. Required for a
       * single-artifact platform; absent when `compose` is set.
       */
      size: number

      /**
       * Archive/compression format, or omitted for a single raw executable.
       * One of `tar.gz`, `tar.xz`, `tar.zst`, `tar`, `gz`, `xz`, `zst`, `zip`.
       */
      format?: string

      /** If true, extract under `rw/` (mutable) instead of `ro/` (shared). */
      rw?: boolean

      /**
       * A dotslash-compatibility fallback for `rw`, which it should not co-exist
       * with.
       */
      readonly?: boolean

      /**
       * Where the executable is found within the extraction directory; it is
       * run from `<extraction dir>/<path>`. Required.
       *
       * - Archive formats: the path to the executable inside the extracted
       *   archive.
       * - Single-file formats (`gz`/`xz`/`zst`/none): the path the decompressed
       *   file — or, with no format, the raw binary — is placed at; that file
       *   is itself the executable.
       * - With `compose` set: the path to the executable within the composed
       *   directory.
       */
      path: string

      /**
       * Sources to fetch the artifact from, tried in order. xslash only fetches
       * `http`/`https` URLs; a provider with no `url` is skipped. Required for a
       * single-artifact platform; absent when `compose` is set.
       */
      providers?: {
        /** Provider type, kept for dotslash compatibility. */
        type?: string

        /** The download URL. Must start with `http://` or `https://`. */
        url: string
      }[]

      /**
       * Several artifacts composed into one directory, as an alternative to a
       * single artifact. Mutually exclusive with the sibling artifact fields
       * above.
       */
      compose?: {
        /** Hash algorithm used for `digest`; `blake3` is the only supported value. */
        hash: string

        /** Expected digest of the downloaded artifact, in lowercase hex. */
        digest: string

        /** Expected size of the downloaded artifact, in bytes. */
        size: number

        /**
         * Archive/compression format, or omitted for a single raw file. One of
         * `tar.gz`, `tar.xz`, `tar.zst`, `tar`, `gz`, `xz`, `zst`, `zip`.
         */
        format?: string

        /**
         * For an archive, strip this leading in-archive subpath: extract only
         * the entries under it, rebasing them onto the destination (the composed
         * root, or `path` when also set) with the prefix dropped; entries
         * outside it are skipped. Omitted extracts the whole archive. Ignored
         * for single-file formats. Lets archives that each wrap their payload in
         * a differing leading directory merge into one tree (e.g. the Rust SDK's
         * component tarballs onto a shared sysroot).
         */
        strip?: string

        /**
         * Where this artifact goes inside the composed directory. For an archive
         * format, the subpath it's extracted into (optional; defaults to the
         * composed root). For a single-file format (`gz`/`xz`/`zst`/none), the
         * path the single file is placed at; required.
         */
        path?: string

        /** Sources to fetch this artifact from, tried in order. */
        providers: {
          /** Provider type, kept for dotslash compatibility. */
          type?: string

          /** The download URL. Must start with `http://` or `https://`. */
          url: string
        }[]
      }[]
    }
  }
}
