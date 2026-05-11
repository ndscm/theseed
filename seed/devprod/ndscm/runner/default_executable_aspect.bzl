"""Aspect that extracts the default executable from a target's DefaultInfo."""

def _default_executable_aspect_impl(target, _ctx):
    if DefaultInfo in target and target[DefaultInfo].files_to_run and target[DefaultInfo].files_to_run.executable:
        executable = target[DefaultInfo].files_to_run.executable
        return [OutputGroupInfo(default_executable = depset([executable]))]
    return []

default_executable_aspect = aspect(
    implementation = _default_executable_aspect_impl,
)
