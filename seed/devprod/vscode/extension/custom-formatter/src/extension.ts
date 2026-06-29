/**
 * Custom Formatter — runs an arbitrary shell command as a VS Code formatter.
 *
 * Two modes are supported, configured per language under `customFormatter`:
 *
 * - `run`: a standard formatter invoked on **Format Document** / **Format on
 *   Save**. The document text is piped to the command's stdin and the formatted
 *   result is read from its stdout (see {@link provideEdits}).
 * - `runAfterSave`: a command run after a manual save that formats the saved
 *   file in place, after which the buffer is reverted to reload it from disk
 *   (see {@link handleSave}).
 *
 * @module
 */
import * as child_process from "node:child_process"
import * as path from "node:path"

import * as vscode from "vscode"

/** Configuration section under which all settings live (`customFormatter.*`). */
const CONFIG_SECTION = "customFormatter"

/** Shared output channel; assigned in {@link activate}. */
let output: vscode.OutputChannel

/**
 * Logs an error to the output channel and surfaces a notification offering to
 * reveal it.
 *
 * @param message - The error detail to append to the output channel.
 */
const showError = (message: string): void => {
  output.appendLine(`Error: ${message}`)
  vscode.window
    .showErrorMessage(
      "Custom Formatter failed. See the Custom Formatter output for details.",
      "Show Output",
    )
    .then((choice) => {
      if (choice === "Show Output") {
        output.show(true)
      }
    })
}

/**
 * Returns the working directory to run a command in for the given document:
 * the workspace folder containing it, or the document's own directory when it
 * is outside any folder.
 *
 * @param document - The document being formatted.
 * @returns An absolute filesystem path to use as the command's cwd.
 */
const workspaceCwd = (document: vscode.TextDocument): string => {
  const folder = vscode.workspace.getWorkspaceFolder(document.uri)
  if (folder === undefined) {
    return path.dirname(document.uri.fsPath)
  }
  return folder.uri.fsPath
}

/**
 * Reads the list of language identifiers the formatter is registered for from
 * `customFormatter.languages`.
 *
 * @returns The configured language identifiers, or an empty array when unset.
 */
const getAllLanguages = (): string[] => {
  return (
    vscode.workspace
      .getConfiguration(CONFIG_SECTION)
      .get<string[]>("languages") ?? []
  )
}

/**
 * Runs a command through the shell, writing `input` to its stdin and collecting
 * its stdout. The command's stderr is written to the output channel.
 *
 * @param command - The shell command line to execute.
 * @param input - Text piped to the command's stdin.
 * @param cwd - Working directory for the command.
 * @param token - Cancellation token; cancelling kills the child process.
 * @returns The command's stdout.
 * @throws If the command errors or exits with a non-zero code.
 */
const runStdin = (
  command: string,
  input: string,
  cwd: string,
  token: vscode.CancellationToken,
): Promise<string> => {
  return new Promise((resolve, reject) => {
    output.appendLine(`Command: ${command}`)
    output.appendLine(`Cwd: ${cwd}`)
    const child = child_process.spawn(command, { cwd, shell: true })
    const cancel = token.onCancellationRequested(() => child.kill())

    let stdout = ""
    child.stdout.on("data", (chunk: Buffer) => {
      stdout += chunk.toString()
    })
    child.stderr.on("data", (chunk: Buffer) => {
      output.appendLine(chunk.toString())
    })
    child.on("error", (err) => {
      cancel.dispose()
      reject(err)
    })
    child.on("close", (code) => {
      cancel.dispose()
      if (code === 0) {
        resolve(stdout)
        return
      }
      reject(new Error(`formatter exited with code ${code}`))
    })

    child.stdin.end(input)
  })
}

/**
 * Runs a command through the shell, trusting only its exit code. Both stdout and
 * stderr are written to the output channel.
 *
 * @param command - The shell command line to execute.
 * @param cwd - Working directory for the command.
 * @throws If the command errors or exits with a non-zero code.
 */
