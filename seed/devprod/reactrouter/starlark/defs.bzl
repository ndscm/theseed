"""Bazel macros for building React Router webapps with i18n support.

Produces a per-language static build of a React Router application and merges
them into a single output directory, with the default language at the root and
each extra language under its own subdirectory.
"""

load("//seed/devprod/starlark/archive:defs.bzl", "merge_tar")
load("//seed/devprod/starlark/jsrun:defs.bzl", "js_run_binary_tar")

def react_router_webapp(
        name = "webapp",
        srcs = [],
        out = "",
        node_modules = ":node_modules",
        package_json = "package.json",
        tsconfig_json = "tsconfig.json",
        vite_config = "vite.config.ts",
        react_router_binary = ":react-router",
        react_router_config = "react-router.config.ts",
        react_router_build_args = ["build"],
        deps = [],
        default_language = "en",
        extra_languages = ["es"],
        **kwargs):
    """Builds a React Router webapp with per-language static output.

    Each language gets its own build pass with ``BUILD_LANGUAGE`` and
    ``DEFAULT_LANGUAGE`` set in the environment. The default language is
    placed at the root of the merged output directory, while extra languages
    are nested under ``/<lang>/``.

    Args:
        name: Base name for the generated targets.
        srcs: Application source files (components, styles, locales, etc.).
        out: Filename for the resulting tar archive containing the merged output.
        node_modules: Label for the npm dependencies target.
        package_json: Path to package.json.
        tsconfig_json: Path to tsconfig.json.
        vite_config: Path to the Vite configuration file.
        react_router_binary: Label for the react-router js_binary.
        react_router_config: Path to the React Router configuration file.
        react_router_build_args: Arguments passed to the react-router CLI.
        deps: Additional build-time dependencies.
        default_language: Language code for the primary build, served at ``/``.
        extra_languages: Language codes for additional builds, each served at
            ``/<lang>/``.
        **kwargs: Forwarded to the underlying build rules (e.g. ``tags``,
            ``visibility``).
    """
    js_run_binary_tar(
        name = name + "_" + default_language,
        srcs = [
            package_json,
            react_router_config,
            tsconfig_json,
            vite_config,
            node_modules,
        ] + srcs + deps,
        args = react_router_build_args,
        chdir = native.package_name(),
        env = {
            "BUILD_LANGUAGE": default_language,
            "DEFAULT_LANGUAGE": default_language,
        },
        mnemonic = "ReactRouter",
        out_dir = "dist/client",
        out_tar = name + "_" + default_language + ".tar",
        progress_message = "Compile %{label}",
        tool = react_router_binary,
        **kwargs
    )
    for lang in extra_languages:
        js_run_binary_tar(
            name = name + "_" + lang,
            srcs = [
                package_json,
                react_router_config,
                tsconfig_json,
                vite_config,
                node_modules,
            ] + srcs + deps,
            args = react_router_build_args,
            chdir = native.package_name(),
            env = {
                "BUILD_LANGUAGE": lang,
                "DEFAULT_LANGUAGE": default_language,
            },
            mnemonic = "ReactRouter",
            out_dir = "dist/" + lang + "/client",
            out_tar = name + "_" + lang + ".tar",
            progress_message = "Compile %{label}",
            tool = react_router_binary,
            **kwargs
        )
    merge_tar(
        name = name,
        srcs = {":" + name + "_" + default_language: "."} |
               {":" + name + "_" + lang: lang for lang in extra_languages},
        out = out or name + ".tar",
    )
