"""Run a js_binary as a build action and package the output as a tarball.

A lightweight alternative to aspect_rules_js's js_run_binary that produces a
single tar archive from a declared output directory. A Go helper (jsruntar)
copies source-tree files into the output tree, invokes the js_binary tool,
and archives the result — replacing shell commands with a portable binary
that works across operating systems.

Based on https://github.com/aspect-build/rules_js/blob/v3.1.0/js/private/js_run_binary.bzl
"""

load("@aspect_rules_js//js:providers.bzl", "JsInfo")

def _js_run_binary_tar_impl(ctx):
    out_dir = ctx.actions.declare_directory(ctx.attr.out_dir)
    out_tar = ctx.actions.declare_file(ctx.attr.out_tar)

    transitive = []
    for src in ctx.attr.srcs:
        transitive.append(src.files)
        if JsInfo in src:
            js_info = src[JsInfo]
            transitive.append(js_info.transitive_sources)
            transitive.append(js_info.npm_sources)

    tool_runfiles = ctx.attr.tool[DefaultInfo].default_runfiles
    if tool_runfiles:
        transitive.append(tool_runfiles.files)

    all_inputs = depset(transitive = transitive)

    copies = {}
    for src in ctx.attr.srcs:
        for f in src.files.to_list():
            if f.is_source:
                copies[f.path] = ctx.bin_dir.path + "/" + f.short_path

    copies_json = ctx.actions.declare_file(ctx.attr.name + "_copies.json")
    ctx.actions.write(output = copies_json, content = json.encode(copies))

    env = {
        "BAZEL_BINDIR": ctx.bin_dir.path,
        "BAZEL_BUILD_FILE_PATH": ctx.build_file_path,
        "BAZEL_COMPILATION_MODE": ctx.var["COMPILATION_MODE"],
        "BAZEL_PACKAGE": ctx.label.package,
        "BAZEL_TARGET_CPU": ctx.var["TARGET_CPU"],
        "BAZEL_TARGET_NAME": ctx.label.name,
        "BAZEL_TARGET": str(ctx.label),
        "BAZEL_WORKSPACE": ctx.workspace_name,
        "JS_BINARY__CHDIR": ctx.attr.chdir if ctx.attr.chdir else ctx.label.package,
        "JS_BINARY__SILENT_ON_SUCCESS": "1",
        "JS_BINARY__USE_EXECROOT_ENTRY_POINT": "1",
        "JS_BINARY__PATCH_NODE_FS": "1",
    }
    env.update(ctx.attr.env)

    ctx.actions.run(
        outputs = [out_dir, out_tar],
        inputs = depset([copies_json], transitive = [all_inputs]),
        executable = ctx.executable.jsruntar,
        tools = [ctx.attr.tool[DefaultInfo].files_to_run],
        arguments = [
            "--out_dir",
            out_dir.path,
            "--out_tar",
            out_tar.path,
            "--copies",
            copies_json.path,
            "--tool",
            ctx.executable.tool.path,
            "--",
        ] + ctx.attr.args,
        env = env,
        mnemonic = ctx.attr.mnemonic,
        progress_message = ctx.attr.progress_message,
    )

    return [DefaultInfo(files = depset([out_tar]))]

_js_run_binary_tar = rule(
    implementation = _js_run_binary_tar_impl,
    attrs = {
        "srcs": attr.label_list(allow_files = True),
        "tool": attr.label(executable = True, cfg = "exec"),
        "args": attr.string_list(),
        "out_dir": attr.string(),
        "out_tar": attr.string(),
        "chdir": attr.string(),
        "env": attr.string_dict(),
        "mnemonic": attr.string(default = "JsRunBinary"),
        "progress_message": attr.string(),
        "jsruntar": attr.label(
            mandatory = True,
            executable = True,
            cfg = "exec",
        ),
    },
    doc = "Runs a js_binary tool and archives its output directory into a tarball.",
)

def js_run_binary_tar(
        name,
        srcs,
        tool,
        args,
        out_dir,
        out_tar,
        chdir = "",
        env = {},
        mnemonic = "JsRunBinary",
        progress_message = "",
        **kwargs):
    """Runs a js_binary tool and archives its output directory into a tarball.

    This macro wraps the underlying rule with the same interface. Source files
    from the source tree are automatically copied into the output tree before
    the tool runs, so a js_binary that expects to find its sources relative to
    its working directory (set via ``chdir``) will resolve them correctly.

    Transitive npm packages and source files are gathered from JsInfo providers
    on ``srcs`` targets, so a ``:node_modules`` label in ``srcs`` is sufficient
    to make the full dependency graph available to the tool.

    Args:
        name: Target name.
        srcs: Source files and dependencies for the build action.
        tool: A js_binary target to run.
        args: Command-line arguments passed to the tool.
        out_dir: Name of the output directory the tool writes to.
        out_tar: Filename for the resulting tar archive.
        chdir: Working directory for the tool, relative to the execroot.
            Defaults to the current package path.
        env: Additional environment variables for the action.
        mnemonic: A short label for the action shown in build output.
        progress_message: Message displayed during the build. Supports
            Bazel's ``%{label}`` substitution.
        **kwargs: Forwarded to the underlying rule (e.g. ``tags``,
            ``visibility``).
    """
    _js_run_binary_tar(
        name = name,
        srcs = srcs,
        tool = tool,
        args = args,
        out_dir = out_dir,
        out_tar = out_tar,
        chdir = chdir,
        env = env,
        mnemonic = mnemonic,
        progress_message = progress_message,
        jsruntar = "//seed/devprod/starlark/jsrun/jsruntar",
        **kwargs
    )
