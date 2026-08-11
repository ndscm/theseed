import * as Protobuf from "@bufbuild/protobuf"
import * as ProtobufWkt from "@bufbuild/protobuf/wkt"
import { createClient } from "@connectrpc/connect"
import type { Client as ConnectClient } from "@connectrpc/connect"
import { createGrpcWebTransport } from "@connectrpc/connect-web"
import React, { useCallback, useEffect, useMemo, useState } from "react"

import {
  type BrainInput,
  type BrainStep,
} from "../../../../../gajetto/proto/brain_pb"
import {
  HooinDictateService,
  type ListBrainStepsRequest,
  ListBrainStepsRequestSchema,
  type ListBrainStepsResponse,
  type PersonTopic,
  type SendBrainInputRequest,
  SendBrainInputRequestSchema,
  type SubscribeBrainStepRequest,
  SubscribeBrainStepRequestSchema,
} from "../../../proto/dictate_pb"

interface HooinDictateServiceInterface {
  ListBrainSteps: (
    personId: string,
    options?: {
      pageToken?: string
      filter?: string
      orderBy?: string
    },
  ) => Promise<ListBrainStepsResponse>
  SendBrainInput: (
    personId: string,
    brainInput: BrainInput,
  ) => Promise<BrainStep>
  SendBrainInputStreamBrainStep: (
    personId: string,
    brainInput: BrainInput,
  ) => AsyncIterable<BrainStep>
  SubscribeBrainStep: (
    personTopics: PersonTopic[],
    options?: {
      since?: ProtobufWkt.Timestamp
    },
  ) => AsyncIterable<BrainStep>
}

export const HooinDictateServiceContext =
  React.createContext<HooinDictateServiceInterface | null>(null)

export const HooinDictateServiceProvider: React.FC<{
  children?: React.ReactNode
}> = ({ children }) => {
  const [clientGrpcWeb, setClientGrpcWeb] = useState<ConnectClient<
    typeof HooinDictateService
  > | null>(null)

  useEffect(() => {
    const baseUrl = "/"
    const grpcWebTransport = createGrpcWebTransport({
      baseUrl,
      useBinaryFormat: true,
      fetch: (input, init) => fetch(input, { ...init, credentials: "include" }),
    })
    const client = createClient(HooinDictateService, grpcWebTransport)
    setClientGrpcWeb(client)
  }, [])

  const ListBrainSteps = useCallback(
    async (
      personId: string,
      options?: {
        pageToken?: string
        filter?: string
        orderBy?: string
      },
    ): Promise<ListBrainStepsResponse> => {
      if (!clientGrpcWeb) {
        throw new Error("Hooin Dictate service not initialized")
      }
      const requestPb: ListBrainStepsRequest = Protobuf.create(
        ListBrainStepsRequestSchema,
        {
          parent: `people/${personId}`,
          pageToken: options?.pageToken ?? "",
          filter: options?.filter ?? "",
          orderBy: options?.orderBy ?? "",
        },
      )
      return await clientGrpcWeb.listBrainSteps(requestPb)
    },
    [clientGrpcWeb],
  )

  const SendBrainInput = useCallback(
    async (personId: string, brainInput: BrainInput): Promise<BrainStep> => {
      if (!clientGrpcWeb) {
        throw new Error("Hooin Dictate service not initialized")
      }
      const requestPb: SendBrainInputRequest = Protobuf.create(
        SendBrainInputRequestSchema,
        { personId, brainInput },
      )
      return await clientGrpcWeb.sendBrainInput(requestPb)
    },
    [clientGrpcWeb],
  )

  const SendBrainInputStreamBrainStep = useCallback(
    (personId: string, brainInput: BrainInput): AsyncIterable<BrainStep> => {
      if (!clientGrpcWeb) {
        throw new Error("Hooin Dictate service not initialized")
      }
      const requestPb: SendBrainInputRequest = Protobuf.create(
        SendBrainInputRequestSchema,
        { personId, brainInput },
      )
      return clientGrpcWeb.sendBrainInputStreamBrainStep(requestPb)
    },
    [clientGrpcWeb],
  )

  const SubscribeBrainStep = useCallback(
    (
      personTopics: PersonTopic[],
      options?: {
        since?: ProtobufWkt.Timestamp
      },
    ): AsyncIterable<BrainStep> => {
      if (!clientGrpcWeb) {
        throw new Error("Hooin Dictate service not initialized")
      }
      const requestPb: SubscribeBrainStepRequest = Protobuf.create(
        SubscribeBrainStepRequestSchema,
        { personTopics, since: options?.since },
      )
      return clientGrpcWeb.subscribeBrainStep(requestPb)
    },
    [clientGrpcWeb],
  )

  const serviceInterface = useMemo<HooinDictateServiceInterface | null>(() => {
    if (!clientGrpcWeb) {
      return null
    }
    return {
      ListBrainSteps,
      SendBrainInput,
      SendBrainInputStreamBrainStep,
      SubscribeBrainStep,
    }
  }, [
    clientGrpcWeb,
    ListBrainSteps,
    SendBrainInput,
    SendBrainInputStreamBrainStep,
    SubscribeBrainStep,
  ])

  return (
    <HooinDictateServiceContext.Provider value={serviceInterface}>
      {children}
    </HooinDictateServiceContext.Provider>
  )
}

export const useHooinDictateService = () => {
  return React.useContext(HooinDictateServiceContext)
}

export default HooinDictateServiceContext
