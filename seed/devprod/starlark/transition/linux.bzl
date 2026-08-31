"""Bazel rules for building multi-platform artifacts using transitions."""

def _linux_x64_transition_impl(_settings, _attr):
    """Transition to build for Linux x64."""
    return {
        "//command_line_option:platforms": "//seed/devprod/platform:linux-x64",
    }

linux_x64_transition = transition(
    implementation = _linux_x64_transition_impl,
    inputs = [],
    outputs = ["//command_line_option:platforms"],
)
