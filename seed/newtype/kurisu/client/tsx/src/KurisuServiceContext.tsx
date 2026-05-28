import * as Protobuf from "@bufbuild/protobuf"
import { createClient } from "@connectrpc/connect"
import type { Client as ConnectClient } from "@connectrpc/connect"
import { createGrpcWebTransport } from "@connectrpc/connect-web"
import React, { useCallback, useEffect, useMemo, useState } from "react"

import {
  type CreateSiliconJwtRequest,
  CreateSiliconJwtRequestSchema,
  KurisuService,
  type SiliconJwt,
} from "../../../proto/kurisu_pb"

interface KurisuServiceInterface {
  CreateSiliconJwt: (personId: string) => Promise<SiliconJwt>
}

export const KurisuServiceContext =
  React.createContext<KurisuServiceInterface | null>(null)

export const KurisuServiceProvider: React.FC<{
  children?: React.ReactNode
}> = ({ children }) => {
  const [clientGrpcWeb, setClientGrpcWeb] = useState<ConnectClient<
    typeof KurisuService
  > | null>(null)

  useEffect(() => {
    const baseUrl = "/"
    const grpcWebTransport = createGrpcWebTransport({
      baseUrl,
      useBinaryFormat: true,
      fetch: (input, init) => fetch(input, { ...init, credentials: "include" }),
    })
    const client = createClient(KurisuService, grpcWebTransport)
    setClientGrpcWeb(client)
  }, [])

  const CreateSiliconJwt = useCallback(
    async (personId: string): Promise<SiliconJwt> => {
      if (!clientGrpcWeb) {
        throw new Error("Kurisu service not initialized")
      }
      const requestPb: CreateSiliconJwtRequest = Protobuf.create(
        CreateSiliconJwtRequestSchema,
        { personId },
      )
      return await clientGrpcWeb.createSiliconJwt(requestPb)
    },
    [clientGrpcWeb],
  )

  const serviceInterface = useMemo<KurisuServiceInterface | null>(() => {
    if (!clientGrpcWeb) {
      return null
    }
    return {
      CreateSiliconJwt,
    }
  }, [clientGrpcWeb, CreateSiliconJwt])

  return (
    <KurisuServiceContext.Provider value={serviceInterface}>
      {children}
    </KurisuServiceContext.Provider>
  )
}

export const useKurisuService = () => {
  return React.useContext(KurisuServiceContext)
}

export default KurisuServiceContext
