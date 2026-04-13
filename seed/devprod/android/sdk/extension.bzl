"""Android SDK module extension for Bazel.

Downloads Android SDK packages using Bazel's built-in repository cache,
parsing the official Android SDK repository XML to resolve download URLs
and integrity hashes. This replaces sdkmanager for hermetic, cached builds.
"""

# --- SHA1 hex to SRI integrity format ---

_B64 = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789+/"

_HEX = {
    "0": 0,
    "1": 1,
    "2": 2,
    "3": 3,
    "4": 4,
    "5": 5,
    "6": 6,
    "7": 7,
    "8": 8,
    "9": 9,
    "a": 10,
    "b": 11,
    "c": 12,
    "d": 13,
    "e": 14,
    "f": 15,
}

def _sha1_to_integrity(sha1_hex):
    """Convert a SHA1 hex string to an SRI integrity value (sha1-<base64>)."""
    bytes = []
    for i in range(0, len(sha1_hex), 2):
        bytes.append(_HEX[sha1_hex[i]] * 16 + _HEX[sha1_hex[i + 1]])
    b64 = []
    for i in range(0, len(bytes), 3):
        b0 = bytes[i]
        b1 = bytes[i + 1] if i + 1 < len(bytes) else 0
        b2 = bytes[i + 2] if i + 2 < len(bytes) else 0
        b64.append(_B64[(b0 >> 2) & 0x3F])
        b64.append(_B64[((b0 << 4) | (b1 >> 4)) & 0x3F])
        b64.append(_B64[((b1 << 2) | (b2 >> 6)) & 0x3F] if i + 1 < len(bytes) else "=")
        b64.append(_B64[b2 & 0x3F] if i + 2 < len(bytes) else "=")
    return "sha1-" + "".join(b64)

# --- Repository XML parsing ---

def _extract_tag_value(xml, tag):
    """Extract text content of the first <tag>...</tag> in xml."""
    open_tag = "<%s" % tag
    idx = xml.find(open_tag)
    if idx == -1:
        return None
    gt = xml.find(">", idx)
    if gt == -1:
        return None
    close = xml.find("</%s>" % tag, gt)
    if close == -1:
        return None
    return xml[gt + 1:close].strip()

def _parse_archives(pkg_xml):
    """Extract list of {url, sha1, host_os} from a <remotePackage> block."""
    archives = []
    for part in pkg_xml.split("<archive>")[1:]:
        end = part.find("</archive>")
        if end == -1:
            continue
        block = part[:end]
        url = _extract_tag_value(block, "url")
        sha1 = _extract_tag_value(block, "checksum")
        host_os = _extract_tag_value(block, "host-os")
        if url and sha1:
            archives.append({"url": url, "sha1": sha1, "host_os": host_os or ""})
    return archives

def _find_package(xml_content, package_path):
    """Find a non-obsolete remotePackage and return its archives.

    The marker '<remotePackage path="X">' (without obsolete="true" before path)
    guarantees we match the canonical entry first.
    """
    marker = "<remotePackage path=\"%s\">" % package_path
    idx = xml_content.find(marker)
    if idx == -1:
        return None
    end = xml_content.find("</remotePackage>", idx)
    if end == -1:
        return None
    return _parse_archives(xml_content[idx:end])

def _parse_packages(xml_content, package_paths):
    """Parse repository XML for all requested packages.

    Returns dict mapping package_path -> list of archive dicts.
    """
    result = {}
    for pkg in package_paths:
        archives = _find_package(xml_content, pkg)
        if not archives:
            fail("Package '%s' not found in repository XML" % pkg)
        result[pkg] = archives
    return result

# --- Repository rule ---

def _host_os(repository_ctx):
    """Map repository_ctx.os.name to Android SDK archive os values."""
    name = repository_ctx.os.name.lower()
    if "linux" in name:
        return "linux"
    if "mac" in name:
        return "macosx"
    fail("Unsupported host OS for Android SDK: %s" % repository_ctx.os.name)

