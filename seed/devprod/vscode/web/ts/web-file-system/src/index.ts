// The page's half of the VS Code web file system.
//
// A page that embeds the workbench and wants it to open something other than
// what the browser can already reach — a container, a server, another person's
// machine — has to hand VS Code a file system. It cannot hand it one directly:
// VS Code registers a file system only from an extension, and an extension runs
// in the extension host, which is a web worker. Nothing in a worker can see the
// page's `window`.
//
// So the page hands its file system to this instead, which serves it to the
// worker: every question the editor asks arrives here as a message, is answered
// by that file system, and goes back the way it came. The extension is the same
// for every app — what differs is what the page serves.
//
//     serveWebFileSystem(new MyFileSystem())
//
// The channel is same-origin, and so is the extension host once the workbench is
// told not to run it from a CDN. Both halves must agree on the channel's name
// and on the messages below; the page sets the name through the workbench's
// `configurationDefaults`, or leaves it as it is here.

// webFileSystemScheme is the scheme the extension registers this file system
// under, and so the scheme of every URI that reaches it: a folder the workbench
// opens on it is `web-file-system://<authority>/<path>`. What the authority
// means is the page's business — whose machine, which bucket, which repository —
// and the extension neither reads it nor cares.
//
// It is fixed rather than chosen by the app. One page serves one file system, so
// there is nothing for a second scheme to name, and a fixed one is what lets the
// extension say in its manifest which file system it is the provider of — which
// is what VS Code activates it on.
export const webFileSystemScheme = "web-file-system"

// defaultWebFileSystemChannel is the BroadcastChannel both halves meet on. An
// app that embeds more than one workbench in one origin gives each its own name.
export const defaultWebFileSystemChannel = "vscode-web-file-system"

// WebFileType says what a path is. The values are VS Code's own, and are a
// bitmask: a symbolic link to a directory is `Directory | SymbolicLink`, because
// an editor opening it wants to know both that it points and what it points at.
export const WebFileType = {
  Unknown: 0,
  File: 1,
  Directory: 2,
  SymbolicLink: 64,
} as const

// WebFileSystemUri is a URI taken apart, because a URI cannot be sent as one:
// what crosses the channel is cloned, and VS Code's own URI class does not
// survive that. The authority is whose file system it is, and the path is the
// path on it.
export interface WebFileSystemUri {
  scheme: string
  authority: string
  path: string
}

export interface WebFileStat {
  // What the path is. See WebFileType.
  type: number

  // When the file was last changed and last modified, in milliseconds since the
  // epoch, and how many bytes it holds.
  ctime: number
  mtime: number
  size: number
}

// WebFileSystemErrorCode is what an editor can act on. Anything a file system
// refuses for a reason not named here is still an error — it is simply one the
// editor can only show, rather than one it knows what to do about.
export type WebFileSystemErrorCode =
  | "FileNotFound"
  | "FileExists"
  | "FileNotADirectory"
  | "FileIsADirectory"
  | "NoPermissions"
  | "Unavailable"

// WebFileSystemError is what a file system throws to say which of those it is.
export class WebFileSystemError extends Error {
  readonly code: WebFileSystemErrorCode

  constructor(code: WebFileSystemErrorCode, message: string) {
    super(message)
    this.name = "WebFileSystemError"
    this.code = code
  }
}

// WebFileSystem is what a page registers. It is VS Code's file system provider,
// with the editor's own types taken out of it: the page does not have them, and
// a plain path in and plain bytes out is all that has to cross the channel.
//
// Watching is not part of it. Nothing here is told when a file changes behind
// its back, so an editor showing one is showing what was there when it last
// asked.
export interface WebFileSystem {
  stat(uri: WebFileSystemUri): Promise<WebFileStat>

  // The entries of a directory, each with its name and what it is.
  readDirectory(uri: WebFileSystemUri): Promise<[string, number][]>

  readFile(uri: WebFileSystemUri): Promise<Uint8Array>