const runShell = (command: string, cwd: string): Promise<void> => {
  return new Promise((resolve, reject) => {
    output.appendLine(`Command: ${command}`)
    output.appendLine(`Cwd: ${cwd}`)
    const child = child_process.spawn(command, { cwd, shell: true })

    child.stdout.on("data", (chunk: Buffer) => {
      output.appendLine(chunk.toString())
    })
    child.stderr.on("data", (chunk: Buffer) => {
      output.appendLine(chunk.toString())
    })
    child.on("error", (err) => {
      reject(err)
    })
    child.on("close", (code) => {
      if (code === 0) {
        resolve()
        return
      }
      reject(new Error(`formatter exited with code ${code}`))
    })

    child.stdin.end()
  })
}

/**
 * Runs the `run` command as a standard formatter: the document text is piped to
 * stdin and the formatted result is read from stdout. Errors are surfaced via
 * {@link showError}.
 *
 * @param document - The document to format.
 * @param token - Cancellation token forwarded to the child process.
 * @returns A single full-document replacement edit, or an empty array when the
 *   command is unset, cancelled, failed, or produced no change.
 */
const provideEdits = async (
  document: vscode.TextDocument,
  token: vscode.CancellationToken,
): Promise<vscode.TextEdit[]> => {
  const config = vscode.workspace.getConfiguration(CONFIG_SECTION, document)

  const command = config.get<string>("run") ?? ""
  if (command.length === 0) {
    return []
  }

  const text = document.getText()

  output.appendLine(`[${new Date().toISOString()}] Run ${document.uri.fsPath}`)
  output.appendLine(`Command: ${command}`)

  let formatted: string
  const start = Date.now()
  try {
    formatted = await runStdin(command, text, workspaceCwd(document), token)
  } catch (err) {
    showError(err instanceof Error ? err.message : String(err))
    return []
  }

  output.appendLine(`Done (${Date.now() - start}ms)`)

  if (token.isCancellationRequested || formatted === text) {
    return []
  }

  const fullRange = new vscode.Range(
    document.positionAt(0),
    document.positionAt(text.length),
  )
  return [vscode.TextEdit.replace(fullRange, formatted)]
}

/** Document formatting provider backed by {@link provideEdits}. */
const provider: vscode.DocumentFormattingEditProvider = {
  provideDocumentFormattingEdits(document, _options, token) {
    return provideEdits(document, token)
  },
}

/**
 * Single-quotes a file path so it is safe to interpolate as one literal
 * argument into a POSIX shell command line. Embedded single quotes are escaped
 * as the usual `'\''` sequence (close quote, escaped quote, reopen quote). This
 * keeps paths containing shell metacharacters (e.g. `$(...)`, `;`, `|`) or
 * spaces as inert data rather than executable shell syntax.
 *
 * @param filePath - The raw file path to quote.
 * @returns The path wrapped in single quotes, safe to substitute into a
 *   command run with `shell: true`.
 */
const escapeFilePath = (filePath: string): string => {
  return `'${filePath.replaceAll("'", "'\\''")}'`
}

/**
 * URIs of documents whose pending save was triggered manually (Ctrl+S / "File:
 * Save"). Populated in `onWillSaveTextDocument`, where the save reason is
 * available, and consumed in {@link handleSave} so that auto-saves (after a
 * delay or on focus change) do not run `runAfterSave`.
 */
const manualSaves = new Set<string>()

/**
 * Runs the `runAfterSave` command after a document in a registered language is
 * saved manually. The command formats the saved file in place (the `{{FILE}}`
 * placeholder is replaced with its shell-quoted path); on success the editor is
 * reverted so
 * it reloads the formatted contents from disk. Errors are surfaced via
 * {@link showError}.
 *
 * Returns early (doing nothing) for auto-saves, unregistered languages, when
 * `editor.formatOnSave` is not enabled, or when `runAfterSave` is unset.
 *
 * @param document - The document that was just saved.
 */
