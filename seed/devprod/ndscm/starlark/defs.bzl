"""Bazel rules for copying build outputs into the source tree.

Provides nd_copyout (general-purpose) and nd_bootstrap (tagged for the
bootstrap phase) so that ndscm can materialise generated files before
downstream phases run.
"""

def _nd_copyout_impl(ctx):
    flat_dsts = json.decode(ctx.attr.flat_dsts)

    src_map = {
        str(src.label): src
        for src in ctx.attr.flat_srcs
    }

    all_src_files = []
    dst_map = {}
    for dst_entry in flat_dsts:
        dst_src_files = []
        strip_common_prefix = False
        for src_entry in dst_entry["srcs"]:
            target = src_map[src_entry["label"]]
            output_group = src_entry["output_group"]
            if output_group:
                files = target[OutputGroupInfo][output_group].to_list()
            else:
                files = target.files.to_list()
            dst_src_files.extend(files)
            all_src_files.extend(files)
            if src_entry.get("strip_common_prefix", False):
                strip_common_prefix = True

        srcs_paths = [f.short_path for f in dst_src_files]
        if strip_common_prefix and dst_src_files:
            parts_list = [f.short_path.split("/") for f in dst_src_files]
            common_parts = []
            for i in range(min([len(p) for p in parts_list]) - 1):
                candidate = parts_list[0][i]
                if all([p[i] == candidate for p in parts_list]):
                    common_parts.append(candidate)
                else:
                    break
            srcs_paths = ["/".join(common_parts)]

        dst_map[dst_entry["dst"]] = {
            "dir": dst_entry["dir"],
            "srcs": srcs_paths,
        }

    dst_map_file = ctx.actions.declare_file(ctx.label.name + "_dst_map.json")
    ctx.actions.write(
        output = dst_map_file,
        content = json.encode(dst_map),
    )

    launcher = ctx.actions.declare_file(ctx.label.name + ".sh")
    ctx.actions.write(
        output = launcher,
        content = """#!/usr/bin/env bash
set -euo pipefail
RUNFILES_DIR="$0.runfiles/{workspace}"
cd "$RUNFILES_DIR"
exec "./{copyout}" "${{BUILD_WORKSPACE_DIRECTORY}}/{package}" "./{dst_map}"
""".format(
            workspace = ctx.workspace_name,
            copyout = ctx.executable._copyout.short_path,
            package = ctx.label.package,
            dst_map = dst_map_file.short_path,
        ),
        is_executable = True,
    )

    runfiles = ctx.runfiles(files = all_src_files + [dst_map_file])
    runfiles = runfiles.merge(ctx.attr._copyout[DefaultInfo].default_runfiles)

    return [
        DefaultInfo(
            executable = launcher,
            runfiles = runfiles,
        ),
    ]

_nd_copyout = rule(
    implementation = _nd_copyout_impl,
    executable = True,
    attrs = {
        "flat_srcs": attr.label_list(allow_files = True),
        "flat_dsts": attr.string(),
        "_copyout": attr.label(
            default = "//seed/devprod/ndscm/starlark/copyout",
            executable = True,
            cfg = "exec",
        ),
    },
)

def nd_copyout(name, dsts, **kwargs):
    """Copies bazel-built outputs into the source tree.

    Args:
        name: Rule name.
        dsts: List of dicts, each with "dst" (relative path), optional "dir" (bool),
            and "srcs" (list of dicts with "src" label and optional "output_group").
        **kwargs: Additional args passed to the underlying rule.
    """
    flat_srcs_set = {}
    flat_srcs = []
    flat_dsts = []

    for dst_entry in (dsts or []):
        srcs = []
        for src_entry in dst_entry["srcs"]:
            src_label = src_entry["src"]
            canonical = str(native.package_relative_label(src_label))
            if canonical not in flat_srcs_set:
                flat_srcs_set[canonical] = True
                flat_srcs.append(src_label)
            srcs.append({
                "label": canonical,
                "output_group": src_entry.get("output_group", ""),
                "strip_common_prefix": src_entry.get("strip_common_prefix", False),
            })

        flat_dsts.append({
            "dst": dst_entry["dst"],
            "dir": dst_entry.get("dir", False),
            "srcs": srcs,
        })

    _nd_copyout(
        name = name,
        flat_srcs = flat_srcs,
        flat_dsts = json.encode(flat_dsts),
        **kwargs
    )

def nd_bootstrap(name = "bootstrap", dsts = None):
    nd_copyout(
        name = name,
        dsts = dsts,
        tags = ["nd-bootstrap"],
    )
