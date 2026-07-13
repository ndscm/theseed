import * as Protobuf from "@bufbuild/protobuf"
import { createClient } from "@connectrpc/connect"
import type { Client as ConnectClient } from "@connectrpc/connect"
import { createGrpcWebTransport } from "@connectrpc/connect-web"
import React, { useCallback, useEffect, useMemo, useState } from "react"

import { type FileStat } from "../../../../../../infra/filesystem/proto/simplefs_pb"
import {
  CreateDirectoryRequestSchema,
  DeleteRequestSchema,
  GetUserHomeRequestSchema,
  HooinRaidService,
  ReadDirectoryRequestSchema,
  ReadFileRequestSchema,
  RenameRequestSchema,
  StatRequestSchema,
  WriteFileRequestSchema,
} from "../../../proto/raid_pb"

export interface HooinRaidServiceInterface {
  /**
   * Reports the home directory on a person's workstation — where an editor
   * opening that workstation starts.
   *
   * It is asked for rather than assumed: where a person's home is is the
   * workstation's business, and a caller that guessed it would be guessing at
   * the shape of a container it cannot see.
   */
  GetUserHome: (personId: string) => Promise<string>

  /**
   * The calls below reach the file system of a person's workstation, one whole
   * path at a time. Each names the person and an absolute path on their
   * workstation, and nothing is held between two of them: a file is opened,
   * read, and let go by naming it.
   *
   * They fail the way the workstation's own file system does — a path that is
   * not there is `Code.NotFound`, one the person may not touch is
   * `Code.PermissionDenied` — because that is what hooin carries back. A caller
   * driving an editor with these must render those refusals as the editor's own,
   * or it will show "something went wrong" where it means "no such file".
   */
  Stat: (personId: string, path: string) => Promise<FileStat>

  ReadDirectory: (
    personId: string,
    path: string,
  ) => Promise<Record<string, number>>

  CreateDirectory: (personId: string, path: string) => Promise<void>

  ReadFile: (personId: string, path: string) => Promise<Uint8Array>

  WriteFile: (
    personId: string,
    path: string,
    content: Uint8Array,
    options: { create: boolean; overwrite: boolean },
  ) => Promise<void>

  Delete: (
    personId: string,
    path: string,
    options: { recursive: boolean },
  ) => Promise<void>

  Rename: (
    personId: string,
    sourcePath: string,
    destinationPath: string,
    options: { overwrite: boolean },
  ) => Promise<void>
}

export const HooinRaidServiceContext =
  React.createContext<HooinRaidServiceInterface | null>(null)

export const HooinRaidServiceProvider: React.FC<{
  children?: React.ReactNode
}> = ({ children }) => {
  const [clientGrpcWeb, setClientGrpcWeb] = useState<ConnectClient<
    typeof HooinRaidService
  > | null>(null)

  useEffect(() => {
    const baseUrl = "/"
    const grpcWebTransport = createGrpcWebTransport({
      baseUrl,
      useBinaryFormat: true,
      fetch: (input, init) => fetch(input, { ...init, credentials: "include" }),
    })
    const client = createClient(HooinRaidService, grpcWebTransport)
    setClientGrpcWeb(client)
  }, [])

  const GetUserHome = useCallback(
    async (personId: string): Promise<string> => {
      if (!clientGrpcWeb) {
        throw new Error("Hooin Raid service not initialized")
      }
      const response = await clientGrpcWeb.getUserHome(
        Protobuf.create(GetUserHomeRequestSchema, { personId }),
      )
      return response.path
    },
    [clientGrpcWeb],
  )

  const Stat = useCallback(
    async (personId: string, path: string): Promise<FileStat> => {
      if (!clientGrpcWeb) {
        throw new Error("Hooin Raid service not initialized")
      }
      return clientGrpcWeb.stat(
        Protobuf.create(StatRequestSchema, {
          personId,
          request: { path },
        }),
      )
    },
    [clientGrpcWeb],
  )

  const ReadDirectory = useCallback(
    async (personId: string, path: string): Promise<Record<string, number>> => {
      if (!clientGrpcWeb) {
        throw new Error("Hooin Raid service not initialized")
      }
      const response = await clientGrpcWeb.readDirectory(
        Protobuf.create(ReadDirectoryRequestSchema, {
          personId,
          request: { path },
        }),
      )
      return response.entries
    },
    [clientGrpcWeb],
  )

  const CreateDirectory = useCallback(
    async (personId: string, path: string): Promise<void> => {
      if (!clientGrpcWeb) {
        throw new Error("Hooin Raid service not initialized")
      }
      await clientGrpcWeb.createDirectory(
        Protobuf.create(CreateDirectoryRequestSchema, {
          personId,
          request: { path },
        }),
      )
    },
    [clientGrpcWeb],
  )

  const ReadFile = useCallback(
    async (personId: string, path: string): Promise<Uint8Array> => {
      if (!clientGrpcWeb) {
        throw new Error("Hooin Raid service not initialized")
      }
      const response = await clientGrpcWeb.readFile(
        Protobuf.create(ReadFileRequestSchema, {
          personId,
          request: { path },
        }),
      )
      return response.content
    },
    [clientGrpcWeb],
  )

  const WriteFile = useCallback(
    async (
      personId: string,
      path: string,
      content: Uint8Array,
      options: { create: boolean; overwrite: boolean },
    ): Promise<void> => {
      if (!clientGrpcWeb) {
        throw new Error("Hooin Raid service not initialized")
      }
      await clientGrpcWeb.writeFile(
        Protobuf.create(WriteFileRequestSchema, {
          personId,
          request: {
            path,
            content,
            create: options.create,
            overwrite: options.overwrite,
          },
        }),
      )
    },
    [clientGrpcWeb],
  )

  const Delete = useCallback(
    async (
      personId: string,
      path: string,
      options: { recursive: boolean },
    ): Promise<void> => {
      if (!clientGrpcWeb) {
        throw new Error("Hooin Raid service not initialized")
      }
      await clientGrpcWeb.delete(
        Protobuf.create(DeleteRequestSchema, {
          personId,
          request: {
            path,
            recursive: options.recursive,
          },
        }),
      )
    },
    [clientGrpcWeb],
  )

  const Rename = useCallback(
    async (
      personId: string,
      sourcePath: string,
      destinationPath: string,
      options: { overwrite: boolean },
    ): Promise<void> => {
      if (!clientGrpcWeb) {
        throw new Error("Hooin Raid service not initialized")
      }
      await clientGrpcWeb.rename(
        Protobuf.create(RenameRequestSchema, {
          personId,
          request: {
            sourcePath,
            destinationPath,
            overwrite: options.overwrite,
          },
        }),
      )
    },
    [clientGrpcWeb],
  )

  const serviceInterface = useMemo<HooinRaidServiceInterface | null>(() => {
    if (!clientGrpcWeb) {
      return null
    }
    return {
      GetUserHome,
      Stat,
      ReadDirectory,
      CreateDirectory,
      ReadFile,
      WriteFile,
      Delete,
      Rename,
    }
  }, [
    clientGrpcWeb,
    GetUserHome,
    Stat,
    ReadDirectory,
    CreateDirectory,
    ReadFile,
    WriteFile,
    Delete,
    Rename,
  ])

  return (
    <HooinRaidServiceContext.Provider value={serviceInterface}>
      {children}
    </HooinRaidServiceContext.Provider>
  )
}

export const useHooinRaidService = () => {
  return React.useContext(HooinRaidServiceContext)
}

export default HooinRaidServiceContext
