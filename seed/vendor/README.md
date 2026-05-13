# Seed Vendor

This directory contains third-party packages that theseed officially supports at
tier 1 and tier 2.

## Naming Conventions

Each project lives in its own directory and must include a
`project.MODULE.bazel` file, where `project` matches the directory name.

Directory names at this level must be lowercase. Dashes are allowed but not
recommended; underscores and other symbols are not allowed at this level (though
they are allowed one level down).

## Directory Structure

How a project is organized depends on what kind of project it is.

### Well-Known Projects

Well-known projects get their own directory under their well-known name.

### Languages

Languages get their own directory. If the language name is a single word, use
the full name (e.g. `python`, `java`, `shell`). If the language name is multiple
words, use the file extension (e.g. `ts`, `cc`).

### GitHub-hosted Projects

For GitHub-hosted projects, the module file should live at
`org/repo/repo.MODULE.bazel`. If the repo is the primary repo of its org (e.g.
`keycloak/keycloak`, `abseil/abseil-cpp`), the structure can be simplified to
`org/org.MODULE.bazel`.

## Adding a New Project

For projects that don't fall into the categories above, contact the theseed
maintainers for approval before adding a new directory.
