import * as Protobuf from "@bufbuild/protobuf"
import React, { useCallback, useEffect, useRef, useState } from "react"
import { useTranslation } from "react-i18next"

import { TerminalIcon } from "lucide-react"

import "@xterm/xterm/css/xterm.css"

import tw from "../../../../../../../../devprod/ts/grouping-tailwind"
import {
  type TerminalInputFrame,
  TerminalInputFrameSchema,
} from "../../../../../../../../infra/terminal/proto/terminal_pb"
import { useHooinInvadeService } from "../../../../../../../hooin/invade/client/tsx/HooinInvadeServiceContext"

const UbuntuTerminalTheme = {
  background: "#300a24",
  black: "#2e3436",
  blue: "#3465a4",
  brightBlack: "#555753",
  brightBlue: "#729fcf",
  brightCyan: "#34e2e2",
  brightGreen: "#8ae234",
  brightMagenta: "#ad7fa8",
  brightRed: "#ef2929",
  brightWhite: "#eeeeec",
  brightYellow: "#fce94f",
  cursor: "#ffffff",
  cursorAccent: "#300a24",
  cyan: "#06989a",
  foreground: "#ffffff",
  green: "#4e9a06",
  magenta: "#75507b",
  red: "#cc0000",
  selectionBackground: "#b5d5ff",
  white: "#d3d7cf",
  yellow: "#c4a000",
}

