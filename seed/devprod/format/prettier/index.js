const path = require("node:path")

// Prettier resolves plugins via a bundled import-meta-resolve polyfill that
// walks node_modules from process.cwd(). By NOT chdir-ing, cwd stays in the
// Bazel runfiles where node_modules lives, so plugin resolution works.
// We instead point --config and --ignore-path to the workspace root.
if (process.env.BUILD_WORKING_DIRECTORY) {
  process.argv.push(
    "--config",
    path.resolve(process.env.BUILD_WORKING_DIRECTORY, ".prettierrc"),
    "--ignore-path",
    path.resolve(process.env.BUILD_WORKING_DIRECTORY, ".prettierignore"),
  )
}

require("prettier/bin/prettier.cjs")
