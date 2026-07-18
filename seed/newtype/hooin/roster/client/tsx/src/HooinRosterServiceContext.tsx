import * as Protobuf from "@bufbuild/protobuf"
import { createClient } from "@connectrpc/connect"
import type { Client as ConnectClient } from "@connectrpc/connect"
import { createGrpcWebTransport } from "@connectrpc/connect-web"
import React, { useCallback, useEffect, useMemo, useState } from "react"

import {
  type GetTeamMemberRequest,
  GetTeamMemberRequestSchema,
  type GetTeamRequest,
  GetTeamRequestSchema,
  HooinRosterService,
  type ListTeamMembersRequest,
  ListTeamMembersRequestSchema,
  type ListTeamMembersResponse,
  type Team,
  type TeamMember,
} from "../../../proto/roster_pb"

interface HooinRosterServiceInterface {
  GetTeam: () => Promise<Team>
  GetTeamMember: (
    personId: string,
    options?: {
      handle?: string
    },
  ) => Promise<TeamMember>
  ListTeamMembers: () => Promise<ListTeamMembersResponse>
}

export const HooinRosterServiceContext =
  React.createContext<HooinRosterServiceInterface | null>(null)

export const HooinRosterServiceProvider: React.FC<{
  children?: React.ReactNode
}> = ({ children }) => {
  const [clientGrpcWeb, setClientGrpcWeb] = useState<ConnectClient<
    typeof HooinRosterService
  > | null>(null)

  useEffect(() => {
    const baseUrl = "/"
    console.info("Hooin Roster Service Base Url:", baseUrl)
    const grpcWebTransport = createGrpcWebTransport({
      baseUrl,
      useBinaryFormat: true,
      fetch: (input, init) => fetch(input, { ...init, credentials: "include" }),
    })
    const client = createClient(HooinRosterService, grpcWebTransport)
    setClientGrpcWeb(client)
  }, [])

  const GetTeam = useCallback(async (): Promise<Team> => {
    if (!clientGrpcWeb) {
      throw new Error("Hooin Roster service not initialized")
    }
    const requestPb: GetTeamRequest = Protobuf.create(GetTeamRequestSchema, {})
    return await clientGrpcWeb.getTeam(requestPb)
  }, [clientGrpcWeb])

  const GetTeamMember = useCallback(
    async (
      personId: string,
      options?: { handle?: string },
    ): Promise<TeamMember> => {
      if (!clientGrpcWeb) {
        throw new Error("Hooin Roster service not initialized")
      }
      const requestPb: GetTeamMemberRequest = Protobuf.create(
        GetTeamMemberRequestSchema,
        {
          personId,
          handle: options?.handle ?? "",
        },
      )
      return await clientGrpcWeb.getTeamMember(requestPb)
    },
    [clientGrpcWeb],
  )

  const ListTeamMembers =
    useCallback(async (): Promise<ListTeamMembersResponse> => {
      if (!clientGrpcWeb) {
        throw new Error("Hooin Roster service not initialized")
      }
      const requestPb: ListTeamMembersRequest = Protobuf.create(
        ListTeamMembersRequestSchema,
        {},
      )
      return await clientGrpcWeb.listTeamMembers(requestPb)
    }, [clientGrpcWeb])

  const serviceInterface = useMemo<HooinRosterServiceInterface | null>(() => {
    if (!clientGrpcWeb) {
      return null
    }
    return {
      GetTeam,
      GetTeamMember,
      ListTeamMembers,
    }
  }, [clientGrpcWeb, GetTeam, GetTeamMember, ListTeamMembers])

  return (
    <HooinRosterServiceContext.Provider value={serviceInterface}>
      {children}
    </HooinRosterServiceContext.Provider>
  )
}

export const useHooinRosterService = () => {
  return React.useContext(HooinRosterServiceContext)
}

export default HooinRosterServiceContext
