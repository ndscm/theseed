import React, { useEffect, useRef } from "react"

import {
  type WebFileSystem,
  serveWebFileSystem,
  webFileSystemScheme,
} from "../../ts/web-file-system"

/**
 * The VS Code workbench, drawn in an iframe on this page, open on a file system
 * the page answers for.
 *
 * The editor lives exactly as long as this is mounted: an iframe is created when
 * this appears and removed when it goes, taking the workbench, its worker, and
 * its file system with it. The iframe is why — VS Code allows one workbench per
 * page and disposing it does not give the page another, so drawn in this page the
 * editor could be opened once and never again without a reload the embedding app
 * cannot do. An iframe is a page of its own: a fresh one is a fresh workbench, and
 * mounting a second one here costs an iframe, not the app.
 *
 * The file system is answered here, on this page, over a channel the iframe's
 * worker shares with it — the two are of one origin, the iframe being served
 * beside this page. So nothing of the file system crosses the iframe as itself:
 * it stays here, and the editor reaches it the same way it would from the page.
 */
export const VscodeWebWorkbench: React.FC<{
  className?: string

  vscodePrefix?: string

  /**
   * The file system the editor opens, and the whole of what it will ever know
   * about the files it shows. It is served to the extension host — a worker in
   * the iframe, which can reach nothing else on this page — for as long as this
   * is mounted.
   */
  webFileSystem: WebFileSystem

  /**
   * The folder the editor opens. `workspaceAuthority` is whose file system it is
   * on, and rides along on every question the editor asks about a path;
   * `workspacePath` is the folder itself.
   *
   * Changing either opens a different workspace, which is a different workbench.
   * The iframe is replaced with a fresh one on the other folder; nothing is
   * preserved across that but what the file system holds.
   */
  workspaceAuthority: string

  workspacePath: string

  /**
   * Told when the workbench could not be opened at all — its files are not where
   * they are supposed to be, typically. There is nothing here to retry: the
   * caller shows what went wrong, and the next visit starts from nothing.
   */
  onError?: (error: unknown) => void
}> = ({
  webFileSystem,
  workspaceAuthority,
  workspacePath,
  className,
  vscodePrefix,
  onError,
}) => {
  vscodePrefix = (vscodePrefix ?? "/vscode").replace(/\/+$/, "")
  const refOfContainer = useRef<HTMLDivElement>(null)

  // The callback the iframe may have to reach for later is held rather than
  // closed over: a caller that passes a new function every render must not
  // thereby tear the iframe down and open the editor again.
  const refOfOnError = useRef(onError)
  refOfOnError.current = onError

  // The editor's file system is answered here, on the page. The extension that
  // registers it with VS Code runs in the extension host, which is a worker in
  // the iframe and can reach none of this — so what it asks arrives on a channel
  // it shares with this page across the iframe, and this is what answers on it.
  //
  // It is serving before the workbench is created and until after it is gone: the
  // two are mounted and unmounted together, and an editor on its way out has no
  // questions left to ask.
  useEffect(() => {
    const served = serveWebFileSystem(webFileSystem)
    return () => {
      served.dispose()
    }
  }, [webFileSystem])

  useEffect(() => {
    const container = refOfContainer.current
    if (!container) {
      return
    }

    const iframe = document.createElement("iframe")
    iframe.src = `${vscodePrefix}/vscode-web/embed.html`
    iframe.style.height = "100%"
    iframe.style.width = "100%"
    iframe.style.border = "none"
    iframe.allow = "clipboard-read; clipboard-write"

    // Both halves post same-origin — the iframe is served from this page's own
    // origin — and each acts only on a message from the other: this on messages
    // from the iframe it made, the iframe on messages from its parent.
    const onMessage = (event: MessageEvent) => {
      if (
        event.origin !== window.location.origin ||
        event.source !== iframe.contentWindow
      ) {
        return
      }

      const message = event.data
      if (!message || typeof message.type !== "string") {
        return
      }

      // The bootstrap has loaded and is waiting to be told what to open. What it
      // opens is decomposed into what survives being cloned onto a message — a
      // URI does not — and put back together on the far side.
      if (message.type === "ready") {
        iframe.contentWindow?.postMessage(
          {
            type: "create",
            additionalExtensions: [
              `${vscodePrefix}/vscode-extension/web-file-system`,
            ],
            workspaceScheme: webFileSystemScheme,
            workspaceAuthority,
            workspacePath,
          },
          window.location.origin,
        )
        return
      }

      // The workbench never opened — its files are not where they are supposed to
      // be, typically. The iframe is left where it is; the caller decides what
      // becomes of it, and unmounting is what removes it.
      if (message.type === "error") {
        refOfOnError.current?.(message.message)
      }
    }

    window.addEventListener("message", onMessage)
    container.appendChild(iframe)

    // The editor is removed with the iframe when it leaves the React tree, and is
    // not kept running behind a page no longer showing it: the iframe holds a
    // worker, an extension host, and a file system it would go on asking about.
    return () => {
      window.removeEventListener("message", onMessage)
      iframe.remove()
    }
  }, [workspaceAuthority, workspacePath, vscodePrefix])

  return <div ref={refOfContainer} className={className} />
}

export default VscodeWebWorkbench
