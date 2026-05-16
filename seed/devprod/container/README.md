# Container

This directory contains container support for theseed.

## Usage

Each container build in theseed lives in a directory named `container` and
includes a `Containerfile`.

Set the `CONTAINER_ENGINE` environment variable to `docker` or `podman` to use
the container tools in theseed.

## Legacy Usage

You may encounter the following patterns in older parts of theseed. They should
not be used in new projects.

`Dockerfile` was used when theseed only supported `docker` as the container
engine.

`CONTAINER_CLI` was used when theseed only supported container commands
compatible with all container engines.
