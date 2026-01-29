"""
DevProd Go Proto Starlark Rules

The `go_connect_library` provides a rule to compile protobuf service into a
connectrpc service The protobuf bazel go team and connectrpc go team don't love
each other. The protoc team wants the language plugins to respect --lang_out
option, and the protobuf go team wants to respect go_package. But the connectrpc
team has their strong opinion to change the package name with "connect" suffix.
So simply applying a customized proto compiler to go_proto_library is not an
option for this case. We create a protoc wrapper to generate the .connect.go
file and feed it to the go_library rule.

"""

load("@protobuf//bazel/common:proto_info.bzl", "ProtoInfo")
load("@rules_go//go:def.bzl", "GoLibrary", "go_library")

def _connect_go_impl(ctx):
    proto_info = ctx.attr.proto[ProtoInfo]
    outputs = []
    for src in proto_info.direct_sources:
        purename = src.basename[:-(len(src.extension) + 1)]
        outputs.append(
            ctx.actions.declare_file(
                purename + "pbconnect/" + purename + ".connect.go",
                sibling = src,
            ),
        )

    inputs = depset(proto_info.direct_sources, transitive = [proto_info.transitive_descriptor_sets])
    args = ctx.actions.args()
    args.add("--plugin")
    args.add_joined(["protoc-gen-connect-go", ctx.executable.protoc_gen_connect_go.path], join_with = "=")
    args.add("--connect-go_opt")
    args.add_joined(["paths", "source_relative"], join_with = "=")
    args.add("--connect-go_out")
    args.add_joined([ctx.bin_dir.path], join_with = "=")
    args.add("--descriptor_set_in")
    args.add_joined(proto_info.transitive_descriptor_sets, join_with = ":")
    args.add_all(proto_info.direct_sources)

    ctx.actions.run(
        executable = ctx.executable.protoc,
        arguments = [args],
        mnemonic = "ProtocGenConnectGo",
        progress_message = "Generating .connect.go from %{label}",
        inputs = inputs,
        outputs = outputs,
        tools = [ctx.executable.protoc_gen_connect_go],
    )

    return [
        DefaultInfo(
            files = depset(outputs),
        ),
    ]

_connect_go = rule(
    implementation = _connect_go_impl,
    attrs = dict({
        "deps": attr.label_list(
            providers = [GoLibrary],
        ),
        "proto": attr.label(
            providers = [ProtoInfo],
            mandatory = True,
        ),
        "protoc": attr.label(
            default = "@protobuf//:protoc",
            executable = True,
            cfg = "exec",
        ),
        "protoc_gen_connect_go": attr.label(
            default = "@com_connectrpc_connect//cmd/protoc-gen-connect-go:protoc-gen-connect-go",
            executable = True,
            cfg = "exec",
        ),
    }),
)

def go_connect_library(name, proto, importpath, **kwargs):
    _connect_go(
        name = name + "_go",
        proto = proto,
    )
    go_library(
        name = name,
        srcs = [":" + name + "_go"],
        importpath = importpath,
        **kwargs
    )
