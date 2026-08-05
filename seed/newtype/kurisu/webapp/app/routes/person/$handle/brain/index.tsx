import * as Protobuf from "@bufbuild/protobuf"
import * as ProtobufWkt from "@bufbuild/protobuf/wkt"
import React, { useCallback, useEffect, useMemo, useRef, useState } from "react"
import { useTranslation } from "react-i18next"

import { BracketsIcon, CornerDownLeftIcon, SendIcon } from "lucide-react"

import tw from "../../../../../../../../devprod/ts/grouping-tailwind"
import {
  type BrainInput,
  BrainInputSchema,
  type BrainStep,
  BrainStepSchema,
} from "../../../../../../../gajetto/proto/brain_pb"
import { useHooinDictateService } from "../../../../../../../hooin/dictate/client/tsx/HooinDictateServiceContext"
import { PersonTopicSchema } from "../../../../../../../hooin/dictate/proto/dictate_pb"
import { useHooinRosterService } from "../../../../../../../hooin/roster/client/tsx/HooinRosterServiceContext"
import BrainThreadPanel, {
  type BrainThread,
} from "../../../../../components/BrainThreadPanel"

const sortThreadsByEmitTime = (threads: BrainThread[]): BrainThread[] =>
  [...threads].sort(
    (a, b) =>
      ProtobufWkt.timestampMs(a.emitTime) - ProtobufWkt.timestampMs(b.emitTime),
  )

const PersonBrainPage: React.FC<{ params: { handle: string } }> = ({
  params,
}) => {
  const { handle } = params
  const { t } = useTranslation("person")
  const rosterService = useHooinRosterService()
  const dictateService = useHooinDictateService()

  const [personId, setPersonId] = useState<string>("")
  const [threads, setThreads] = useState<{
    [threadUuid: string]: BrainThread | undefined
  }>({})
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

  const showStep = useCallback((step: BrainStep) => {
    setThreads((prev) => {
      const existing = prev[step.threadUuid]
      if (!existing) {
        return {
          ...prev,
          [step.threadUuid]: {
            threadUuid: step.threadUuid,
            emitTime: step.emitTime ?? ProtobufWkt.timestampNow(),
            topic: step.topic,
            steps: { [step.uuid]: step },
          },
        }
      }
      // A given server step can reach us from both the live subscription and a
      // ListBrainSteps history page; ignore the duplicate. (Optimistic input
      // steps carry a distinct client uuid, so they are never dropped here.)
      if (existing.steps[step.uuid]) {
        return prev
      }
      // A thread's emitTime is the earliest emitTime among its steps, so
      // the thread is ordered by when it started (its first step), not by
      // last activity. Keep the smaller of the incoming step's time and
      // the thread's current time. Because steps can arrive out of order
      // (a live/newest step before history backfills earlier ones), this
      // value only drifts downward toward the true start as pages load.
      const threadEmitTime =
        step.emitTime &&
        ProtobufWkt.timestampMs(step.emitTime) <
          ProtobufWkt.timestampMs(existing.emitTime)
          ? step.emitTime
          : existing.emitTime
      return {
        ...prev,
        [step.threadUuid]: {
          ...existing,
          emitTime: threadEmitTime,
          topic: existing.topic || step.topic,
          steps: { ...existing.steps, [step.uuid]: step },
        },
      }
    })
  }, [])

  const reload = useCallback(
    async (cancelled: () => boolean) => {
      if (!dictateService || !personId) {
        return
      }
      const personTopic = Protobuf.create(PersonTopicSchema, {
        personId,
        topic: "",
      })
      const stream = dictateService.SubscribeBrainStep([personTopic])
      try {
        for await (const step of stream) {
          if (cancelled()) {
            break
          }
          showStep(step)
        }
      } catch (err: unknown) {
        if (cancelled()) {
          return
        }
        throw err
      }
    },
    [dictateService, personId, showStep],
  )

  useEffect(() => {
    let cancelled = false
    void reload(() => cancelled)
    return () => {
      cancelled = true
    }
  }, [reload])

  const sortedThreads = useMemo(() => {
    const threadList = Object.values(threads).filter(
      (thread): thread is BrainThread => thread !== undefined,
    )
    return sortThreadsByEmitTime(threadList)
  }, [threads])

  useEffect(() => {
    refOfThreadsEnd.current?.scrollIntoView({ behavior: "smooth" })
  }, [threads])

  const sendThreadInput = useCallback(
    async (topic: string, threadUuid: string, content: { text?: string }) => {
      const text = (content.text || "").trim()
      if (!text || !dictateService || !personId) {
        return
      }
      const uuid = crypto.randomUUID()
      const brainInput: BrainInput = Protobuf.create(BrainInputSchema, {
        uuid,
        threadUuid,
        text,
        topic,
      })
      showStep(
        Protobuf.create(BrainStepSchema, {
          type: "input",
          uuid: crypto.randomUUID(),
          threadUuid,
          topic,
          emitTime: ProtobufWkt.timestampNow(),
          data: { text },
        }),
      )
      const result = await dictateService.SendBrainInput(personId, brainInput)
      return result
    },
    [dictateService, personId, showStep],
  )

  const sendMessage = useCallback(async () => {
    const text = inputText.trim()
    if (!text || !dictateService || !personId) {
      return
    }

    const uuid = crypto.randomUUID()
    const topic = (isTopicShown && inputTopic) || "default"
    const brainInput: BrainInput = Protobuf.create(BrainInputSchema, {
      uuid,
      threadUuid: uuid,
      text,
      topic,
    })
    showStep(
      Protobuf.create(BrainStepSchema, {
        type: "input",
        uuid: crypto.randomUUID(),
        threadUuid: uuid,
        topic,
        emitTime: ProtobufWkt.timestampNow(),
        data: { text },
      }),
    )
    setInputText("")
    const result = await dictateService.SendBrainInput(personId, brainInput)
    return result
  }, [inputText, dictateService, personId, isTopicShown, inputTopic, showStep])

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
        <div
          className={tw({ layout: "mx-auto flex max-w-3xl flex-col gap-4" })}
        >
          {sortedThreads.map((thread) => (
            <BrainThreadPanel
              key={thread.threadUuid}
              thread={thread}
              onSend={sendThreadInput}
            />
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
            layout: "mx-auto flex max-w-3xl flex-col items-stretch",
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
                layout:
                  "field-sizing-content max-h-48 min-h-0 flex-1 resize-none overflow-y-auto",
                state: "focus:outline-none",
              })}
              rows={1}
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