const handleSave = async (document: vscode.TextDocument): Promise<void> => {
  // Only run for manual saves; ignore auto-saves (after delay or on focus
  // change), whose reason was recorded in onWillSaveTextDocument.
  if (!manualSaves.delete(document.uri.toString())) {
    return
  }

  output.appendLine(`Saved document language: ${document.languageId}`)
  if (!getAllLanguages().includes(document.languageId)) {
    return
  }

  // Honor the editor's format-on-save setting (with per-language overrides), so
  // runAfterSave behaves like any other save-time formatter.
  const formatOnSave = vscode.workspace
    .getConfiguration("editor", document)
    .get<boolean>("formatOnSave")
  if (formatOnSave !== true) {
    return
  }

  const config = vscode.workspace.getConfiguration(CONFIG_SECTION, document)
  const command = config.get<string>("runAfterSave") ?? ""
  if (command.length === 0) {
    return
  }

  // Shell-quote the path: it is untrusted data (a crafted filename in an
  // untrusted repo could otherwise inject shell commands at save time) and a
  // path with spaces would split into multiple shell words.
  const resolved = command.replaceAll(
    "{{FILE}}",
    escapeFilePath(document.uri.fsPath),
  )

  output.appendLine(
    `[${new Date().toISOString()}] RunAfterSave ${document.uri.fsPath}`,
  )
  output.appendLine(`Command: ${resolved}`)

  const start = Date.now()
  try {
    await runShell(resolved, workspaceCwd(document))
  } catch (err) {
    showError(err instanceof Error ? err.message : String(err))
    return
  }

  output.appendLine(`Done (${Date.now() - start}ms)`)

  // Reload the saved buffer from disk so the in-place edits are picked up. The
  // revert command acts on the active editor, which is the editor that was just
  // saved in the common case.
  if (
    vscode.window.activeTextEditor?.document.uri.toString() ===
    document.uri.toString()
  ) {
    await vscode.commands.executeCommand("workbench.action.files.revert")
  }
}

/** Current provider registration, if any; replaced by {@link registerProvider}. */
let registration: vscode.Disposable | undefined

/**
 * (Re-)registers the formatting provider for exactly the language identifiers
 * listed in `customFormatter.languages`. Languages must be registered
 * explicitly: when the list is empty no provider is registered and the
 * extension formats nothing.
 */
const registerProvider = (): void => {
  registration?.dispose()
  registration = undefined

  const languages = getAllLanguages()
  if (languages.length === 0) {
    return
  }

  const selector = languages.map((language) => ({
    scheme: "file",
    language,
  }))
  registration = vscode.languages.registerDocumentFormattingEditProvider(
    selector,
    provider,
  )
}

/**
 * Extension entry point. Creates the output channel, registers the formatting
 * provider, and wires up listeners for configuration changes and saves. All
 * disposables are tracked on the extension context.
 *
 * @param context - The extension context provided by VS Code.
 */
const activate = (context: vscode.ExtensionContext): void => {
  output = vscode.window.createOutputChannel("Custom Formatter")
  context.subscriptions.push(output)

  context.subscriptions.push(
    new vscode.Disposable(() => {
      registration?.dispose()
      registration = undefined
    }),
  )

  registerProvider()
  context.subscriptions.push(
    vscode.workspace.onDidChangeConfiguration((event) => {
      if (event.affectsConfiguration(`${CONFIG_SECTION}.languages`)) {
        registerProvider()
      }
    }),
  )

  context.subscriptions.push(
    vscode.workspace.onWillSaveTextDocument((event) => {
      if (event.reason === vscode.TextDocumentSaveReason.Manual) {
        manualSaves.add(event.document.uri.toString())
      }
    }),
  )
  context.subscriptions.push(
    vscode.workspace.onDidSaveTextDocument((document) => {
      void handleSave(document)
    }),
  )
}

export { activate }
