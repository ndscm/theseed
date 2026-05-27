# Container

This directory contains container support for theseed.

## Usage

Each container build in theseed lives in a directory named `container` and
includes a `Containerfile`.

Set the `CONTAINER_ENGINE` environment variable to `docker` or `podman` to use
the container tools in theseed.

## Deployment

Deployment scripts live in a `deploy/<cloud>` directory next to the `container`
directory, where `<cloud>` is one of `gcloud`, `aws`, `azure`, `onpremise`, etc.

A general-purpose deploy script is named with a `deploy` prefix and has sensible
defaults for all parameters. A convenience script for a specific instance uses a
`quick-` prefix and typically only takes secrets as arguments. If a secret is
not provided, the script should check whether one was stored previously and
reuse it. Maintainers are responsible for logging into the instance and removing
stored secrets when they are no longer needed.

## Legacy Usage

You may encounter the following patterns in older parts of theseed. They should
not be used in new projects.

`Dockerfile` was used when theseed only supported `docker` as the container
engine.
