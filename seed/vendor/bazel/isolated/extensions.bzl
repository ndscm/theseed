load("@bazel_tools//tools/build_defs/repo:http.bzl", "http_archive")

def _bazel_impl(_mctx):
    http_archive(
        name = "bazel",
        integrity = "sha256-s+9rSX6wrF/jyazGKsiudT507M6d7HHowE6MksM5jIE=",
        strip_prefix = "bazel-8.6.0",
        urls = ["https://github.com/bazelbuild/bazel/archive/refs/tags/8.6.0.tar.gz"],
    )

bazel = module_extension(implementation = _bazel_impl)
