# normalize-xattrs

Rewrites namespaced **v3** file-capability xattrs to the portable **v2** form
inside a `podman save` (docker-archive) tar, in place. The capability is kept;
only its encoding changes.

## Why

`build_freedom_image` builds with `podman build --userns auto` so a package's
postinst can chown to high UIDs. Under that user namespace the kernel stores a
file capability (e.g. `cap_net_raw` on `/usr/bin/ping`, set by `iputils-ping`)
as a `VFS_CAP_REVISION_3` xattr whose `rootid` is that build's mapped root.

A consumer that loads the image under a *different*, single-uid user namespace
— e.g. a nested rootless `podman load` — cannot map that `rootid`, so applying
the xattr fails the whole load with:

```
lsetxattr /usr/bin/ping: invalid argument
```

The `VFS_CAP_REVISION_2` form drops the `rootid` (implicitly 0), so the kernel
stamps it with the consumer's own namespace root on write. The capability is
preserved and the image loads in any namespace.

## Usage

```
normalize-xattrs <archive.tar>
```

Operates in place. `build_freedom_image` (see
`//seed/devprod/container/starlark:container.bzl`) runs it on the saved tar when
called with `normalize_xattrs = True`, so no manual step is needed; the tool is
skipped entirely otherwise.

It is a no-op when no namespaced capability is present (e.g. the debian freedom
image, whose base ships no `setcap`, so `ping` never gets a capability).

## How it works

The capability lives in one layer, so converting it changes that layer's tar
bytes and therefore its `diff_id`. To keep the archive internally consistent the
tool:

1. rewrites every v3 cap to v2 in each layer referenced by `manifest.json`;
2. renames the changed layer blob to its new content digest;
3. rewrites the `diff_id` in the image config (which renames the config blob);
4. points `manifest.json` at the new blob names.

Blob digests are recomputed by hashing content rather than trusting file names,
so the layout is handled the same whether blobs are named `<digest>.tar` or
`blobs/sha256/<digest>`.
