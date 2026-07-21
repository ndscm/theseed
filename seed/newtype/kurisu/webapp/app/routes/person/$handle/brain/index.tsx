import * as Protobuf from "@bufbuild/protobuf"
import React, { useCallback, useEffect, useRef, useState } from "react"
import { useTranslation } from "react-i18next"

import { BracketsIcon, CornerDownLeftIcon, SendIcon } from "lucide-react"

import tw from "../../../../../../../../devprod/ts/grouping-tailwind"
import {
  type BrainInput,
  BrainInputSchema,
} from "../../../../../../../gajetto/proto/brain_pb"
import { useHooinDictateService } from "../../../../../../../hooin/dictate/client/tsx/HooinDictateServiceContext"
import { useHooinRosterService } from "../../../../../../../hooin/roster/client/tsx/HooinRosterServiceContext"
import BrainThreadPanel, {
  type BrainThread,
} from "../../../../../components/BrainThreadPanel"

const PersonBrainPage: React.FC<{ params: { handle: string } }> = ({
  params,
}) => {
  const { handle } = params
  const { t } = useTranslation("person")
  const rosterService = useHooinRosterService()
  const dictateService = useHooinDictateService()

  const [personId, setPersonId] = useState<string>("")
  const [threads, setThreads] = useState<BrainThread[]>([])
  const [isTopicShown, setIsTopicShown] = useState(false)
  const [inputTopic, setInputTopic] = useState("")
  const [inputText, setInputText] = useState("")
  const [isMultilineEnabled, setIsMultilineEnabled] = useState(true)
  const refOfThreadsEnd = useRef<HTMLDivElement>(null)

  const personHandle = handle
    .trim()
    .replace(/^@/, "")
    .replace(/@$/, "")
    .toLowerCase()
    .trim()

  useEffect(() => {
    void (async () => {
      if (!rosterService) {
        return
      }
      const member = await rosterService.GetTeamMember("", {
        handle: personHandle,
      })
      setPersonId(member.personId)
    })()
  }, [rosterService, personHandle])

  const reload = useCallback(async () => {
    if (!rosterService) {
      return
    }
  }, [rosterService])

  useEffect(() => {
    reload()
  }, [reload])

  useEffect(() => {
    refOfThreadsEnd.current?.scrollIntoView({ behavior: "smooth" })
  }, [threads])

  const sendMessage = useCallback(() => {
    const text = inputText.trim()
    if (!text || !dictateService || !personId) {
      return
    }

    const uuid = crypto.randomUUID()
    const brainInput: BrainInput = Protobuf.create(BrainInputSchema, {
      uuid,
      taskUuid: uuid,
      text,
      topic: (isTopicShown && inputTopic) || "default",
    })
    const stream = dictateService.SendBrainInputStreamBrainStep(
      personId,
      brainInput,
    )
    const newThread: BrainThread = {
      personId,
      input: brainInput,
      steps: [],
      live: stream,
    }
    setThreads((prev) => [...prev, newThread])
    setInputText("")
  }, [inputText, dictateService, personId, isTopicShown, inputTopic])

  const handleKeyDown = useCallback(
    (e: React.KeyboardEvent<HTMLTextAreaElement>) => {
      if (e.key === "Enter") {
        if (e.nativeEvent.isComposing) {
          // Let the IME commit its candidate; don't send or intercept.
          return
        }
        if (e.shiftKey) {
          // Shift+Enter always inserts a carriage return.
          return
        }
        if (e.ctrlKey || e.metaKey) {
          // Ctrl/Cmd+Enter always sends.
          e.preventDefault()
          sendMessage()
          return
        }
        if (!isMultilineEnabled) {
          // Plain Enter sends when multi-line mode is off; otherwise it
          // inserts a carriage return via the default behavior.
          e.preventDefault()
          sendMessage()
          return
        }
      }
    },
    [sendMessage, isMultilineEnabled],
  )

  return (
    <main className={tw({ layout: "flex min-h-0 flex-1 flex-col" })}>
      <div className={tw({ layout: "min-h-0 flex-1 overflow-auto px-7 py-6" })}>
        <div className={tw({ layout: "flex max-w-3xl flex-col gap-4" })}>
          {threads.map((thread) => (
            <BrainThreadPanel key={thread.input.uuid} thread={thread} />
          ))}
          <div ref={refOfThreadsEnd} />
        </div>
      </div>

      <div
        className={tw({
          layout: "sticky bottom-0 shrink-0 border-t px-7 py-4",
          appearance: "border-base-300 bg-base-100",
        })}
      >
        <div
          className={tw({
            layout: "flex max-w-3xl flex-col items-stretch",
            appearance: "border-base-300 rounded-lg border p-1",
          })}
        >
          {isTopicShown && (
            <div
              className={tw({
                layout: "flex",
              })}
            >
              <input
                type="text"
                className={tw({
                  component: "input input-ghost",
                  layout: "flex-1",
                  state: "focus:outline-none",
                })}
                placeholder={t("brain.topicPlaceholder", "Topic")}
                value={inputTopic}
                onChange={(e) => setInputTopic(e.target.value)}
              />
            </div>
          )}
          <div
            className={tw({
              layout: "flex",
            })}
          >
            <textarea
              className={tw({
                component: "textarea textarea-ghost",
                layout: "flex-1 resize-none",
                state: "focus:outline-none",
              })}
              rows={3}
              placeholder={t("brain.brainInputPlaceholder", "Raw Brain Input")}
              value={inputText}
              onChange={(e) => setInputText(e.target.value)}
              onKeyDown={handleKeyDown}
            />
          </div>
          <div className={tw({ layout: "flex justify-end gap-1" })}>
            <button
              type="button"
              className={tw(
                {
                  component: `btn btn-square btn-ghost`,
                  layout: "shrink-0",
                  state:
                    "hover:border-base-300 hover:bg-transparent hover:shadow-none",
                },
                isTopicShown
                  ? {
                      appearance: "text-primary",
                      state: "hover:text-primary",
                    }
                  : {
                      appearance: "text-base-content/40",
                      state: "hover:text-base-content/40",
                    },
              )}
              onClick={() => {
                if (isTopicShown) {
                  setInputTopic("")
                }
                setIsTopicShown((prev) => !prev)
              }}
              aria-pressed={isTopicShown}
            >
              <BracketsIcon />
            </button>
            <button
              type="button"
              className={tw(
                {
                  component: `btn btn-square btn-ghost`,
                  layout: "shrink-0",
                  state:
                    "hover:border-base-300 hover:bg-transparent hover:shadow-none",
                },
                isMultilineEnabled
                  ? {
                      appearance: "text-primary",
                      state: "hover:text-primary",
                    }
                  : {
                      appearance: "text-base-content/40",
                      state: "hover:text-base-content/40",
                    },
              )}
              onClick={() => setIsMultilineEnabled((prev) => !prev)}
              aria-pressed={isMultilineEnabled}
              title={
                isMultilineEnabled
                  ? t(
                      "brain.multilineOnTooltip",
                      "Multi-line enabled. Use <Ctrl> + <Enter> to send messages",
                    )
                  : t(
                      "brain.multilineOffTooltip",
                      "Use <Enter> to send messages",
                    )
              }
            >
              <CornerDownLeftIcon />
            </button>
            <button
              className={tw({
                component: "btn btn-square btn-primary btn-ghost",
                layout: "shrink-0",
              })}
              onClick={sendMessage}
              disabled={!inputText.trim() || !dictateService || !personId}
            >
              <SendIcon />
            </button>
          </div>
        </div>
      </div>
    </main>
  )
}

export default PersonBrainPage
