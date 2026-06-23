# ndscm

ndscm is a command line tool that helps you work better in repositories with
heavy code review and heavy CI. You develop in a single worktree and keep each
commit as small as possible. Small commits make it much easier to split,
reorder, and resolve conflicts. Once you feel confident about a feature, you can
pick the good commits for others to review—and you can keep developing on your
dev head without waiting for the review to be merged.

Here's a simplified overview of a dev branch.

```asciiflow
 |
 *
 |
 * <- origin/main pointer
 |
 * personal
 |
 * debugging
 |
 * changes
 |
 * <- base pointer
 |
 * the
 |
 * first
 |
 * feature
 |
 * <- change/first pointer
 |
 * the
 |
 * second
 |
 * feature
 |
 * <- change/second pointer
 |
 * other
 |
 * ongoing
 |
 * features
 |
 * <- dev head pointer
```

## Getting Started

Download the ndscm CLI and install it into `${HOME}/.local/bin`. Then run

```bash
ndscm setup
```

to add the shell setup statement to your `.zshrc` (or `.bashrc` if you don't use
zsh).

Restart your terminal to load the `nd` wrapper, then clone your working
repository with

```bash
nd connect yourrepo git@github.com:organization/yourrepo.git
```

## Commands

### Commit Management

- `nd connect`: download a repository
- `nd dev`: cd into the development worktree
- `nd submit`: cut the development branch and push commits to the remote for
  review

### Script Management

- `nd run`: Run selected phases
- `nd check`: Run all phases

## Caveat

Only git is supported for now.

The name `nd` is a nod to `cd`. Instead of moving between directories, `nd`
moves between branches and worktrees.
