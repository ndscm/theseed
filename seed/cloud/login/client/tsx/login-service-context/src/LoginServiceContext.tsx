import * as Protobuf from "@bufbuild/protobuf"
import { createClient } from "@connectrpc/connect"
import type { Client as ConnectClient } from "@connectrpc/connect"
import { createGrpcWebTransport } from "@connectrpc/connect-web"
import React, { useEffect, useState } from "react"

import {
  type GetLoginStatusRequest,
  GetLoginStatusRequestSchema,
  LoginService,
  type LoginStatus,
} from "../../../../proto/login_pb"

interface LoginServiceInterface {
  GetLoginStatus: () => Promise<LoginStatus>
}

export const LoginServiceContext =
  React.createContext<LoginServiceInterface | null>(null)

const wrapInterface = (
  grpcWeb: ConnectClient<typeof LoginService>,
): LoginServiceInterface => {
  return {
    GetLoginStatus: async () => {
      const requestPb: GetLoginStatusRequest = Protobuf.create(
        GetLoginStatusRequestSchema,
        {},
      )
      const replyPb = await grpcWeb.getLoginStatus(requestPb)
      return replyPb
    },
  }
}

export const LoginServiceProvider: React.FC<{
  children?: React.ReactNode
}> = ({ children }) => {
  const [serviceInterface, setServiceInterface] =
    useState<LoginServiceInterface | null>(null)
  useEffect(() => {
    const baseUrl = "/"
    console.info("Login Service Base Url:", baseUrl)
    const grpcWebTransport = createGrpcWebTransport({
      baseUrl,
      useBinaryFormat: true,
      fetch: (input, init) => fetch(input, { ...init, credentials: "include" }),
    })
    const grpcWebClient = createClient(LoginService, grpcWebTransport)
    const wrapped = wrapInterface(grpcWebClient)
    setServiceInterface(wrapped)
  }, [])

  return (
    <LoginServiceContext.Provider value={serviceInterface}>
      {children}
    </LoginServiceContext.Provider>
  )
}

export const useLoginService = () => {
  return React.useContext(LoginServiceContext)
}

export default LoginServiceContext
