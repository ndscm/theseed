"""
Print output groups of a target.

Usage:
    bazel build --nobuild --aspects //seed/devprod/bazel:output_groups.bzl%print_output_groups //some:target

See: https://stackoverflow.com/questions/61252620/how-to-list-the-output-groups-of-a-bazel-rule
"""

def _print_output_groups_impl(target, _):
    for output_group in target.output_groups:
        print("output group: " + output_group)
    return []

print_output_groups = aspect(
    implementation = _print_output_groups_impl,
)
