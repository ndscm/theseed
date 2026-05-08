"""Rule to expose the Go SDK's gofmt binary from the resolved Go toolchain."""

load("@platforms//host:constraints.bzl", "HOST_CONSTRAINTS")

def _gofmt_bin(ctx):
    sdk = ctx.toolchains["@rules_go//go:toolchain"].sdk
    gofmt = None
    for f in sdk.tools.to_list():
        if f.basename == "gofmt" or f.basename == "gofmt.exe":
            gofmt = f
            break
    if gofmt == None:
        fail("gofmt not found in Go SDK tools")
    out = ctx.actions.declare_file(ctx.label.name)
    ctx.actions.symlink(output = out, target_file = gofmt, is_executable = True)
    return [
        DefaultInfo(
            files = depset([out]),
            runfiles = ctx.runfiles([gofmt]),
            executable = out,
        ),
    ]

gofmt_bin = rule(
    implementation = _gofmt_bin,
    toolchains = ["@rules_go//go:toolchain"],
    exec_compatible_with = HOST_CONSTRAINTS,
    executable = True,
)
