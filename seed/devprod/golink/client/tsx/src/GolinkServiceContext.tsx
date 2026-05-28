import * as Protobuf from "@bufbuild/protobuf"
import { createClient } from "@connectrpc/connect"
import type { Client as ConnectClient } from "@connectrpc/connect"
import { createGrpcWebTransport } from "@connectrpc/connect-web"
import React, { useCallback, useEffect, useMemo, useState } from "react"

import {
  type CreateLinkRequest,
  CreateLinkRequestSchema,
  type DeleteLinkRequest,
  DeleteLinkRequestSchema,
  type GetLinkRequest,
  GetLinkRequestSchema,
  GolinkService,
  type Link,
  type ListLinksRequest,
  ListLinksRequestSchema,
  type ListLinksResponse,
  type UpdateLinkRequest,
  UpdateLinkRequestSchema,
} from "../../../proto/golink_pb"

interface GolinkServiceInterface {
  CreateLink: (link: Link) => Promise<Link>
  GetLink: (key: string) => Promise<Link>
  UpdateLink: (link: Link, updatePaths?: string[]) => Promise<Link>
  DeleteLink: (key: string, options?: { etag?: string }) => Promise<void>
  ListLinks: (options?: {
    pageSize?: number
    pageToken?: string
  }) => Promise<ListLinksResponse>
}

export const GolinkServiceContext =
  React.createContext<GolinkServiceInterface | null>(null)

export const GolinkServiceProvider: React.FC<{
  children?: React.ReactNode
}> = ({ children }) => {
  const [clientGrpcWeb, setClientGrpcWeb] = useState<ConnectClient<
    typeof GolinkService
  > | null>(null)

  useEffect(() => {
    const baseUrl = "/"
    console.info("Golink Service Base Url:", baseUrl)
    const grpcWebTransport = createGrpcWebTransport({
      baseUrl,
      useBinaryFormat: true,
      fetch: (input, init) => fetch(input, { ...init, credentials: "include" }),
    })
    const client = createClient(GolinkService, grpcWebTransport)
    setClientGrpcWeb(client)
  }, [])

  const CreateLink = useCallback(
    async (link: Link): Promise<Link> => {
      if (!clientGrpcWeb) {
        throw new Error("Golink service not initialized")
      }
      const requestPb: CreateLinkRequest = Protobuf.create(
        CreateLinkRequestSchema,
        { link },
      )
      return await clientGrpcWeb.createLink(requestPb)
    },
    [clientGrpcWeb],
  )

  const GetLink = useCallback(
    async (key: string): Promise<Link> => {
      if (!clientGrpcWeb) {
        throw new Error("Golink service not initialized")
      }
      const requestPb: GetLinkRequest = Protobuf.create(GetLinkRequestSchema, {
        key,
      })
      return await clientGrpcWeb.getLink(requestPb)
    },
    [clientGrpcWeb],
  )

  const UpdateLink = useCallback(
    async (link: Link, updatePaths?: string[]): Promise<Link> => {
      if (!clientGrpcWeb) {
        throw new Error("Golink service not initialized")
      }
      const requestPb: UpdateLinkRequest = Protobuf.create(
        UpdateLinkRequestSchema,
        {
          link,
          updateMask: { paths: updatePaths },
          etag: link.etag,
        },
      )
      return await clientGrpcWeb.updateLink(requestPb)
    },
    [clientGrpcWeb],
  )

  const DeleteLink = useCallback(
    async (key: string, options?: { etag?: string }): Promise<void> => {
      if (!clientGrpcWeb) {
        throw new Error("Golink service not initialized")
      }
      const requestPb: DeleteLinkRequest = Protobuf.create(
        DeleteLinkRequestSchema,
        {
          key,
          etag: options?.etag || "",
        },
      )
      await clientGrpcWeb.deleteLink(requestPb)
    },
    [clientGrpcWeb],
  )

  const ListLinks = useCallback(
    async (options?: {
      pageSize?: number
      pageToken?: string
    }): Promise<ListLinksResponse> => {
      if (!clientGrpcWeb) {
        throw new Error("Golink service not initialized")
      }
      const requestPb: ListLinksRequest = Protobuf.create(
        ListLinksRequestSchema,
        {
          pageSize: options?.pageSize || 0,
          pageToken: options?.pageToken || "",
        },
      )
      return await clientGrpcWeb.listLinks(requestPb)
    },
    [clientGrpcWeb],
  )

  const serviceInterface = useMemo<GolinkServiceInterface | null>(() => {
    if (!clientGrpcWeb) {
      return null
    }
    return {
      CreateLink,
      GetLink,
      UpdateLink,
      DeleteLink,
      ListLinks,
    }
  }, [clientGrpcWeb, CreateLink, GetLink, UpdateLink, DeleteLink, ListLinks])

  return (
    <GolinkServiceContext.Provider value={serviceInterface}>
      {children}
    </GolinkServiceContext.Provider>
  )
}

export const useGolinkService = () => {
  return React.useContext(GolinkServiceContext)
}

export default GolinkServiceContext
