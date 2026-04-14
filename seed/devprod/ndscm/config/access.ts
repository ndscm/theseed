/**
 * Access control configuration for a directory.
 */
export type AccessConfig = {
  /**
   * Person who leads the code development in the directory.
   */
  owner?: string[]

  /**
   * People who can approve the code in the directory.
   */
  maintainer?: string[]

  /**
   * People who are automatically assigned to review code in the directory.
   */
  reviewer?: string[]

  /**
   * People who are automatically assigned to review code by the reviewer.
   */
  coworker?: string[]

  /**
   * People who can read the code in the directory.
   */
  reader?: string[]

  /**
   * People who can build the code in the directory.
   */
  builder?: string[]

  /**
   * File-specific access control. The key is the file name relative to the
   * directory.
   */
  files?: {
    /**
     * The file name relative to the directory.
     */
    [fileName: string]: {
      /**
       * Person who leads the code development in the directory.
       */
      owner?: string[]

      /**
       * People who can approve the code in the file.
       */
      maintainer?: string[]

      /**
       * People who are automatically assigned to review code in the file.
       */
      reviewer?: string[]

      /**
       * People who are automatically assigned to review code by the reviewer
       * in the file.
       */
      coworker?: string[]

      /**
       * People who can read the code in the file.
       */
      reader?: string[]

      /**
       * People who can build the code in the file.
       */
      builder?: string[]
    }
  }
}
