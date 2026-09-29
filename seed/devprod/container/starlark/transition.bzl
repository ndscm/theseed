"""Build container layer targets under a fixed target platform, on any host.

rules_distroless resolves .deb packages only for the architectures it's
configured with (here linux/amd64) and keys its generated targets on
`select({"//:linux_amd64": ...})` with no default condition. On a non-linux host
(e.g. macOS) nothing matches and analysis fails.

Assembling the resulting layer is hermetic and host-independent, though: it needs
only the correct *target* platform, not a matching host machine. These rules
transition their wrapped dependency to a linux platform so the targets build on
any host, forwarding the providers the rules_img image rules consume
(DefaultInfo, LayersInfo, OutputGroupInfo).

Two variants:

  - `layer_linux_x64` hardcodes the linux/amd64 platform.
  - `layer_single_platform` takes the target `platform` as an attribute.
"""

load("@rules_img//img:providers.bzl", "LayersInfo")
load("//seed/devprod/platform:transition.bzl", "platform_transition")
load("//seed/devprod/starlark/transition:linux.bzl", "linux_x64_transition")

def _layer_linux_x64_impl(ctx):
    # Attributes carrying a transition are always lists, even for a 1:1 split.
    target = ctx.attr.actual[0]
    default = target[DefaultInfo]
    providers = [
        DefaultInfo(
            files = default.files,
            runfiles = default.default_runfiles,
        ),
    ]

    # Forward the layer provider so image rules can consume the wrapped target.
    if LayersInfo in target:
        providers.append(target[LayersInfo])
    if OutputGroupInfo in target:
        providers.append(target[OutputGroupInfo])
    return providers

layer_linux_x64 = rule(
    implementation = _layer_linux_x64_impl,
    attrs = {
        "actual": attr.label(
            mandatory = True,
            cfg = linux_x64_transition,
            doc = "Target to build for linux/amd64 regardless of the host.",
        ),
        "_allowlist_function_transition": attr.label(
            default = "@bazel_tools//tools/allowlists/function_transition_allowlist",
        ),
    },
    doc = "Re-exposes `actual`, transitioned to the linux/amd64 platform.",
)

def _layer_single_platform_impl(ctx):
    # Attributes carrying a transition are always lists, even for a 1:1 split.
    target = ctx.attr.native[0]

    # Forward DefaultInfo unchanged (preserving files, runfiles and executable),
    # plus the layer provider so image rules can consume the wrapped target.
    providers = [target[DefaultInfo]]
    if LayersInfo in target:
        providers.append(target[LayersInfo])
    if OutputGroupInfo in target:
        providers.append(target[OutputGroupInfo])
    return providers

layer_single_platform = rule(
    implementation = _layer_single_platform_impl,
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
    doc = "Re-exposes `native`, transitioned to the platform given by `platform`.",
)
