"""Secret rules"""

def local_secret(name, local_path):
    native.genrule(
        name = name,
        outs = [name.upper()],
        cmd_bash = "cat {local_path} > $@".format(local_path = local_path),
        tags = [
            "local",
            "manual",
            "no-remote-cache",
        ],
    )

def _internal_local_structure_secret_impl(ctx):
    out = ctx.actions.declare_directory(ctx.attr.name + ".dir")
    ctx.actions.run_shell(
        outputs = [out],
        command = """mkdir -p $(dirname {out}/{structure_path}) && cat {local_path} > {out}/{structure_path}""".format(
            out = out.path,
            structure_path = ctx.attr.structure_path,
            local_path = ctx.attr.local_path,
        ),
        mnemonic = "LocalStructureSecret",
        progress_message = "Generating directory: {out}".format(out = out.path),
    )
    return [DefaultInfo(files = depset([out]))]

_intenal_local_structure_secret = rule(
    implementation = _internal_local_structure_secret_impl,
    attrs = {
        "structure_path": attr.string(),
        "local_path": attr.string(),
    },
    doc = "Generates a directory with structured secret file.",
)

def local_structure_secret(name, structure_path, local_path, tags = None, **kwargs):
    _intenal_local_structure_secret(
        name = name,
        structure_path = structure_path,
        local_path = local_path,
        tags = list(set(
            (tags if tags else []) +
            ["local", "manual", "no-remote-cache"],
        )),
        **kwargs
    )
