import * as Protobuf from "@bufbuild/protobuf"
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
  type PersonTopic,
  type SendBrainInputRequest,
  SendBrainInputRequestSchema,
  type SubscribeBrainStepRequest,
  SubscribeBrainStepRequestSchema,
} from "../../../proto/dictate_pb"

interface HooinDictateServiceInterface {
  SendBrainInput: (
    personId: string,
    brainInput: BrainInput,
  ) => Promise<BrainStep>
  SendBrainInputStreamBrainStep: (
    personId: string,
    brainInput: BrainInput,
  ) => AsyncIterable<BrainStep>
  SubscribeBrainStep: (personTopics: PersonTopic[]) => AsyncIterable<BrainStep>
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
    (personTopics: PersonTopic[]): AsyncIterable<BrainStep> => {
      if (!clientGrpcWeb) {
        throw new Error("Hooin Dictate service not initialized")
      }
      const requestPb: SubscribeBrainStepRequest = Protobuf.create(
        SubscribeBrainStepRequestSchema,
        { personTopics },
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
      SendBrainInput,
      SendBrainInputStreamBrainStep,
      SubscribeBrainStep,
    }
  }, [
    clientGrpcWeb,
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
