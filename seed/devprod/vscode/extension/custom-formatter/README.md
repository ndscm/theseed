# VSCode Custom Formatter

Run a custom command to format any file type in VS Code. It supports two modes,
either or both of which can be configured per language:

- **`run`** — a standard formatter. The document text is piped to the command's
  stdin and the formatted result is read from its stdout. This drives **Format
  Document** (`Shift+Alt+F`) and **Format on Save** through a
  `DocumentFormattingEditProvider`.
- **`runAfterSave`** — a shell command run after the file is saved. It formats
  the saved file in place (the `{{FILE}}` placeholder is replaced with its
  path), only its exit code is trusted, and on success the editor is reverted to
  reload the result from disk. Use this for path-based, in-place tools such as
  `ndscm format` that do not read stdin or write stdout.

Both modes act only on the language identifiers listed in
`customFormatter.languages`. With no languages configured the extension does
nothing.

## Configuration

| Setting                        | Default                   | Description                                                                                       |
| ------------------------------ | ------------------------- | ------------------------------------------------------------------------------------------------- |
| `customFormatter.run`          | `""`                      | Standard stdin/stdout formatter command. Empty disables it.                                       |
| `customFormatter.runAfterSave` | `"ndscm format {{FILE}}"` | Shell command run after save. `{{FILE}}` is replaced with the saved file path. Empty disables it. |
| `customFormatter.languages`    | `[]`                      | Language identifiers the extension applies to. Empty does nothing.                                |

## How it works

`run` follows the standard formatter contract: the editor's text goes to the
command's stdin, the formatted text comes back on stdout, and the edit is
applied to the document without touching disk. Unsaved changes are preserved.

`runAfterSave` runs against the real file after VS Code has saved it. The
command formats the file in place; the extension trusts the exit code rather
than parsing output, then reverts the editor so it reloads the formatted
contents. A non-zero exit code surfaces an error notification and leaves the
editor as saved.

Both commands run with the workspace folder as their working directory, and
their stderr is written to the **Custom Formatter** output channel.

## Install

```sh
./install.sh
```
