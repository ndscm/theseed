import { type DirConfig } from "./seed/devprod/ndscm/config/dir"

export default {
  format: {
    ts: {
      watch:
        "\.(ts|tsx|js|jsx|mts|cts|mjs|cjs|css|scss|less|json|yaml|yml|md|mdx|html|vue)$",
      run: 'prettier --write "${target}"',
    },
    cc: {
      watch: "\.(c|cc|cpp|h|hh|hpp)$",
      run: 'clang-format -i "${target}"',
    },
    java: {
      watch: "\.(java)$",
      run: 'clang-format -i "${target}"',
    },
  },
  bootstrap: {
    pnpm: {
      target: "node_modules",
      watch: "^pnpm-lock.yaml$",
      run: "pnpm install",
    },
    uv: {
      target: ".venv",
      watch: "^uv.lock$",
      run: "uv sync",
    },
  },
  tidy: {
    go: {
      target: ["go.mod", "go.sum"],
      watch: "\.go$",
      run: "go mod tidy",
    },
    pnpm: {
      target: "pnpm-lock.yaml",
      watch: [
        //
        "(^|/)package.json$",
        "^pnpm-workspace.yaml$",
      ],
      run: "pnpm install",
    },
    uv: {
      target: [
        //
        "uv.lock",
        "requirements.txt",
        "requirements_darwin.txt",
      ],
      watch: "^pyproject.toml$",
      run: `if [[ "$(uname)" == "Darwin" ]] ; then ; uv pip freeze --color never >./requirements_darwin.txt ; else ; uv pip freeze --color never >./requirements.txt ; fi ;`,
    },
    modules_mapping: {
      target: [
        //
        "gazelle_python.yaml",
        "gazelle_python_modules_mapping_darwin.json",
        "gazelle_python_modules_mapping_linux.json",
      ],
      watch: "^pyproject.toml$",
      run: [
        "bazel run //seed/devprod/python/modules_mapping:generate",
        "bazel run //:gazelle_python_manifest.update",
      ],
    },
    gazelle: {
      target: ["BUILD.bazel"], // the other build files are considered as side effects.
      watch: "\.(go|py)$",
      run: "bazel run //:gazelle",
    },
    bazel: {
      target: ["MODULE.bazel", "MODULE.bazel.lock"],
      watch: "(^|/)BUILD.bazel$",
      run: "bazel mod tidy",
    },
  },
  test: {
    bazel: {
      watch: "^.*$",
      run: "bazel test //...",
    },
  },
} satisfies DirConfig
