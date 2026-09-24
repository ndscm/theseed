"""Build a modules mapping under a fixed target platform.

The gazelle `modules_mapping` rule reads the wheels resolved for the current
target platform. To capture every platform's wheels from a single workstation
we transition the mapping target onto an explicit platform, so `//...:generate`
can regenerate all of the committed per-OS mappings in one `bazel run`.
"""

def _platform_transition_impl(_settings, attr):
    return {"//command_line_option:platforms": str(attr.platform)}

_platform_transition = transition(
    implementation = _platform_transition_impl,
    inputs = [],
    outputs = ["//command_line_option:platforms"],
)

def _modules_mapping_platform_impl(ctx):
    native = ctx.attr.native[0][DefaultInfo].files.to_list()
    if len(native) != 1:
        fail("expected exactly one file from native, got: {}".format(native))
    out = ctx.actions.declare_file(ctx.attr.out)
    ctx.actions.symlink(output = out, target_file = native[0])
    return [DefaultInfo(files = depset([out]))]

modules_mapping_platform = rule(
    implementation = _modules_mapping_platform_impl,
    attrs = {
        "native": attr.label(
            mandatory = True,
            cfg = _platform_transition,
            doc = "The modules_mapping target to evaluate under `platform`.",
        ),
        "platform": attr.label(
            mandatory = True,
            doc = "The platform to resolve wheels for.",
        ),
        "out": attr.string(
            mandatory = True,
            doc = "Name of the generated mapping file.",
        ),
        "_allowlist_function_transition": attr.label(
            default = "@bazel_tools//tools/allowlists/function_transition_allowlist",
        ),
    },
)
