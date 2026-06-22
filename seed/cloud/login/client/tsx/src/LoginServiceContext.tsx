import * as Protobuf from "@bufbuild/protobuf"
import { createClient } from "@connectrpc/connect"
import type { Client as ConnectClient } from "@connectrpc/connect"
import { createGrpcWebTransport } from "@connectrpc/connect-web"
import React, { useCallback, useEffect, useMemo, useState } from "react"

import {
  type GetLoginStatusRequest,
  GetLoginStatusRequestSchema,
  LoginService,
  type LoginStatus,
} from "../../../proto/login_pb"

export type { LoginStatus }

interface LoginServiceInterface {
  current: LoginStatus | undefined
  loading: boolean
  GetLoginStatus: () => Promise<LoginStatus>
  reload: () => Promise<void>
  login: () => Promise<void>
}

export const LoginServiceContext =
  React.createContext<LoginServiceInterface | null>(null)

export const LoginServiceProvider: React.FC<{
  children?: React.ReactNode
}> = ({ children }) => {
  const [current, setCurrent] = useState<LoginStatus | undefined>(undefined)
  const [loading, setLoading] = useState(true)
  const [clientGrpcWeb, setClientGrpcWeb] = useState<ConnectClient<
    typeof LoginService
  > | null>(null)

  useEffect(() => {
    const baseUrl = "/"
    console.info("Login Service Base Url:", baseUrl)
    const grpcWebTransport = createGrpcWebTransport({
      baseUrl,
      useBinaryFormat: true,
      fetch: (input, init) => fetch(input, { ...init, credentials: "include" }),
    })
    const client = createClient(LoginService, grpcWebTransport)
    setClientGrpcWeb(client)
  }, [])

  const GetLoginStatus = useCallback(async (): Promise<LoginStatus> => {
    if (!clientGrpcWeb) {
      throw new Error("Login service not initialized")
    }
    const requestPb: GetLoginStatusRequest = Protobuf.create(
      GetLoginStatusRequestSchema,
      {},
    )
    const replyPb = await clientGrpcWeb.getLoginStatus(requestPb)
    return replyPb
  }, [clientGrpcWeb])

  const reload = useCallback(async () => {
    if (!clientGrpcWeb) {
      return
    }
    setLoading(true)
    try {
      const loginStatus = await GetLoginStatus()
      setCurrent(loginStatus)
    } finally {
      setLoading(false)
    }
  }, [clientGrpcWeb, GetLoginStatus])

  const login = useCallback(async (): Promise<void> => {
    if (loading) {
      return
    }
    setLoading(true)
    window.location.href = `${window.location.origin}/auth/login?return=${encodeURIComponent(window.location.href)}`
  }, [loading])

  useEffect(() => {
    reload()
  }, [reload])

  const serviceInterface = useMemo<LoginServiceInterface | null>(() => {
    if (!clientGrpcWeb) {
      return null
    }
    return {
      current,
      loading,
      GetLoginStatus,
      reload,
      login,
    }
  }, [clientGrpcWeb, current, loading, GetLoginStatus, reload])

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
