"""Build container layer targets under a fixed target platform, on any host.

rules_distroless resolves .deb packages only for the architectures it's
configured with (here linux/amd64) and keys its generated targets on
`select({"//:linux_amd64": ...})` with no default condition. On a non-linux host
(e.g. macOS) nothing matches and analysis fails.

Assembling the resulting layer is hermetic and host-independent, though: it needs
only the correct *target* platform, not a matching host machine.
`layer_single_platform` transitions its wrapped dependency to the given
`platform` so the targets build on any host, forwarding the providers the
rules_img image rules consume (DefaultInfo, LayersInfo, OutputGroupInfo).
"""

load("@rules_img//img:providers.bzl", "LayersInfo")
load("//seed/devprod/platform:transition.bzl", "platform_transition")

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
