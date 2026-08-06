"""Container image rules backed by a system container engine."""

load("@bazel_skylib//rules:common_settings.bzl", "BuildSettingInfo")
load("@seed_devprod_container_starlark_engine_detect//:local.bzl", "LOCAL_CONTAINER_ENGINES")

# build_freedom_image needs a local container engine on PATH. When none is
# detected (CI, most workstations), mark the target incompatible so
# `bazel build //...` skips it instead of failing when it shells out to a
# missing engine. See //seed/devprod/container/starlark:engine.bzl.
_ENGINE_COMPATIBLE_WITH = [] if LOCAL_CONTAINER_ENGINES else ["@platforms//:incompatible"]

def _build_freedom_image_impl(ctx):
    out = ctx.actions.declare_file(ctx.label.name + ".tar")
    base = ctx.file.base
    containerfile = ctx.file.containerfile
    image_tag = ctx.attr.image_tag
    engine = ctx.attr._engine[BuildSettingInfo].value

    # Rootless podman's default subuid/subgid range is too small for RUN steps
    # that create users or chown to high UIDs (e.g. postinst scripts). Give the
    # build a full user namespace.
    build_opts = ""
    if "podman" in engine:
        build_opts = "--userns auto:size=65536"
        if ctx.attr.squash:
            build_opts += " --squash"
    else:
        fail("Unsupported engine: {engine}".format(engine = engine))

    # Load the Bazel-built base OCI layout into the engine's storage and pass its
    # image id to `--from`, overriding the first FROM. This keeps the base
    # hermetic instead of pulling from a registry; `--pull=never` ensures the
    # base comes only from what we just loaded.
    ctx.actions.run_shell(
        outputs = [out],
        inputs = [containerfile, base] + ctx.files.srcs,
        # Assemble a build context from the Containerfile plus srcs. srcs come
        # from all over the build graph, so stage them by basename into a scratch
        # context dir where `podman build` can COPY them.
        command = """
set -euxo pipefail

context_dir="$(mktemp -d)"
# srcs may include read-only tree artifacts; make the staged copies writable so
# the cleanup below can remove them.
trap 'chmod -R u+w "$context_dir" 2>/dev/null || true; rm -rf "$context_dir"' EXIT

cp -L "{containerfile}" "$context_dir/Containerfile"

# srcs is newline-separated so paths with spaces survive word splitting; the
# variable also keeps an empty srcs from becoming a `for f in ; do` error.
srcs='{srcs}'
IFS='
'
for f in $srcs; do
  cp -RL "$f" "$context_dir/"
done
unset IFS

ls -ahl "$context_dir"

base_id="$("{engine}" pull -q "oci:{base}")"
"{engine}" build {build_opts} \\
  --pull=never \\
  --from "${{base_id}}" \\
  -f "$context_dir/Containerfile" \\
  -t "{image_tag}" \\
  "$context_dir"
"{engine}" save -o "{out}" "{image_tag}"
""".format(
            engine = engine,
            base = base.path,
            build_opts = build_opts,
            containerfile = containerfile.path,
            image_tag = image_tag,
            srcs = "\n".join([f.path for f in ctx.files.srcs]),
            out = out.path,
        ),
        # The engine comes from PATH, uses the host's container storage and runs
        # apt inside the build, so it can't run sandboxed or on a remote executor.
        use_default_shell_env = True,
        execution_requirements = {
            "no-sandbox": "1",
            "no-remote": "1",
            "local": "1",
        },
        mnemonic = "BuildFreedomImage",
        progress_message = "Building container image: {out}".format(out = out.short_path),
    )
    return [DefaultInfo(files = depset([out]))]

_build_freedom_image = rule(
    implementation = _build_freedom_image_impl,
    attrs = {
        "base": attr.label(
            allow_single_file = True,
            mandatory = True,
            doc = "OCI image layout used as the base (overrides the first FROM).",
        ),
        "containerfile": attr.label(
            allow_single_file = True,
            mandatory = True,
            doc = "The Containerfile describing the image to build.",
        ),
        "image_tag": attr.string(
            mandatory = True,
            doc = "Tag applied to the built image.",
        ),
        "srcs": attr.label_list(
            allow_files = True,
            doc = "Additional files that make up the build context.",
        ),
        "squash": attr.bool(
            default = False,
            doc = "Pass podman --squash to collapse the built layers into one.",
        ),
        "_engine": attr.label(
            default = "//seed/devprod/container/starlark:engine",
            providers = [BuildSettingInfo],
            doc = "The container engine build setting (only podman is supported yet).",
        ),
    },
    doc = "Builds a container image with the system engine and saves it to a tar.",
)

def build_freedom_image(
        name,
        base,
        containerfile,
        image_tag,
        srcs = [],
        squash = False,
        **kwargs):
    """Builds a "freedom image": a non-hermetic container image.

    Unlike rules_oci or rules_img, which assemble layers hermetically without
    running the image, this drives a real engine (podman; docker not supported
    yet) to build the Containerfile. RUN steps and maintainer scripts (dpkg
    postinst, ldconfig, ...) execute for real — the "freedom" being that the
    build isn't confined to the Bazel sandbox. The tradeoff: it needs the engine
    on PATH, uses the host's container storage, and can't run sandboxed or
    remotely. Mainly for manual debugging; prefer rules_img for production images.

    The target is `target_compatible_with` a detected engine (see engine.bzl), so
    a host with none on PATH skips it as "incompatible" instead of breaking
    `bazel build //...`.

    Args:
        name: Target name; the built image is saved to <name>.tar.
        base: OCI image layout used as the base (overrides the first FROM).
        containerfile: The Containerfile describing the image to build.
        image_tag: Tag applied to the built image.
        srcs: Additional files staged into the build context.
        squash: Pass podman --squash to collapse the built layers into one (default False).
        **kwargs: Passed through to the underlying rule (e.g. visibility).
    """
    _build_freedom_image(
        name = name,
        base = base,
        containerfile = containerfile,
        image_tag = image_tag,
        srcs = srcs,
        squash = squash,
        target_compatible_with = _ENGINE_COMPATIBLE_WITH,
        **kwargs
    )
