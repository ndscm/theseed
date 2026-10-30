"""Pack a complete xslash entry into a tar at build time.

``xslash_pack`` runs ``xslash-helper create`` to rewrite a skeleton entry (e.g.
toggling ``final``) and generate its side-by-side symlinks, then archives the
entry and symlinks into a single ``.tar`` - symlinks aren't representable as
plain Bazel outputs.

The skeleton must already carry every platform's ``size``/``hash``/``digest``,
so ``create`` trusts them and runs offline, needing no network or cache.
Skeletons with missing digests (which would need network probing) are
unsupported.
"""

def _xslash_pack_impl(ctx):
    out_entry = ctx.attr.out
    out_tar = ctx.actions.declare_file(ctx.label.name + ".tar")

    # xslashtar runs `xslash-helper create` into a scratch dir and tars the
    # result, preserving the symlinks. It supplies --out itself; the rest of the
    # create invocation follows the `--` terminator.
    args = ctx.actions.args()
    args.add("--xslash_helper", ctx.executable._xslash_helper)
    args.add("--out_entry", out_entry)
    args.add("--out_tar", out_tar)
    args.add("--")
    args.add("create")
    args.add("--skeleton", ctx.file.skeleton)
    if ctx.attr.final:
        args.add("--final")
    for name in ctx.attr.symlinks:
        args.add("--symlink", name)

    ctx.actions.run(
        outputs = [out_tar],
        inputs = [ctx.file.skeleton],
        executable = ctx.executable._xslashtar,
        tools = [ctx.executable._xslash_helper],
        arguments = [args],
        mnemonic = "XslashPack",
        progress_message = "Packing xslash entry %{label}",
    )

    return [DefaultInfo(files = depset([out_tar]))]

_xslash_pack = rule(
    implementation = _xslash_pack_impl,
    attrs = {
        "skeleton": attr.label(
            allow_single_file = True,
            mandatory = True,
            doc = "Complete xslash entry to rewrite.",
        ),
        "final": attr.bool(
            default = False,
            doc = "Set the `final` flag in the generated entry.",
        ),
        "out": attr.string(
            mandatory = True,
            doc = "Name of the generated entry, including the `.xslash` suffix.",
        ),
        "symlinks": attr.string_list(
            doc = "Bare file names to symlink to the entry, beside it.",
        ),
        "_xslash_helper": attr.label(
            default = "//seed/prebuilt/xslash:xslash-helper",
            executable = True,
            cfg = "exec",
        ),
        "_xslashtar": attr.label(
            default = "//seed/devprod/starlark/xslash/xslashtar",
            executable = True,
            cfg = "exec",
        ),
    },
    doc = "Packs a rewritten xslash entry and its symlinks into a tar.",
)

def xslash_pack(name, skeleton, out, final = False, symlinks = [], **kwargs):
    """Pack a rewritten xslash entry into a tar; see the module docs.

    The entry is ``out`` and the tar is ``<name>.tar``.

    Args:
        name: Target name; the tar is ``<name>.tar``.
        skeleton: Complete xslash entry file to rewrite.
        final: Whether to set the ``final`` flag.
        out: Name of the generated entry, including the ``.xslash`` suffix.
        symlinks: Bare file names to symlink to the entry, beside it.
        **kwargs: Forwarded to the underlying rule (e.g. ``tags``,
            ``visibility``).
    """
    _xslash_pack(
        name = name,
        skeleton = skeleton,
        final = final,
        out = out,
        symlinks = symlinks,
        **kwargs
    )
