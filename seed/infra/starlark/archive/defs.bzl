"""Archive rules"""

def _internal_create_impl(ctx):
    # out = ctx.actions.declare_file(ctx.attr.name + ".zip")
    out = ctx.outputs.out
    ctx.actions.run_shell(
        outputs = [out],
        inputs = ctx.files.srcs,
        tools = [ctx.executable.create_tool],
        command = """{create_tool_path} --out "{out}" --subdir "{subdir}" {strip_components_flag} {srcs}""".format(
            create_tool_path = ctx.executable.create_tool.path,
            out = out.path,
            subdir = ctx.attr.subdir,
            strip_components_flag = "--strip_components {v}".format(v = ctx.attr.strip_components) if ctx.attr.strip_components >= 0 else "",
            srcs = " ".join([src.path for src in ctx.files.srcs]),
        ),
        mnemonic = "CreateArchive",
        progress_message = "Generating directory: {out}".format(out = out.path),
    )
    return [DefaultInfo(files = depset([out]))]

_internal_create = rule(
    implementation = _internal_create_impl,
    attrs = {
        "srcs": attr.label_list(
            allow_files = True,
        ),
        "out": attr.output(
            mandatory = True,
        ),
        "subdir": attr.string(),
        "strip_components": attr.int(),
        "create_tool": attr.label(
            mandatory = True,
            executable = True,
            cfg = "exec",
        ),
    },
    doc = "Generates a directory with some dummy files inside.",
)

def create_tar_gz(name, srcs, out, subdir = "", strip_components = -1, **kwargs):
    _internal_create(
        name = name,
        srcs = srcs,
        out = out,
        subdir = subdir,
        strip_components = strip_components,
        create_tool = "//seed/infra/starlark/archive:create_tar_gz",
        **kwargs
    )

def create_zip(name, srcs, out, subdir = "", strip_components = -1, **kwargs):
    _internal_create(
        name = name,
        srcs = srcs,
        out = out,
        subdir = subdir,
        strip_components = strip_components,
        create_tool = "//seed/infra/starlark/archive:create_zip",
        **kwargs
    )

def _internal_extract_impl(ctx):
    out = ctx.actions.declare_directory(ctx.attr.name)
    ctx.actions.run_shell(
        outputs = [out],
        inputs = ctx.files.srcs,
        tools = [ctx.executable.extract_tool],
        command = """{extract_tool_path} --out "{out}" --subdir "{subdir}" {srcs}""".format(
            extract_tool_path = ctx.executable.extract_tool.path,
            out = out.path,
            subdir = ctx.attr.subdir,
            srcs = " ".join([src.path for src in ctx.files.srcs]),
        ),
        mnemonic = "ExtractArchive",
        progress_message = "Generating directory: {out}".format(out = out.path),
    )
    return [DefaultInfo(files = depset([out]))]

_internal_extract = rule(
    implementation = _internal_extract_impl,
    attrs = {
        "srcs": attr.label_list(
            allow_files = True,
        ),
        "subdir": attr.string(),
        "extract_tool": attr.label(
            mandatory = True,
            executable = True,
            cfg = "exec",
        ),
    },
    doc = "Generates a directory with some dummy files inside.",
)

def extract_tar(name, srcs, subdir = "", **kwargs):
    _internal_extract(
        name = name,
        srcs = srcs,
        subdir = subdir,
        extract_tool = "//seed/infra/starlark/archive:extract_tar",
        **kwargs
    )
