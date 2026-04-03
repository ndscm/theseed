import * as Protobuf from "@bufbuild/protobuf"
import { createClient } from "@connectrpc/connect"
import type { Client as ConnectClient } from "@connectrpc/connect"
import { createGrpcWebTransport } from "@connectrpc/connect-web"
import React, { useCallback, useEffect, useMemo, useState } from "react"

import {
  type Category,
  type CreateStuffRequest,
  CreateStuffRequestSchema,
  type DeleteStuffRequest,
  DeleteStuffRequestSchema,
  type GetStuffRequest,
  GetStuffRequestSchema,
  type ListCategoriesRequest,
  ListCategoriesRequestSchema,
  type ListCategoriesResponse,
  type ListStuffRequest,
  ListStuffRequestSchema,
  type ListStuffResponse,
  type Stuff,
  StuffService,
  type UpdateStuffRequest,
  UpdateStuffRequestSchema,
} from "../../../../proto/stuff_pb"

interface StuffServiceInterface {
  ListCategories: () => Promise<ListCategoriesResponse>
  CreateStuff: (stuff: Stuff) => Promise<Stuff>
  GetStuff: (uuid: string) => Promise<Stuff>
  UpdateStuff: (stuff: Stuff) => Promise<Stuff>
  DeleteStuff: (uuid: string) => Promise<void>
  ListStuff: (options?: { filter?: string }) => Promise<ListStuffResponse>
}

export const StuffServiceContext =
  React.createContext<StuffServiceInterface | null>(null)

export const StuffServiceProvider: React.FC<{
  children?: React.ReactNode
}> = ({ children }) => {
  const [clientGrpcWeb, setClientGrpcWeb] = useState<ConnectClient<
    typeof StuffService
  > | null>(null)

  useEffect(() => {
    const baseUrl = "/"
    console.info("Stuff Service Base Url:", baseUrl)
    const grpcWebTransport = createGrpcWebTransport({
      baseUrl,
      useBinaryFormat: true,
      fetch: (input, init) => fetch(input, { ...init, credentials: "include" }),
    })
    const client = createClient(StuffService, grpcWebTransport)
    setClientGrpcWeb(client)
  }, [])

  const ListCategories =
    useCallback(async (): Promise<ListCategoriesResponse> => {
      if (!clientGrpcWeb) {
        throw new Error("Stuff service not initialized")
      }
      const requestPb: ListCategoriesRequest = Protobuf.create(
        ListCategoriesRequestSchema,
        {},
      )
      return await clientGrpcWeb.listCategories(requestPb)
    }, [clientGrpcWeb])

  const CreateStuff = useCallback(
    async (stuff: Stuff): Promise<Stuff> => {
      if (!clientGrpcWeb) {
        throw new Error("Stuff service not initialized")
      }
      const requestPb: CreateStuffRequest = Protobuf.create(
        CreateStuffRequestSchema,
        { stuff },
      )
      return await clientGrpcWeb.createStuff(requestPb)
    },
    [clientGrpcWeb],
  )

  const GetStuff = useCallback(
    async (uuid: string): Promise<Stuff> => {
      if (!clientGrpcWeb) {
        throw new Error("Stuff service not initialized")
      }
      const requestPb: GetStuffRequest = Protobuf.create(
        GetStuffRequestSchema,
        { uuid },
      )
      return await clientGrpcWeb.getStuff(requestPb)
    },
    [clientGrpcWeb],
  )

  const UpdateStuff = useCallback(
    async (stuff: Stuff): Promise<Stuff> => {
      if (!clientGrpcWeb) {
        throw new Error("Stuff service not initialized")
      }
      const requestPb: UpdateStuffRequest = Protobuf.create(
        UpdateStuffRequestSchema,
        { stuff },
      )
      return await clientGrpcWeb.updateStuff(requestPb)
    },
    [clientGrpcWeb],
  )

  const DeleteStuff = useCallback(
    async (uuid: string): Promise<void> => {
      if (!clientGrpcWeb) {
        throw new Error("Stuff service not initialized")
      }
      const requestPb: DeleteStuffRequest = Protobuf.create(
        DeleteStuffRequestSchema,
        { uuid },
      )
      await clientGrpcWeb.deleteStuff(requestPb)
    },
    [clientGrpcWeb],
  )

  const ListStuff = useCallback(
    async (options?: { filter?: string }): Promise<ListStuffResponse> => {
      if (!clientGrpcWeb) {
        throw new Error("Stuff service not initialized")
      }
      const requestPb: ListStuffRequest = Protobuf.create(
        ListStuffRequestSchema,
        {
          filter: options?.filter || "",
        },
      )
      return await clientGrpcWeb.listStuff(requestPb)
    },
    [clientGrpcWeb],
  )

  const serviceInterface = useMemo<StuffServiceInterface | null>(() => {
    if (!clientGrpcWeb) {
      return null
    }
    return {
      ListCategories,
      CreateStuff,
      GetStuff,
      UpdateStuff,
      DeleteStuff,
      ListStuff,
    }
  }, [
    clientGrpcWeb,
    ListCategories,
    CreateStuff,
    GetStuff,
    UpdateStuff,
    DeleteStuff,
    ListStuff,
  ])

  return (
    <StuffServiceContext.Provider value={serviceInterface}>
      {children}
    </StuffServiceContext.Provider>
  )
}

export const useStuffService = () => {
  return React.useContext(StuffServiceContext)
}

export default StuffServiceContext
