"""Bazel rules and macros for Ent ORM code generation.

Runs `ent generate` in GOPATH mode (`GO111MODULE=off`) against a hermetic
GOPATH built by `go_path`, then exposes generated sources as `go_library`
targets grouped by package.

Rules:
    _ent_full_srcs: Generates all Ent source files and returns them via
        OutputGroupInfo with static groups (ent, enttest, hook, migrate,
        predicate, runtime) and one group per entity.

Macros:
    ent_full_srcs: Sets up go_path and invokes _ent_full_srcs.
    go_ent_library: Creates go_library targets for every output group.
"""

load("@rules_go//go:def.bzl", "GoInfo", "GoPath", "go_context", "go_library", "go_path")

def _ent_full_srcs_impl(ctx):
    go = go_context(ctx)
    schema_importpath = ctx.attr.schema[GoInfo].importpath
    target_importpath = ctx.attr.importpath

    entities = []
    for f in ctx.attr.schema[GoInfo].srcs:
        if f.basename.endswith(".go"):
            entity = f.basename[:-3]
            entities.append(entity)
    entities = sorted(entities)
    if sorted(ctx.attr.entities) != entities:
        fail("Please specify entities: {entities}".format(entities = entities))

    output_groups = {}
    output_groups["ent"] = []
    output_groups["ent"].append(ctx.actions.declare_file("ent/client.go"))
    output_groups["ent"].append(ctx.actions.declare_file("ent/ent.go"))
    output_groups["ent"].append(ctx.actions.declare_file("ent/generate.go"))
    output_groups["ent"].append(ctx.actions.declare_file("ent/mutation.go"))
    output_groups["ent"].append(ctx.actions.declare_file("ent/runtime.go"))
    output_groups["ent"].append(ctx.actions.declare_file("ent/tx.go"))
    output_groups["enttest"] = []
    output_groups["enttest"].append(ctx.actions.declare_file("ent/enttest/enttest.go"))
    output_groups["hook"] = []
    output_groups["hook"].append(ctx.actions.declare_file("ent/hook/hook.go"))
    output_groups["migrate"] = []
    output_groups["migrate"].append(ctx.actions.declare_file("ent/migrate/migrate.go"))
    output_groups["migrate"].append(ctx.actions.declare_file("ent/migrate/schema.go"))
    output_groups["predicate"] = []
    output_groups["predicate"].append(ctx.actions.declare_file("ent/predicate/predicate.go"))
    output_groups["runtime"] = []
    output_groups["runtime"].append(ctx.actions.declare_file("ent/runtime/runtime.go"))
    for entity in entities:
        output_groups["ent"].append(ctx.actions.declare_file("ent/{entity}_create.go".format(entity = entity)))
        output_groups["ent"].append(ctx.actions.declare_file("ent/{entity}_delete.go".format(entity = entity)))
        output_groups["ent"].append(ctx.actions.declare_file("ent/{entity}_query.go".format(entity = entity)))
        output_groups["ent"].append(ctx.actions.declare_file("ent/{entity}_update.go".format(entity = entity)))
        output_groups["ent"].append(ctx.actions.declare_file("ent/{entity}.go".format(entity = entity)))
        output_groups[entity] = []
        output_groups[entity].append(ctx.actions.declare_file("ent/{entity}/{entity}.go".format(entity = entity)))
        output_groups[entity].append(ctx.actions.declare_file("ent/{entity}/where.go".format(entity = entity)))

    outputs = []
    for output_group in output_groups.values():
        outputs.extend(output_group)

    ctx.actions.run_shell(
        mnemonic = "EntGenerate",
        progress_message = "Generating ent source in {target_dir}".format(target_dir = outputs[0].dirname),
        command = """
set -eu

export PATH="$PWD/{go_bin_dir}:$PATH"
export GOCACHE="$PWD/.gocache"
export GOPATH="$PWD/{gopath_dir}"
export GO111MODULE=off

mkdir -p "$GOPATH/src/{target_importpath}"
printf "package ent\\n" > "$GOPATH/src/{target_importpath}/generate.go"

"{ent_bin}" generate --target "$GOPATH/src/{target_importpath}" "$GOPATH/src/{schema_importpath}"

mkdir -p "{target_dir}"
cp -r "$GOPATH/src/{target_importpath}/"* "{target_dir}"
""".format(
            go_bin_dir = go.sdk.go.dirname,
            ent_bin = ctx.executable._ent_bin.path,
            gopath_dir = ctx.attr.gopath[GoPath].gopath_file.path,
            target_dir = outputs[0].dirname,
            schema_importpath = schema_importpath,
            target_importpath = target_importpath,
        ),
        tools = [ctx.executable._ent_bin, go.sdk.go],
        inputs = [ctx.attr.gopath[GoPath].gopath_file],
        outputs = outputs,
    )

    return [
        DefaultInfo(
            files = depset(outputs),
        ),
        OutputGroupInfo(**output_groups),
    ]

