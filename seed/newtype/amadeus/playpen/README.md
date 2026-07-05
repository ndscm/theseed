# Playpen Workstation

Playpen is an emulated workstation. It runs a podman container with systemd as
the init process, and lets the outer service open a login-shell session inside
it (e.g. `/usr/bin/zsh -i`) and hand that session to a long-lived program.

## Container lifecycle

`BootPlaypenController` starts the container detached. It runs systemd as init
under a private cgroup namespace parented at `/amadeus`, mounts the shared
playpen home, and creates the login user.

`PlaypenController.Shutdown` stops and removes the container
(`podman rm --force`).

`PlaypenController.Home` reports the login user's home directory inside the
container. Because the shared playpen home is mounted at `/home`, this is
`/home/<userHandle>`.

## Sessions

`PlaypenController.StartTty` opens a shell session inside the container as the
playpen user via `podman exec --interactive`, returning a `PlaypenTty`. The
session starts the playpen user's login shell (`zsh -i`), so `~/.zshrc` is
sourced and its PATH and environment are in effect before any workload runs.

The session is _not_ attached to a pseudo-terminal. That keeps stdout and stderr
as separate streams (a PTY would merge them), at the cost of the shell not
seeing a real terminal — `zsh -i` starts without job control or a prompt, which
is what we want for programmatic driving.

## Delegating the session

`PlaypenTty.Delegate` hands the session to a single long-lived program. It
creates the working directory (as the playpen user, inside the container), `cd`s
into it, and `exec`s the program so it _replaces_ the shell:

```go
tty, err := controller.StartTty(ctx, "/usr/bin/zsh", []string{"-i"})
err = tty.Delegate(controller.Home()+"/topic/default", "claude", claudeArgs)

go io.Copy(os.Stdout, tty.Stdout)     // stream stdout live
io.WriteString(tty.Stdin, inputLine)  // push input continuously
err = tty.Wait()                      // blocks until the program exits
```

Because `exec` replaces the shell, the program owns `Stdin`, `Stdout`, and
`Stderr` for its whole life. A caller can therefore push input to it
continuously (e.g. a `claude` subprocess in stream-json mode) and read its
output as it is produced, rather than one command at a time. `Wait` reaps the
program; `Close` ends the session instead by closing its stdin.

Since the shell is gone after `exec`, there is no prompt or completion marker
mixed into the program's output — `Stdout` carries only what the program itself
writes.
