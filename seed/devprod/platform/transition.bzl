"""Build a target under a specific platform, regardless of the host.

Bazel resolves toolchains and platform-dependent dependencies (such as
Python wheels or container layers) based on the target platform of the build.
`single_platform` uses an incoming transition to force its wrapped target to
be configured for a chosen platform, so a single build can materialize the
outputs of the same target for several platforms at once (e.g. resolving both
linux/amd64 and darwin/arm64 wheels without running two separate builds).
"""

def _platform_transition_impl(_settings, attr):
    return {"//command_line_option:platforms": str(attr.platform)}

platform_transition = transition(
    implementation = _platform_transition_impl,
    inputs = [],
    outputs = ["//command_line_option:platforms"],
)

def _single_platform_impl(ctx):
    # `native` is transitioned to `platform`, so it comes back as a 1-element
    # list. Forward its providers unchanged, alias-style. Starlark can't
    # enumerate providers generically (unlike the native `alias` rule), so we
    # forward the ones consumers rely on: DefaultInfo for file outputs and
    # OutputGroupInfo for named groups.
    target = ctx.attr.native[0]
    providers = [target[DefaultInfo]]
    if OutputGroupInfo in target:
        providers.append(target[OutputGroupInfo])
    return providers

single_platform = rule(
    implementation = _single_platform_impl,
    doc = """Re-expose a target's outputs, built for a specific platform.

Behaves like `alias`, but forces `native` to be configured for `platform`
via an incoming transition. This lets the same underlying target be built
for multiple platforms in one build, each `single_platform` instance
yielding the outputs for its own platform.
""",
    attrs = {
        "native": attr.label(
            mandatory = True,
            cfg = platform_transition,
            doc = "The target to build under `platform`. Its outputs are forwarded unchanged.",
        ),
        "platform": attr.label(
            mandatory = True,
            doc = "The platform to build `native` for.",
        ),
        "_allowlist_function_transition": attr.label(
            default = "@bazel_tools//tools/allowlists/function_transition_allowlist",
        ),
    },
)
