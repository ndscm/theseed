import * as Protobuf from "@bufbuild/protobuf"
import { createClient } from "@connectrpc/connect"
import type { Client as ConnectClient } from "@connectrpc/connect"
import { createGrpcWebTransport } from "@connectrpc/connect-web"
import React, { useCallback, useEffect, useMemo, useState } from "react"

import {
  type TerminalInputFrame,
  type TerminalOutputFrame,
} from "../../../../../../infra/terminal/proto/terminal_pb"
import {
  HooinInvadeService,
  type SendTerminalInputRequest,
  SendTerminalInputRequestSchema,
  type StartTerminalRequest,
  StartTerminalRequestSchema,
} from "../../../proto/invade_pb"

interface HooinInvadeServiceInterface {
  /**
   * Opens a terminal on a person's workstation and streams what it prints,
   * until the shell exits or `signal` is aborted.
   *
   * `start` names the session the terminal is opened under; keystrokes go back
   * to it through `SendTerminalInput`, which is a separate call because
   * grpc-web has no client streaming.
   *
   * Aborting `signal` is the only way to end a terminal early: the stream is
   * what the server holds the shell open under, and abandoning the iterator
   * parked on its next frame would not tell it anything.
   */
  StartTerminal: (
    personId: string,
    start: TerminalInputFrame,
    signal?: AbortSignal,
  ) => AsyncIterable<TerminalOutputFrame>

  /**
   * Types at a terminal `StartTerminal` opened, or tells it its window has
   * been resized. `personId` is the workstation the terminal runs on — the one
   * it was opened with — and the frame's `session_uuid` names which of that
   * person's terminals. The frame's `start` must be unset.
   *
   * Nothing orders two of these calls in flight at once. A caller with more
   * than one keystroke to deliver must await each before sending the next, or
   * the shell may read them out of order.
   */
  SendTerminalInput: (
    personId: string,
    frame: TerminalInputFrame,
  ) => Promise<void>
}

export const HooinInvadeServiceContext =
  React.createContext<HooinInvadeServiceInterface | null>(null)

export const HooinInvadeServiceProvider: React.FC<{
  children?: React.ReactNode
}> = ({ children }) => {
  const [clientGrpcWeb, setClientGrpcWeb] = useState<ConnectClient<
    typeof HooinInvadeService
  > | null>(null)

  useEffect(() => {
    const baseUrl = "/"
    const grpcWebTransport = createGrpcWebTransport({
      baseUrl,
      useBinaryFormat: true,
      fetch: (input, init) => fetch(input, { ...init, credentials: "include" }),
    })
    const client = createClient(HooinInvadeService, grpcWebTransport)
    setClientGrpcWeb(client)
  }, [])

  const StartTerminal = useCallback(
    (
      personId: string,
      start: TerminalInputFrame,
      signal?: AbortSignal,
    ): AsyncIterable<TerminalOutputFrame> => {
      if (!clientGrpcWeb) {
        throw new Error("Hooin Invade service not initialized")
      }
      const requestPb: StartTerminalRequest = Protobuf.create(
        StartTerminalRequestSchema,
        { personId, start },
      )
      return clientGrpcWeb.startTerminal(requestPb, signal ? { signal } : {})
    },
    [clientGrpcWeb],
  )

  const SendTerminalInput = useCallback(
    async (personId: string, frame: TerminalInputFrame): Promise<void> => {
      if (!clientGrpcWeb) {
        throw new Error("Hooin Invade service not initialized")
      }
      const requestPb: SendTerminalInputRequest = Protobuf.create(
        SendTerminalInputRequestSchema,
        { personId, frame },
      )
      await clientGrpcWeb.sendTerminalInput(requestPb)
    },
    [clientGrpcWeb],
  )

  const serviceInterface = useMemo<HooinInvadeServiceInterface | null>(() => {
    if (!clientGrpcWeb) {
      return null
    }
    return {
      StartTerminal,
      SendTerminalInput,
    }
  }, [clientGrpcWeb, StartTerminal, SendTerminalInput])

  return (
    <HooinInvadeServiceContext.Provider value={serviceInterface}>
      {children}
    </HooinInvadeServiceContext.Provider>
  )
}

export const useHooinInvadeService = () => {
  return React.useContext(HooinInvadeServiceContext)
}

export default HooinInvadeServiceContext
