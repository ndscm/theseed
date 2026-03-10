import * as Protobuf from "@bufbuild/protobuf"
import { createClient } from "@connectrpc/connect"
import type { Client as ConnectClient } from "@connectrpc/connect"
import { createGrpcWebTransport } from "@connectrpc/connect-web"
import React, { useEffect, useState } from "react"

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
} from "../../../../proto/golink_pb"

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

const wrapInterface = (
  grpcWeb: ConnectClient<typeof GolinkService>,
): GolinkServiceInterface => {
  return {
    CreateLink: async (link: Link) => {
      const requestPb: CreateLinkRequest = Protobuf.create(
        CreateLinkRequestSchema,
        { link },
      )
      const replyPb = await grpcWeb.createLink(requestPb)
      return replyPb
    },
    GetLink: async (key: string) => {
      const requestPb: GetLinkRequest = Protobuf.create(GetLinkRequestSchema, {
        key,
      })
      const replyPb = await grpcWeb.getLink(requestPb)
      return replyPb
    },
    UpdateLink: async (link: Link, updatePaths?: string[]) => {
      const requestPb: UpdateLinkRequest = Protobuf.create(
        UpdateLinkRequestSchema,
        {
          link,
          updateMask: { paths: updatePaths },
          etag: link.etag,
        },
      )
      const replyPb = await grpcWeb.updateLink(requestPb)
      return replyPb
    },
    DeleteLink: async (key: string, options?: { etag?: string }) => {
      const requestPb: DeleteLinkRequest = Protobuf.create(
        DeleteLinkRequestSchema,
        {
          key,
          etag: options?.etag || "",
        },
      )
      await grpcWeb.deleteLink(requestPb)
    },
    ListLinks: async (options?: { pageSize?: number; pageToken?: string }) => {
      const requestPb: ListLinksRequest = Protobuf.create(
        ListLinksRequestSchema,
        {
          pageSize: options?.pageSize || 0,
          pageToken: options?.pageToken || "",
        },
      )
      const replyPb: ListLinksResponse = await grpcWeb.listLinks(requestPb)
      return replyPb
    },
  }
}

export const GolinkServiceProvider: React.FC<{
  children?: React.ReactNode
}> = ({ children }) => {
  const [serviceInterface, setServiceInterface] =
    useState<GolinkServiceInterface | null>(null)
  useEffect(() => {
    const baseUrl = "/"
    console.info("Golink Service Base Url:", baseUrl)
    const grpcWebTransport = createGrpcWebTransport({
      baseUrl,
      useBinaryFormat: true,
      fetch: (input, init) => fetch(input, { ...init, credentials: "include" }),
    })
    const grpcWebClient = createClient(GolinkService, grpcWebTransport)
    const wrapped = wrapInterface(grpcWebClient)
    setServiceInterface(wrapped)
  }, [])

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
