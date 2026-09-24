"""Module extension that generates a requirements.txt from a uv.lock.

uv.lock is the single source of truth for the Python dependency graph. During
the loading phase this extension runs the pure-Python converter
(export_requirements.py) under the rules_python host interpreter and writes the
result into a generated repository, so `pip.parse` can consume it directly via
`requirements_lock = "@<name>//:requirements.txt"` without a checked-in
requirements.txt.

The converter reimplements `uv export --format requirements.txt` while
preserving per-package extras; see export_requirements.py for details.
"""

def _load_uv_lock_impl(rctx):
    python = rctx.path(rctx.attr.python)
    script = rctx.path(rctx.attr.script)
    uv_lock = rctx.path(rctx.attr.uv_lock)

    result = rctx.execute([python, script, uv_lock])
    if result.return_code != 0:
        fail("export_requirements failed (exit %d):\n%s" % (
            result.return_code,
            result.stderr,
        ))

    rctx.file("requirements.txt", result.stdout)
    rctx.file(
        "BUILD.bazel",
        'exports_files(["requirements.txt"], visibility = ["//visibility:public"])\n',
    )

_load_uv_lock_rule = repository_rule(
    implementation = _load_uv_lock_impl,
    attrs = {
        "python": attr.label(
            mandatory = True,
            doc = "Host Python interpreter used to run the converter, " +
                  "e.g. @python_3_13_host//:python.",
        ),
        "script": attr.label(
            allow_single_file = True,
            default = "//seed/vendor/python:export_requirements.py",
            doc = "The uv.lock -> requirements.txt converter.",
        ),
        "uv_lock": attr.label(
            mandatory = True,
            allow_single_file = True,
            doc = "The uv.lock to convert.",
        ),
    },
)

_load_uv_lock_class = tag_class(
    attrs = {
        "name": attr.string(
            mandatory = True,
            doc = "Name of the generated repository holding requirements.txt.",
        ),
        "python": attr.label(mandatory = True),
        "uv_lock": attr.label(mandatory = True, allow_single_file = True),
    },
)

def _python_locks_impl(module_ctx):
    for mod in module_ctx.modules:
        for tag in mod.tags.load_uv_lock:
            _load_uv_lock_rule(
                name = tag.name,
                python = tag.python,
                uv_lock = tag.uv_lock,
            )

python_locks = module_extension(
    implementation = _python_locks_impl,
    tag_classes = {
        "load_uv_lock": _load_uv_lock_class,
    },
)
