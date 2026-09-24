# Platforms

Common target [platforms](https://bazel.build/concepts/platforms), named after
the platforms they describe.

## Naming

There are too many conventions for naming a platform, and bazel support all the
common aliases. We decided to go with the Go convention of `<kernel>-<arch>`,
because it's short and has no underscore.

The two halves map onto the label: the kernel is the package, the arch is the
target. So `<kernel>-<arch>` is spelled
`//seed/devprod/platform/<kernel>:<arch>`, for example
`//seed/devprod/platform/linux:amd64`.

We deviate from Go in two ways:

- We use `x86` instead of `386` for the 32-bit x86 architecture.
- `android` as kernel means the target additionally requires the Android API.