_ent_full_srcs = rule(
    implementation = _ent_full_srcs_impl,
    attrs = {
        "schema": attr.label(
            mandatory = True,
            providers = [GoInfo],
        ),
        "entities": attr.string_list(mandatory = True),
        "importpath": attr.string(mandatory = True),
        "gopath": attr.label(
            mandatory = True,
            providers = [GoPath],
        ),
        "_go_context_data": attr.label(
            default = Label("@rules_go//:go_context_data"),
        ),
        "_ent_bin": attr.label(
            executable = True,
            default = Label("@io_entgo_ent//cmd/ent"),
            cfg = "exec",
        ),
    },
    toolchains = ["@rules_go//go:toolchain"],
)

def ent_full_srcs(name, schema, entities, importpath, **kwargs):
    """Generate Ent sources with grouped outputs.

    Creates a go_path dependency and an _ent_full_srcs target that provides
    all generated files via DefaultInfo and OutputGroupInfo.

    Args:
        name: Target name. Also creates `<name>_gopath` internally.
        schema: Label of the Ent schema go_library.
        entities: Entity names matching schema .go filenames (without extension).
        importpath: Go import path for the generated ent package.
        **kwargs: Forwarded to _ent_full_srcs.
    """
    go_path(
        name = name + "_gopath",
        deps = [
            schema,
            "@io_entgo_ent//cmd/ent",
        ],
        visibility = ["//visibility:private"],
    )

    _ent_full_srcs(
        name = name,
        schema = schema,
        entities = entities,
        gopath = ":" + name + "_gopath",
        importpath = importpath,
        **kwargs
    )

def _ent_gazelle_impl(ctx):
    importpath = ctx.attr.importpath
    target = ctx.attr.target
    sub_packages = sorted(["enttest", "hook", "migrate", "predicate", "runtime"] + list(ctx.attr.entities))

    content = "# gazelle:resolve go {importpath} {target}\n".format(
        importpath = importpath,
        target = target,
    )
    for sub in sub_packages:
        content += "# gazelle:resolve go {importpath}/{sub} {target}/{sub}\n".format(
            importpath = importpath,
            sub = sub,
            target = target,
        )

    output = ctx.actions.declare_file(ctx.label.name + ".txt")
    ctx.actions.write(output = output, content = content)
    return [DefaultInfo(files = depset([output]))]

_ent_gazelle = rule(
    implementation = _ent_gazelle_impl,
    attrs = {
        "entities": attr.string_list(mandatory = True),
        "importpath": attr.string(mandatory = True),
        "target": attr.string(mandatory = True),
    },
)

