import * as vscode from "vscode"

function activate(context: vscode.ExtensionContext) {
  const disposable = vscode.commands.registerCommand(
    "customFormatter.format",
    () => {
      vscode.window.showInformationMessage("Custom Formatter: Format")
    },
  )

  context.subscriptions.push(disposable)
}

export { activate }
