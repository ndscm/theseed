// A VS Code file system that is somebody else's.
//
// This extension implements no file system. It registers one, and forwards every
// question the editor asks about a path to the page the workbench is embedded in:
// whatever file system that page is serving is what answers. So one extension
// serves any app — a container reached over an RPC, a bucket, a repository, a
// machine on the other side of the world. The page is where an app already has
// what it takes to answer such a question: its client, its session, its login.
//
// The question has to be forwarded because the page cannot simply be reached
// into. VS Code registers a file system only from an extension; an extension runs
// in the extension host; and the extension host is a web worker, which shares
// nothing with the page that made it. What the two do share, being of one origin,
// is a BroadcastChannel — so a question is a message, and an answer is one back.
//
// The page's half of this is `serveWebFileSystem`, in
// //seed/devprod/vscode/web/ts/web-file-system. The two halves must agree on the
// name of the channel and the shape of the messages, which is what the types
// below are. They are written out again here rather than imported: the extension
// host loads an extension as one CommonJS file with nothing but `vscode`
// resolvable around it, so anything imported would have to be bundled into it,
// and nothing bundles this.
import * as vscode from "vscode"

// webFileSystemScheme is what this file system is registered as, and so the
// scheme of every URI that reaches it: `web-file-system://<authority>/<path>`.
// What the authority means is the page's business — whose machine, which bucket,
// which repository — and nothing here reads it.
//
// It is fixed, which is what lets the manifest name it: VS Code activates this
// extension when something asks for a path on this scheme, and not before.
const webFileSystemScheme = "web-file-system"

// Which channel the page is answering on is the workbench's to say — it is the
// one that knows what it is embedding — and it says so in the
// `configurationDefaults` it passes to `create`. What it leaves unsaid is this.
const configurationSection = "webFileSystem"
const defaultChannelName = "vscode-web-file-system"

// requestTimeoutMs bounds one question. Nothing answers it but a page, and a
// page that has gone away — closed, navigated, crashed — would otherwise leave
// the editor waiting on a file forever. Failing is what lets it ask again.
const requestTimeoutMs = 30_000

interface WebFileSystemUri {
  scheme: string
  authority: string
  path: string
}

interface WebFileStat {
  type: number
  ctime: number
  mtime: number
  size: number
}

interface WebFileSystemRequest {
  id: string
  method: string
  args: unknown[]
}

interface WebFileSystemResponseError {
  code?: string
  message: string
}

interface WebFileSystemResponse {
  id: string
  value?: unknown
  error?: WebFileSystemResponseError
}

// toUri takes a URI apart to send it: VS Code's own does not survive being cloned
// onto a channel, and what is left is the whole of the question anyway — whose
// file system it is, and where on it the path lies.
const toUri = (uri: vscode.Uri): WebFileSystemUri => {
  return {
    scheme: uri.scheme,
    authority: uri.authority,
    path: uri.path,
  }
}

// toFileSystemError renders what the page refused as an error VS Code knows what
// to do with: an editor asks whether a path exists by reading it and being told
// it is not there, and asks whether it may overwrite by being told that it is.
//
// An error with no code is one the page could not name. It is shown as it came,
// which is what should happen to a failure nothing here anticipated.
const toFileSystemError = (
  uri: vscode.Uri,
  error: WebFileSystemResponseError,
): Error => {
  switch (error.code) {
    case "FileNotFound":
      return vscode.FileSystemError.FileNotFound(uri)
    case "FileExists":
      return vscode.FileSystemError.FileExists(uri)
    case "FileNotADirectory":
      return vscode.FileSystemError.FileNotADirectory(uri)
    case "FileIsADirectory":
      return vscode.FileSystemError.FileIsADirectory(uri)
    case "NoPermissions":
      return vscode.FileSystemError.NoPermissions(uri)
    case "Unavailable":
      return vscode.FileSystemError.Unavailable(uri)
    default:
      return new Error(error.message || "web file system error")
  }
}

// WebFileSystemProxy is the file system the editor sees. It holds nothing about
// the files it serves: a path is named, the question goes to the page, and the
// answer is what the editor is told.
class WebFileSystemProxy implements vscode.FileSystemProvider {
  private readonly channel: BroadcastChannel

  // pending is what has been asked and not yet answered, under the id the answer
  // will come back with. A channel is a broadcast — every page and every worker
  // of this origin hears everything on it — so an id that is not in here belongs
  // to somebody else's conversation, and is left alone.
  private readonly pending = new Map<
    string,
    (response: WebFileSystemResponse) => void
  >()

