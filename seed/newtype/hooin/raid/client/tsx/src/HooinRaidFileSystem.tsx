import { Code, ConnectError } from "@connectrpc/connect"

import {
  type WebFileStat,
  WebFileSystemError,
  type WebFileSystem as WebFileSystemInterface,
  type WebFileSystemUri,
} from "../../../../../../devprod/vscode/web/ts/web-file-system"
import { type HooinRaidServiceInterface } from "./HooinRaidServiceContext"

// What a path is arrives as the editor's own bitmask — the workstation says it
// in VS Code's values, and hooin carries the number untouched — so there is
// nothing here to render it into. A symbolic link to a directory arrives as a
// directory that points, which is what an editor opening it needs to know.

// toWebFileSystemError renders a refusal from the workstation as one the editor
// knows how to act on: an editor asks whether a path exists by reading it and
// being told that it is not there, and asks whether it may replace one by being
// told that it is.
//
// A failure that is not the file system's — the person is not on duty, the
// connection is gone — has no such name. It is carried across as it came, and
// shown.
const toWebFileSystemError = (error: unknown): unknown => {
  if (!(error instanceof ConnectError)) {
    return error
  }
  switch (error.code) {
    case Code.NotFound:
      return new WebFileSystemError("FileNotFound", error.rawMessage)
    case Code.AlreadyExists:
      return new WebFileSystemError("FileExists", error.rawMessage)
    case Code.PermissionDenied:
    case Code.Unauthenticated:
      return new WebFileSystemError("NoPermissions", error.rawMessage)
    case Code.FailedPrecondition:
      // The workstation refused the path for what it is rather than for where it
      // is: a directory read as a file, a file listed as a directory. Which of
      // the two is in what its file system said, and nowhere else.
      if (error.rawMessage.includes("Not a directory")) {
        return new WebFileSystemError("FileNotADirectory", error.rawMessage)
      }
      if (error.rawMessage.includes("Is a directory")) {
        return new WebFileSystemError("FileIsADirectory", error.rawMessage)
      }
      return new WebFileSystemError("Unavailable", error.rawMessage)
    default:
      return new WebFileSystemError("Unavailable", error.rawMessage)
  }
}

// HooinRaidFileSystem is a person's workstation, as the VS Code workbench asks
// about it. It is a `WebFileSystem`, so a page that embeds a workbench serves
// this and the editor opens the workstation:
//
//     serveWebFileSystem(new HooinRaidFileSystem(raidService))
//
// The editor's questions do not arrive here directly. They are asked in the
// extension host, which is a worker, and forwarded to whatever the page serves —
// which is the point of doing it that way: here, an app is itself again, and the
// raid client and the login it was made with are in reach, where in a worker
// neither would have been.
//
// A URI is `web-file-system://<person>/<path on their workstation>`: the scheme
// is the extension's, the same in every app that uses it, and the authority is
// the person whose workstation the path is on. Nothing is held between two
// calls: a path is named, reached, and let go.
export class HooinRaidFileSystem implements WebFileSystemInterface {
  private readonly raidService: HooinRaidServiceInterface
  private readonly authorities: { [authority: string]: string }

  constructor(
    raidService: HooinRaidServiceInterface,
    authorities?: { [authority: string]: string },
  ) {
    this.raidService = raidService
    this.authorities = authorities || {}
  }

  // getPersonId reads the person a URI names. Its authority is the handle the
  // workbench was opened on — the one it shows — and the id it stands for is
  // what raid is called with, because that is what a role is granted on. An
  // authority with no id known falls through as itself, so a URI that already
  // names an id reaches the workstation it names.
  private getPersonId(authority: string): string {
    return this.authorities[authority] ?? authority
  }

  async stat(uri: WebFileSystemUri): Promise<WebFileStat> {
    const stat = await this.raid(() => {
      return this.raidService.Stat(this.getPersonId(uri.authority), uri.path)
    })
    // The times and the size are 64-bit on the wire, and arrive as bigints. An
    // editor wants numbers, and no file is old enough or large enough for the
    // difference to be one.
    //
    // VS Code's `ctime` is a creation time, so it is the birth time it is given.
    // A file system that keeps no birth time reports 0, and a file the editor
    // thinks was made at the epoch is worse than one it thinks was made when it
    // was last chmod'ed — so the change time stands in where there is nothing
    // better, which is what VS Code's own disk provider ends up doing on the
    // platforms that leave it no choice.
    return {
      type: stat.fileType,
      ctime: Number(stat.creationTimestampMs || stat.changeTimestampMs),
      mtime: Number(stat.modificationTimestampMs),
      size: Number(stat.size),
    }
  }

  async readDirectory(uri: WebFileSystemUri): Promise<[string, number][]> {
    const entries = await this.raid(() => {
      return this.raidService.ReadDirectory(
        this.getPersonId(uri.authority),
        uri.path,
      )
    })
    return Object.entries(entries)
  }

  async readFile(uri: WebFileSystemUri): Promise<Uint8Array> {
    return this.raid(() => {
      return this.raidService.ReadFile(
        this.getPersonId(uri.authority),
        uri.path,
      )
    })
  }

  async writeFile(
    uri: WebFileSystemUri,
    content: Uint8Array,
    options: { create: boolean; overwrite: boolean },
  ): Promise<void> {
    await this.raid(() => {
      return this.raidService.WriteFile(
        this.getPersonId(uri.authority),
        uri.path,
        content,
        options,
      )
    })
  }

  async createDirectory(uri: WebFileSystemUri): Promise<void> {
    await this.raid(() => {
      return this.raidService.CreateDirectory(
        this.getPersonId(uri.authority),
        uri.path,
      )
    })
  }

  async delete(
    uri: WebFileSystemUri,
    options: { recursive: boolean },
  ): Promise<void> {
    await this.raid(() => {
      return this.raidService.Delete(
        this.getPersonId(uri.authority),
        uri.path,
        options,
      )
    })
  }

  async rename(
    source: WebFileSystemUri,
    destination: WebFileSystemUri,
    options: { overwrite: boolean },
  ): Promise<void> {
    // Both paths are on one workstation. Moving a file from one person's to
    // another's would be two calls and a file in between, which is not what a
    // rename is.
    if (source.authority !== destination.authority) {
      throw new WebFileSystemError(
        "NoPermissions",
        "a file cannot be renamed onto another person's workstation",
      )
    }
    await this.raid(() => {
      return this.raidService.Rename(
        this.getPersonId(source.authority),
        source.path,
        destination.path,
        options,
      )
    })
  }

  // raid makes one call and renders what it refused as something the editor can
  // act on. What the workstation's own file system said is the whole of what
  // there is to say — hooin carries its refusal back untouched — so it is read
  // rather than guessed at.
  private async raid<Value>(call: () => Promise<Value>): Promise<Value> {
    try {
      return await call()
    } catch (error) {
      throw toWebFileSystemError(error)
    }
  }
}

export default HooinRaidFileSystem