def go_ent_library(name, schema, entities, importpath, **kwargs):
    """Generate go_library targets for all Ent output groups.

    Creates one go_library per output group from ent_full_srcs: the main ent
    package, enttest, hook, migrate, predicate, runtime, and one per entity.

    Args:
        name: Base target name. Sub-packages use `<name>/<group>` naming.
        schema: Label of the Ent schema go_library.
        entities: Entity names matching schema .go filenames (without extension).
        importpath: Go import path for the generated ent package.
        **kwargs: Forwarded to ent_full_srcs.
    """
    ent_full_srcs(
        name = name + "_full_srcs",
        schema = schema,
        entities = entities,
        importpath = importpath,
        **kwargs
    )

    native.filegroup(
        name = name + "/migrate_srcs",
        srcs = [":" + name + "_full_srcs"],
        output_group = "migrate",
    )
    go_library(
        name = name + "/migrate",
        srcs = [":" + name + "/migrate_srcs"],
        importpath = importpath + "/migrate",
        deps = [
            "@io_entgo_ent//dialect",
            "@io_entgo_ent//dialect/entsql",
            "@io_entgo_ent//dialect/sql/schema",
            "@io_entgo_ent//schema/field",
        ],
    )

    native.filegroup(
        name = name + "/predicate_srcs",
        srcs = [":" + name + "_full_srcs"],
        output_group = "predicate",
    )
    go_library(
        name = name + "/predicate",
        srcs = [":" + name + "/predicate_srcs"],
        importpath = importpath + "/predicate",
        deps = [
            "@io_entgo_ent//dialect/sql",
        ],
    )

    native.filegroup(
        name = name + "/runtime_srcs",
        srcs = [":" + name + "_full_srcs"],
        output_group = "runtime",
    )
    go_library(
        name = name + "/runtime",
        srcs = [":" + name + "/runtime_srcs"],
        importpath = importpath + "/runtime",
    )

    for entity in entities:
        native.filegroup(
            name = name + "/" + entity + "_srcs",
            srcs = [":" + name + "_full_srcs"],
            output_group = entity,
        )
        go_library(
            name = name + "/" + entity,
            srcs = [":" + name + "/" + entity + "_srcs"],
            importpath = importpath + "/" + entity,
            deps = [
                ":" + name + "/predicate",
                "@io_entgo_ent//dialect/sql",
            ],
        )

    native.filegroup(
        name = name + "_srcs",
        srcs = [":" + name + "_full_srcs"],
        output_group = "ent",
    )
    go_library(
        name = name,
        srcs = [":" + name + "_srcs"],
        importpath = importpath,
        deps = [
            schema,
            ":" + name + "/migrate",
            ":" + name + "/predicate",
            ":" + name + "/runtime",
            "@io_entgo_ent//:ent",
            "@io_entgo_ent//dialect",
            "@io_entgo_ent//dialect/sql",
            "@io_entgo_ent//dialect/sql/sqlgraph",
            "@io_entgo_ent//schema/field",
        ] + [
            ":" + name + "/" + entity
            for entity in entities
        ],
    )

    native.filegroup(
        name = name + "/hook_srcs",
        srcs = [":" + name + "_full_srcs"],
        output_group = "hook",
    )
    go_library(
        name = name + "/hook",
        srcs = [":" + name + "/hook_srcs"],
        importpath = importpath + "/hook",
        deps = [
            ":" + name,
        ],
    )

    native.filegroup(
        name = name + "/enttest_srcs",
        srcs = [":" + name + "_full_srcs"],
        output_group = "enttest",
    )
    go_library(
        name = name + "/enttest",
        srcs = [":" + name + "/enttest_srcs"],
        importpath = importpath + "/enttest",
        deps = [
            ":" + name + "/migrate",
            ":" + name + "/runtime",
            ":" + name,
            "@io_entgo_ent//dialect/sql/schema",
        ],
    )

    _ent_gazelle(
        name = name + "_gazelle",
        entities = entities,
        importpath = importpath,
        target = "//{package_name}:{name}".format(package_name = native.package_name(), name = name),
    )
