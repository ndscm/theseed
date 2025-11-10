"""Secret rules"""

def local_secret(name, local_path):
    native.genrule(
        name = name,
        outs = [name.upper()],
        cmd_bash = "cat {local_path} > $@".format(local_path = local_path),
        tags = [
            "local",
            "manual",
            "no-remote-cache",
        ],
    )
