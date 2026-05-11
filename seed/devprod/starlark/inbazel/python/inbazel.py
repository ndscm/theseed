import os
import shutil


def check_sandbox() -> bool:
    """Check if we are running in a bazel linux sandbox by looking for a read-only root mount."""
    with open("/proc/self/mounts", "r") as f:
        mounts = f.read()
    for line in mounts.splitlines():
        mount = line.split()
        if len(mount) >= 4 and mount[1] == "/":
            modes = mount[3].split(",")
            if "ro" in modes:
                return True
            return False
    return False


def unlink(src: str) -> str:
    """Resolve a bazel-wrapped symlink. Raises if src is not a symlink."""
    if not os.path.islink(src):
        raise RuntimeError(
            f"expected symlink, got regular file (not in bazel sandbox?): {src}"
        )
    return os.path.join(os.path.dirname(src), os.readlink(src))


def copy_tree(src: str, dst: str, sandbox: bool = True):
    """Recursively copy a directory tree, resolving bazel symlinks when sandbox is True."""
    for basename in os.listdir(src):
        s = os.path.join(src, basename)
        d = os.path.join(dst, basename)
        if os.path.isdir(s):
            os.makedirs(d, exist_ok=True)
            copy_tree(s, d, sandbox=sandbox)
        else:
            resolved = unlink(s) if sandbox else s
            shutil.copy(resolved, d, follow_symlinks=False)


def copy_srcs(
    srcs: list[str],
    target_folder: str,
    strip_components: int = 0,
    sandbox: bool = True,
):
    """Copy source files/directories into target_folder.

    strip_components controls how much of each src path prefix is removed:
      >= 0: strip the first N path components (like tar --strip-components).
      < 0: flatten — discard the entire src path and place files directly
           under target_folder using only their basename.
    """
    for src in srcs:
        dst = src if strip_components >= 0 else ""
        dst = dst.split(os.sep, strip_components)[-1]
        real_dst = os.path.join(target_folder, dst)
        os.makedirs(os.path.dirname(real_dst), exist_ok=True)
        if os.path.isdir(src):
            copy_tree(src, real_dst, sandbox=sandbox)
            continue
        resolved = unlink(src) if sandbox else src
        shutil.copy(resolved, real_dst, follow_symlinks=False)