  // Nothing on the far side is watched on the editor's behalf, so nothing arrives
  // to fire this. The writes below fire it themselves: a file this editor has
  // just written is one it knows has changed, and saying so is what refreshes the
  // explorer showing it.
  private readonly emitter = new vscode.EventEmitter<vscode.FileChangeEvent[]>()
  readonly onDidChangeFile = this.emitter.event

  constructor(channelName: string) {
    this.channel = new BroadcastChannel(channelName)
    this.channel.onmessage = (event: MessageEvent) => {
      const response = event.data as WebFileSystemResponse
      if (!response || typeof response.id !== "string") {
        return
      }
      const settle = this.pending.get(response.id)
      if (!settle) {
        return
      }
      settle(response)
    }
  }

  // watch is a no-op: nothing on the far side is watched, so nothing is reported.
  // A file changed there by somebody else — an agent, a shell — is seen when the
  // editor next asks about it, and not before.
  watch(): vscode.Disposable {
    return new vscode.Disposable(() => {})
  }

  dispose(): void {
    this.channel.close()
    this.emitter.dispose()
  }

  // call asks the page one question, and waits for the answer to it.
  private async call<Value>(
    method: string,
    args: unknown[],
    uri: vscode.Uri,
  ): Promise<Value> {
    const id = crypto.randomUUID()
    const request: WebFileSystemRequest = { id, method, args }

    const response = await new Promise<WebFileSystemResponse>((resolve) => {
      const timer = setTimeout(() => {
        this.pending.delete(id)
        resolve({
          id,
          error: {
            code: "Unavailable",
            message: `${method} went unanswered for ${requestTimeoutMs}ms`,
          },
        })
      }, requestTimeoutMs)

      this.pending.set(id, (answer) => {
        clearTimeout(timer)
        this.pending.delete(id)
        resolve(answer)
      })

      this.channel.postMessage(request)
    })

    if (response.error) {
      throw toFileSystemError(uri, response.error)
    }
    return response.value as Value
  }

  async stat(uri: vscode.Uri): Promise<vscode.FileStat> {
    const stat = await this.call<WebFileStat>("stat", [toUri(uri)], uri)
    return {
      type: stat.type,
      ctime: stat.ctime,
      mtime: stat.mtime,
      size: stat.size,
    }
  }

  async readDirectory(uri: vscode.Uri): Promise<[string, vscode.FileType][]> {
    return this.call<[string, vscode.FileType][]>(
      "readDirectory",
      [toUri(uri)],
      uri,
    )
  }

  async readFile(uri: vscode.Uri): Promise<Uint8Array> {
    return this.call<Uint8Array>("readFile", [toUri(uri)], uri)
  }

  async writeFile(
    uri: vscode.Uri,
    content: Uint8Array,
    options: { create: boolean; overwrite: boolean },
  ): Promise<void> {
    await this.call<void>("writeFile", [toUri(uri), content, options], uri)
    this.emitter.fire([{ type: vscode.FileChangeType.Changed, uri }])
  }

  async createDirectory(uri: vscode.Uri): Promise<void> {
    await this.call<void>("createDirectory", [toUri(uri)], uri)
    this.emitter.fire([{ type: vscode.FileChangeType.Created, uri }])
  }

  async delete(
    uri: vscode.Uri,
    options: { recursive: boolean },
  ): Promise<void> {
    await this.call<void>("delete", [toUri(uri), options], uri)
    this.emitter.fire([{ type: vscode.FileChangeType.Deleted, uri }])
  }

  async rename(
    source: vscode.Uri,
    destination: vscode.Uri,
    options: { overwrite: boolean },
  ): Promise<void> {
    await this.call<void>(
      "rename",
      [toUri(source), toUri(destination), options],
      source,
    )
    this.emitter.fire([
      { type: vscode.FileChangeType.Deleted, uri: source },
      { type: vscode.FileChangeType.Created, uri: destination },
    ])
  }
}

export function activate(context: vscode.ExtensionContext) {
  const configuration = vscode.workspace.getConfiguration(configurationSection)
  const channelName = configuration.get<string>("channel") || defaultChannelName

  const fileSystem = new WebFileSystemProxy(channelName)
  context.subscriptions.push(fileSystem)
  context.subscriptions.push(
    vscode.workspace.registerFileSystemProvider(
      webFileSystemScheme,
      fileSystem,
      { isCaseSensitive: true },
    ),
  )
}

export function deactivate() {}