  writeFile(
    uri: WebFileSystemUri,
    content: Uint8Array,
    options: { create: boolean; overwrite: boolean },
  ): Promise<void>

  createDirectory(uri: WebFileSystemUri): Promise<void>

  delete(uri: WebFileSystemUri, options: { recursive: boolean }): Promise<void>

  rename(
    source: WebFileSystemUri,
    destination: WebFileSystemUri,
    options: { overwrite: boolean },
  ): Promise<void>
}

// WebFileSystemRequest is one question, and WebFileSystemResponse the answer to
// it. They are matched by `id`: the channel is a broadcast, so both halves hear
// their own messages and everybody else's, and an id nobody is waiting on is one
// to ignore.
export interface WebFileSystemRequest {
  id: string
  method: keyof WebFileSystem
  args: unknown[]
}

export interface WebFileSystemResponseError {
  code?: WebFileSystemErrorCode
  message: string
}

export interface WebFileSystemResponse {
  id: string
  value?: unknown
  error?: WebFileSystemResponseError
}

// toResponseError renders whatever a file system threw as something the editor
// on the other side can act on. A WebFileSystemError says which kind it is;
// anything else is carried across as the message it came with, and shown.
const toResponseError = (error: unknown): WebFileSystemResponseError => {
  if (error instanceof WebFileSystemError) {
    return { code: error.code, message: error.message }
  }
  return { message: error instanceof Error ? error.message : String(error) }
}

// call answers one question by asking the page's file system. The method is
// chosen here rather than looked up on the object, so that a message naming
// something that is not a file system operation reaches nothing.
const call = async (
  fileSystem: WebFileSystem,
  request: WebFileSystemRequest,
): Promise<unknown> => {
  const args = request.args
  switch (request.method) {
    case "stat":
      return fileSystem.stat(args[0] as WebFileSystemUri)
    case "readDirectory":
      return fileSystem.readDirectory(args[0] as WebFileSystemUri)
    case "readFile":
      return fileSystem.readFile(args[0] as WebFileSystemUri)
    case "writeFile":
      return fileSystem.writeFile(
        args[0] as WebFileSystemUri,
        args[1] as Uint8Array,
        args[2] as { create: boolean; overwrite: boolean },
      )
    case "createDirectory":
      return fileSystem.createDirectory(args[0] as WebFileSystemUri)
    case "delete":
      return fileSystem.delete(
        args[0] as WebFileSystemUri,
        args[1] as { recursive: boolean },
      )
    case "rename":
      return fileSystem.rename(
        args[0] as WebFileSystemUri,
        args[1] as WebFileSystemUri,
        args[2] as { overwrite: boolean },
      )
    default:
      throw new WebFileSystemError(
        "Unavailable",
        `${request.method} is not a file system operation`,
      )
  }
}

/**
 * Serves a file system to the VS Code extension host, and keeps serving it until
 * the returned handle is disposed.
 *
 * It must be serving before the workbench is created, and for as long as the
 * workbench lives: the editor asks whenever it wants a file, and a question
 * nobody is listening for goes unanswered until it times out.
 */
export const serveWebFileSystem = (
  fileSystem: WebFileSystem,
  options?: { channel?: string },
): { dispose: () => void } => {
  const channel = new BroadcastChannel(
    options?.channel || defaultWebFileSystemChannel,
  )

  channel.onmessage = async (event: MessageEvent) => {
    const request = event.data as WebFileSystemRequest
    if (!request || typeof request.id !== "string" || !request.method) {
      // Something else is talking on this channel, or a response of ours has
      // come back around to us. Neither is a question.
      return
    }

    let response: WebFileSystemResponse
    try {
      response = { id: request.id, value: await call(fileSystem, request) }
    } catch (error) {
      response = { id: request.id, error: toResponseError(error) }
    }
    channel.postMessage(response)
  }

  return {
    dispose: () => {
      channel.close()
    },
  }
}
