load("@bazel_tools//tools/build_defs/repo:http.bzl", "http_archive")

def _bazel_impl(_mctx):
    http_archive(
        name = "bazel",
        integrity = "sha256-/SOC9qhPwA34GbnibFQ1vzgzg2LEO34INNv8OMuPhQ4=",
        strip_prefix = "bazel-9.2.0",
        urls = ["https://github.com/bazelbuild/bazel/archive/refs/tags/9.2.0.tar.gz"],
    )

bazel = module_extension(implementation = _bazel_impl)
