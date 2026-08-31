"""Force container image targets onto the linux/amd64 platform.

rules_distroless resolves .deb packages only for the architectures it's
configured with (here linux/amd64) and keys its generated targets on
`select({"//:linux_amd64": ...})` with no default condition. On a non-linux host
(e.g. macOS) nothing matches and analysis fails.

Assembling the resulting layer is hermetic and host-independent, though: it needs
only the linux/amd64 *target* platform, not a linux machine. `linux_amd64`
transitions its dependency to that platform so these targets build on any host,
forwarding the providers the rules_img image rules consume.
"""

load("@rules_img//img:providers.bzl", "LayersInfo")
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