def _android_sdk_impl(repository_ctx):
    host = _host_os(repository_ctx)
    packages = json.decode(repository_ctx.attr.packages_json)
    repo_url = repository_ctx.attr.repository_url

    build_lines = [
        "package(default_visibility = [\"//visibility:public\"])",
        "",
    ]

    all_targets = []
    for pkg_path in repository_ctx.attr.package_paths:
        archives = packages[pkg_path]

        # Pick archive matching host OS, or platform-independent archive
        archive = None
        for a in archives:
            if a["host_os"] == host or a["host_os"] == "":
                archive = a
                break
        if not archive:
            fail("No archive for host OS '%s' in package '%s'" % (host, pkg_path))

        url = repo_url + archive["url"]
        integrity = _sha1_to_integrity(archive["sha1"])

        # SDK directory path: "ndk;30.0.14904198" -> "ndk/30.0.14904198"
        pkg_dir = pkg_path.replace(";", "/")

        # Extract into a temporary directory so we can flatten if needed
        tmp = pkg_dir + "__tmp"
        repository_ctx.download_and_extract(
            url = url,
            integrity = integrity,
            output = tmp,
        )

        # If the archive has a single top-level directory, strip it so the
        # package contents sit directly under pkg_dir (matching the layout
        # that sdkmanager would produce).
        ls = repository_ctx.execute(["find", tmp, "-mindepth", "1", "-maxdepth", "1"])
        entries = [e for e in ls.stdout.strip().split("\n") if e]

        if len(entries) == 1:
            probe = repository_ctx.execute(["test", "-d", entries[0]])
            if probe.return_code == 0:
                # Single directory — promote it
                repository_ctx.execute(["mv", entries[0], pkg_dir])
                repository_ctx.execute(["rm", "-rf", tmp])
            else:
                repository_ctx.execute(["mv", tmp, pkg_dir])
        else:
            repository_ctx.execute(["mv", tmp, pkg_dir])

        # Per-package filegroup
        safe = pkg_path.replace(";", "_").replace(".", "_")
        build_lines.append("filegroup(")
        build_lines.append("    name = \"%s\"," % safe)
        build_lines.append("    srcs = glob([\"%s/**\"])," % pkg_dir)
        build_lines.append(")")
        build_lines.append("")
        all_targets.append("\":%s\"" % safe)

    build_lines.append("filegroup(")
    build_lines.append("    name = \"files\",")
    build_lines.append("    srcs = [%s]," % ", ".join(all_targets))
    build_lines.append(")")
    build_lines.append("")

    # Remove unwanted files (e.g. stray BUILD files from SDK archives)
    for f in repository_ctx.attr.remove_files:
        repository_ctx.delete(f)

    repository_ctx.file("BUILD.bazel", "\n".join(build_lines))

_android_sdk = repository_rule(
    implementation = _android_sdk_impl,
    attrs = {
        "package_paths": attr.string_list(mandatory = True),
        "packages_json": attr.string(mandatory = True),
        "remove_files": attr.string_list(default = []),
        "repository_url": attr.string(mandatory = True),
    },
)

# --- Module extension ---

_configure = tag_class(
    attrs = {
        "packages": attr.string_list(mandatory = True),
        "remove_files": attr.string_list(default = []),
        "repository_file": attr.label(mandatory = True),
        "repository_url": attr.string(mandatory = True),
    },
)

def _android_sdk_extension_impl(module_ctx):
    config = None
    for mod in module_ctx.modules:
        for tag in mod.tags.configure:
            if config != None:
                fail("android_sdk_extension.configure() may only be called once")
            config = tag
    if config == None:
        fail("android_sdk_extension requires a configure() tag")

    xml_content = module_ctx.read(config.repository_file)
    packages = config.packages
    pkg_info = _parse_packages(xml_content, packages)

    _android_sdk(
        name = "android_sdk",
        package_paths = packages,
        packages_json = json.encode(pkg_info),
        remove_files = config.remove_files,
        repository_url = config.repository_url,
    )

android_sdk_extension = module_extension(
    implementation = _android_sdk_extension_impl,
    tag_classes = {
        "configure": _configure,
    },
)
