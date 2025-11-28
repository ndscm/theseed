import os
import shutil
import socket
import sys


def check_sandbox() -> bool:
    "Check if we are running in bazel linux sandbox"
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


def unlink(src: str, sandbox: bool = True) -> str:
    "Resolve bazel wrapped symlink"
    if sandbox:
        if not os.path.islink(src):
            raise RuntimeError(
                f"{src} is not a symlink, are you running in bazel linux sandbox?"
            )
        return os.path.join(os.path.dirname(src), os.readlink(src))
    return src


def copy_tree(src: str, dst: str, sandbox: bool = True):
    for basename in os.listdir(src):
        s = os.path.join(src, basename)
        d = os.path.join(dst, basename)
        if os.path.isdir(s):
            os.makedirs(d, exist_ok=True)
            copy_tree(s, d, sandbox=sandbox)
        else:
            shutil.copy(unlink(s, sandbox=sandbox), d, follow_symlinks=False)


def copy_srcs(srcs: list[str], target_folder: str, strip_components: int = -1):
    sandbox = check_sandbox()
    for src in srcs:
        dst = src if strip_components >= 0 else ""
        dst = dst.split(os.sep, strip_components)[-1]
        real_dst = os.path.join(target_folder, dst)
        os.makedirs(os.path.dirname(real_dst), exist_ok=True)
        if os.path.isdir(src):
            copy_tree(src, real_dst, sandbox=sandbox)
            continue
        shutil.copy(unlink(src, sandbox=sandbox), real_dst, follow_symlinks=False)
