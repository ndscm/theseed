"""Bazel rules for building multi-platform artifacts using transitions."""

def _android_arm64_transition_impl(_settings, _attr):
    """Transition to build for Android ARM64."""
    return {
        "//command_line_option:platforms": "//:android-arm64",
    }

android_arm64_transition = transition(
    implementation = _android_arm64_transition_impl,
    inputs = [],
    outputs = ["//command_line_option:platforms"],
)

def _transitioned_binary_impl(ctx):
    """Rule implementation that copies the transitioned binary."""
    src = ctx.file.binary
    out = ctx.actions.declare_file(ctx.attr.name)
    ctx.actions.symlink(output = out, target_file = src)
    return [DefaultInfo(
        files = depset([out]),
        runfiles = ctx.runfiles(files = [out]),
    )]

android_arm64_binary = rule(
    implementation = _transitioned_binary_impl,
    attrs = {
        "binary": attr.label(
            mandatory = True,
            allow_single_file = True,
            cfg = android_arm64_transition,
        ),
        "_allowlist_function_transition": attr.label(
            default = "@bazel_tools//tools/allowlists/function_transition_allowlist",
        ),
    },
)
