if (process.env.BUILD_WORKING_DIRECTORY) {
  process.chdir(process.env.BUILD_WORKING_DIRECTORY)
}
require("prettier/bin/prettier.cjs")
