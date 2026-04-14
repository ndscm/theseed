import { type DirConfig } from "./seed/devprod/ndscm/config/DIR"

export default {
  format: {
    ts: {
      watch: [
        "\\.(ts|tsx|js|jsx|mts|cts|mjs|cjs|css|scss|less|json|yaml|yml|md|mdx|html|vue)$",
      ],
      run: 'bazel run //seed/devprod/format/prettier -- --write "{{TARGET}}"',
    },
    cc: {
      watch: "\\.(c|cc|cpp|h|hh|hpp)$",
      run: 'clang-format -i "{{TARGET}}"',
    },
    go: {
      watch: "\\.(go)$",
      run: 'bazel run //seed/devprod/format/gofmt -- -w "{{TARGET}}"',
    },
    groovy: {
      watch: "(^|/)Jenkinsfile$",
      run: 'bazel run //seed/devprod/format/groovy -- --write "{{TARGET}}"',
    },
    java: {
      watch: "\\.(java)$",
      run: 'clang-format -i "{{TARGET}}"',
    },
  },
  bootstrap: {
    pnpm: {
      target: "node_modules",
      watch: "^pnpm-lock.yaml$",
      run: 'bazel run @pnpm//:pnpm -- --dir "$(pwd)" install',
    },
    uv: {
      target: ".venv",
      watch: "^uv.lock$",
      run: "uv sync",
    },
  },
  tidy: {
    go: {
      target: "go.mod",
      watch: "\\.go$",
      run: "bazel run @rules_go//go -- mod tidy",
    },
    gazelle_python_modules_mapping: {
      target: [
        "gazelle_python_modules_mapping_darwin.json",
        "gazelle_python_modules_mapping_linux.json",
      ],
      watch: "^pyproject.toml$",
      run: "bazel run //seed/devprod/python/modules_mapping:generate",
    },
    gazelle_python: {
      target: "gazelle_python.yaml",
      watch: "^pyproject.toml$",
      run: "bazel run //:gazelle_python_manifest.update",
    },
    gazelle_build: {
      target: "BUILD.bazel",
      watch: "(^|/)BUILD.bazel$",
      run: "bazel run //:gazelle",
    },
    gazelle_module: {
      target: "MODULE.bazel",
      watch: [
        //
        "\\.(go|py)$",
        "(^|/)BUILD.bazel$",
      ],
      run: "bazel run //:gazelle",
    },
  },
  lock: {
    bazel: {
      target: "MODULE.bazel.lock",
      watch: "(^|/)BUILD.bazel$",
      run: "bazel mod tidy",
    },
    go: {
      target: "go.sum",
      watch: "^go.mod$",
      run: "bazel run @rules_go//go -- mod tidy",
    },
    pnpm: {
      target: "pnpm-lock.yaml",
      watch: [
        //
        "(^|/)package.json$",
        "^pnpm-workspace.yaml$",
      ],
      run: 'bazel run @pnpm//:pnpm -- --dir "$(pwd)" install',
    },
    requirements: {
      target: ["requirements.txt", "requirements_darwin.txt"],
      watch: "^pyproject.toml$",
      run: 'uv sync; if [[ "$(uname)" == "Darwin" ]]; then uv pip freeze --color never >./requirements_darwin.txt; else uv pip freeze --color never >./requirements.txt; fi',
    },
    uv: {
      target: "uv.lock",
      watch: "^pyproject.toml$",
      run: "uv sync",
    },
  },
  test: {
    bazel_build: {
      watch: "^.*$",
      run: "bazel build //...",
    },
    bazel_test: {
      watch: "^.*$",
      run: "bazel test //...",
    },
    verify_pnpm_lock: {
      watch: [
        //
        "(^|/)package.json$",
        "^pnpm-workspace.yaml$",
      ],
      run: 'cd ./seed/devprod/ts/tool/verify-pnpm-lock && bazel run @pnpm//:pnpm -- --dir "$(pwd)" run verify-pnpm-lock',
    },
  },
} satisfies DirConfig
