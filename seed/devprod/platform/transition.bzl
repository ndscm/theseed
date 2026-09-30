"""Build a target under a specific platform, regardless of the host.

Bazel resolves toolchains and platform-dependent dependencies (such as
Python wheels or container layers) based on the target platform of the build.
`single_platform` uses an incoming transition to force its wrapped target to
be configured for a chosen platform, so a single build can materialize the
outputs of the same target for several platforms at once (e.g. resolving both
linux/amd64 and darwin/arm64 wheels without running two separate builds).

Pass `executable = True` for a `native` target you intend to `bazel run` or use
as a tool; otherwise `single_platform` re-exposes file outputs alias-style.
"""

def _platform_transition_impl(_settings, attr):
    return {"//command_line_option:platforms": str(attr.platform)}

platform_transition = transition(
    implementation = _platform_transition_impl,
    inputs = [],
    outputs = ["//command_line_option:platforms"],
)

def _single_platform_library_impl(ctx):
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

single_platform_library = rule(
    implementation = _single_platform_library_impl,
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

def _single_platform_binary_impl(ctx):
    target = ctx.attr.native[0]

    # Bazel requires the executable in DefaultInfo to be created by *this* rule,
    # so we can't forward `native`'s executable directly (that's why plain
    # `single_platform` can't wrap a binary). Declare a symlink owned by this
    # rule pointing at it, and hand the symlink back as our executable so
    # `bazel run` and `$(execpath ...)` both work.
    wrapped = target[DefaultInfo].files_to_run.executable
    exe = ctx.actions.declare_file(ctx.label.name)
    ctx.actions.symlink(output = exe, target_file = wrapped, is_executable = True)
    providers = [DefaultInfo(
        executable = exe,
        files = depset([exe]),
        runfiles = target[DefaultInfo].default_runfiles,
    )]
    if OutputGroupInfo in target:
        providers.append(target[OutputGroupInfo])
    if RunEnvironmentInfo in target:
        providers.append(target[RunEnvironmentInfo])
    return providers

single_platform_binary = rule(
    implementation = _single_platform_binary_impl,
    doc = """Like `single_platform`, but for an executable `native` target.

Re-exposes `native`'s executable (and runfiles) built for `platform`, so the
resulting target can be `bazel run` or used as a tool. Whether a Starlark rule
is executable is fixed at definition time, so binaries need this separate rule
rather than a flag on `single_platform`.
""",
    attrs = {
        "native": attr.label(
            mandatory = True,
            cfg = platform_transition,
            doc = "The executable target to build under `platform`. Its executable and runfiles are forwarded.",
        ),
        "platform": attr.label(
            mandatory = True,
            doc = "The platform to build `native` for.",
        ),
        "_allowlist_function_transition": attr.label(
            default = "@bazel_tools//tools/allowlists/function_transition_allowlist",
        ),
    },
    executable = True,
)

def single_platform(name, native, platform, executable = False, **kwargs):
    """Re-expose `native`'s outputs, built for `platform` regardless of the host.

    Dispatches to the right underlying rule: whether a Starlark rule is
    executable is fixed at definition time, so binaries and libraries need
    separate rules that this macro selects between.

    Args:
        name: Name of the target to create.
        native: The target to build under `platform`. When `executable` is
            False its file outputs are forwarded; when True its executable and
            runfiles are forwarded.
        platform: The platform label to configure `native` for (e.g.
            `//seed/devprod/platform/linux:amd64`).
        executable: Set True when `native` is a binary you intend to `bazel run`
            or use as a tool. Defaults to False (alias-style file outputs).
        **kwargs: Common attributes (`visibility`, `tags`, ...) forwarded to the
            underlying rule.
    """
    if executable:
        single_platform_binary(
            name = name,
            native = native,
            platform = platform,
            **kwargs
        )
    else:
        single_platform_library(
            name = name,
            native = native,
            platform = platform,
            **kwargs
        )
