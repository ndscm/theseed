"""Repository rule that generates a platform-selecting repo for Android SDK Command-Line Tools."""

def _android_cmdline_tools_impl(repository_ctx):
    repository_ctx.file("BUILD.bazel", content = """
package(default_visibility = ["//visibility:public"])

alias(
    name = "files",
    actual = select({
        "@platforms//os:linux": "@android_cmdline-tools_linux//:files",
        "@platforms//os:macos": "@android_cmdline-tools_mac//:files",
    }),
)

alias(
    name = "sdkmanager",
    actual = select({
        "@platforms//os:linux": "@android_cmdline-tools_linux//:sdkmanager",
        "@platforms//os:macos": "@android_cmdline-tools_mac//:sdkmanager",
    }),
)
""")

android_cmdline_tools = repository_rule(
    implementation = _android_cmdline_tools_impl,
)
