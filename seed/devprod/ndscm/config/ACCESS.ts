/**
 * Access control settings for a directory.
 *
 * These are hints for tooling — actual enforcement happens in the code
 * management system and the build system.
 *
 * Omitted fields inherit from the parent directory. When a field is provided,
 * it fully overrides the parent value rather than extending it.
 */
export type AccessConfig = {
  /** The person responsible for code in this directory. */
  owner?: string

  /** People who can approve changes in this directory. */
  maintainer?: string[]

  /**
   * People automatically assigned to review any change in this directory,
   * whether submitted by a team member or an outside contributor. Reviewers
   * don't necessarily have approval power.
   */
  reviewer?: string[]

  /**
   * People automatically assigned to review a change when it is submitted by
   * a reviewer. This supports a workflow where reviewers do a first-round
   * triage of external requests, and coworkers handle the final review and
   * approval. Typically a subset of the owner and maintainers.
   */
  coworker?: string[]

  /** People with read access to code in this directory. */
  reader?: string[]

  /** People allowed to build code in this directory. */
  builder?: string[]

  /**
   * Per-file access control overrides. Keys are file names relative to this
   * directory. Directory names are ignored here — define their access control
   * in the directory's own config file.
   */
  files?: {
    /** Same fields as the directory level, scoped to this file. */
    [fileName: string]: {
      owner?: string
      maintainer?: string[]
      reviewer?: string[]
      coworker?: string[]
      reader?: string[]
      builder?: string[]
    }
  }
}