const PersonTerminalPage: React.FC<{ params: { handle: string } }> = ({
  params,
}) => {
  const { handle } = params
  const { t } = useTranslation("person")
  const invadeService = useHooinInvadeService()
  const [connected, setConnected] = useState(false)
  const [endMessage, setEndMessage] = useState("")
  const refOfTerminalContainer = useRef<HTMLDivElement>(null)

  const personHandle = handle
    .trim()
    .replace(/^@/, "")
    .replace(/@$/, "")
    .toLowerCase()
    .trim()

  const connect = useCallback(() => {
    setEndMessage("")
    setConnected(true)
  }, [])

  useEffect(() => {
    const container = refOfTerminalContainer.current
    if (!invadeService || !container || !connected) {
      return
    }

    // Everything below is torn down by the cleanup this effect returns.
    // `disposed` covers the window between the dynamic import resolving and
    // that cleanup running, during which the effect may already be stale.
    let disposed = false
    let dispose = () => {}

    const boot = async () => {
      // xterm reaches for the DOM as soon as it is constructed, and this
      // route is prerendered to HTML at build time, in Node. Importing it
      // here rather than at module scope keeps it out of that run.
      const [{ Terminal }, { FitAddon }] = await Promise.all([
        import("@xterm/xterm"),
        import("@xterm/addon-fit"),
      ])
      if (disposed) {
        return
      }

      const terminal = new Terminal({
        cursorBlink: true,
        fontFamily:
          'ui-monospace, SFMono-Regular, "SF Mono", Menlo, Consolas, monospace',
        fontSize: 16,
        theme: UbuntuTerminalTheme,
      })
      const fitAddon = new FitAddon()
      terminal.loadAddon(fitAddon)
      terminal.open(container)
      fitAddon.fit()

      // The session names this terminal to the calls that type at it. It is
      // unguessable because knowing it, as the same logged-in person, is enough
      // to type at the shell.
      const sessionUuid = crypto.randomUUID()

      // Aborting is what ends the terminal: the stream is what the server holds
      // the shell open under, and a reader that merely walks away leaves it
      // running.
      const abortController = new AbortController()
      let exited = false

      // The terminal goes away with the shell, leaving the button that opens
      // the next one. Aborting is this page tearing the terminal down itself,
      // and has nobody left to tell.
      const onEnd = (message: string) => {
        if (exited || abortController.signal.aborted) {
          return
        }
        exited = true
        setEndMessage(message)
        setConnected(false)
      }

      // Every frame is stamped with its position in the input, counting from 0,
      // so the agent can apply them in order however they arrive. The start
      // frame below is 0; each keystroke or resize takes the next.
      let nextInputIndex = 0
      const takeInputIndex = () => {
        return nextInputIndex++
      }

      // grpc-web has no client streaming, so every keystroke is its own call.
      // A burst of them is in flight at once and may arrive in any order, which
      // the index above is what carries them through.
      //
      // A call that fails has lost what it carried — a keystroke that overtook
      // the stream opening the terminal, typically, since nothing orders the
      // two. The shell is still there, and typing at it again still reaches it,
      // so this is noted rather than treated as the end of the terminal.
      const sendInput = async (frame: TerminalInputFrame) => {
        try {
          if (abortController.signal.aborted || exited) {
            return
          }
          await invadeService.SendTerminalInput(personHandle, frame)
        } catch (error) {
          console.warn("terminal input lost:", error)
        }
      }
      const sendKeystrokes = (input: Uint8Array) => {
        sendInput(
          Protobuf.create(TerminalInputFrameSchema, {
            sessionUuid,
            index: takeInputIndex(),
            input,
          }),
        )
      }
      const sendResize = (rows: number, cols: number) => {
        sendInput(
          Protobuf.create(TerminalInputFrameSchema, {
            sessionUuid,
            index: takeInputIndex(),
            resize: { rows, cols },
          }),
        )
      }

      const readOutput = async () => {
        try {
          const stream = invadeService.StartTerminal(
            personHandle,
            Protobuf.create(TerminalInputFrameSchema, {
              sessionUuid,
              index: takeInputIndex(),
              start: { rows: terminal.rows, cols: terminal.cols },
            }),
            abortController.signal,
          )
          // Output is indexed the same way. Nothing here reorders it: a frame
          // behind what has already been rendered is a straggler and dropped,
          // and a gap is stepped over rather than waited on.
          let nextOutputIndex = 0
          for await (const frame of stream) {
            if (frame.index < nextOutputIndex) {
              continue
            }
            nextOutputIndex = frame.index + 1

            // The fields are not a oneof, so a frame may carry several at once.
            // Render the output before acting on exit, or a shell's last words
            // are lost with it.
            if (frame.output && frame.output.length > 0) {
              terminal.write(frame.output)
            }
            if (frame.error) {
              // Something went wrong around the terminal — input lost in
              // transit, typically — but the shell is still running. Nothing to
              // recover; note it and keep going.
              console.warn("terminal error:", frame.error.message)
            }
            if (frame.exit) {
              onEnd(frame.exit.message)
              return
            }
          }
          // The stream ending without an exit frame means the connection
          // dropped rather than the shell ending, which must still be shown:
          // nothing more will ever arrive on this terminal.
          onEnd(t("terminal.connectionClosed", "Connection closed"))
        } catch (error) {
          onEnd(error instanceof Error ? error.message : String(error))
        }
      }

      readOutput()

      // onData carries what the user typed, already decoded; onBinary
      // carries bytes the terminal could not represent as text, one per
      // char code. Both reach the shell as keystrokes.
      const encoder = new TextEncoder()
      const dataSub = terminal.onData((data) => {
        sendKeystrokes(encoder.encode(data))
      })
      const binarySub = terminal.onBinary((data) => {
        const bytes = new Uint8Array(data.length)
        for (let i = 0; i < data.length; i += 1) {
          bytes[i] = data.charCodeAt(i) & 0xff
        }
        sendKeystrokes(bytes)
      })

      // fit() recomputes rows and cols from the container size; onResize
      // then fires only when they actually changed, so the shell hears
      // about a resize once per real change rather than once per pixel.
      const resizeSub = terminal.onResize(({ rows, cols }) => {
        sendResize(rows, cols)
      })
      const resizeObserver = new ResizeObserver(() => {
        fitAddon.fit()
      })
      resizeObserver.observe(container)

      terminal.focus()

      dispose = () => {
        resizeObserver.disconnect()
        resizeSub.dispose()
        binarySub.dispose()
        dataSub.dispose()
        abortController.abort()
        terminal.dispose()
      }
    }

    boot()

    return () => {
      disposed = true
      dispose()
    }
  }, [invadeService, personHandle, connected, t])

  return (
    <main className={tw({ layout: "flex min-h-0 flex-1 flex-col" })}>
      {connected ? (
        <div className={tw({ layout: "min-h-0 flex-1" })}>
          <div
            ref={refOfTerminalContainer}
            className={tw({
              layout: "h-full w-full overflow-hidden",
            })}
            style={{ backgroundColor: UbuntuTerminalTheme.background }}
          />
        </div>
      ) : (
        <div
          className={tw({
            layout:
              "flex min-h-0 flex-1 flex-col items-center justify-center gap-3",
          })}
        >
          <button
            className={tw({ component: "btn btn-primary" })}
            onClick={connect}
            disabled={!invadeService}
          >
            <TerminalIcon />
            {t("terminal.connect", "Connect")}
          </button>
          {endMessage && (
            <span
              className={tw({
                layout: "max-w-md truncate",
                appearance: "text-base-content/60 text-xs",
              })}
            >
              {endMessage}
            </span>
          )}
        </div>
      )}
    </main>
  )
}

export default PersonTerminalPage
