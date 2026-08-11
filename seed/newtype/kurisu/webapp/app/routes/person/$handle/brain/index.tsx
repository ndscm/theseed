import * as Protobuf from "@bufbuild/protobuf"
import * as ProtobufWkt from "@bufbuild/protobuf/wkt"
import React, {
  useCallback,
  useEffect,
  useLayoutEffect,
  useMemo,
  useRef,
  useState,
} from "react"
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

const RECONNECT_RETRY_DELAY_MS = 10 * 60 * 1000

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
  const [historyFilter, setHistoryFilter] = useState<string>("")
  const [historyPageToken, setHistoryPageToken] = useState<string>("")
  const [isTopicShown, setIsTopicShown] = useState(false)
  const [inputTopic, setInputTopic] = useState("")
  const [inputText, setInputText] = useState("")
  const [isMultilineEnabled, setIsMultilineEnabled] = useState(true)

  const refOfThreadsContainer = useRef<HTMLDivElement>(null)
  const refOfThreadsStart = useRef<HTMLDivElement>(null)
  const refOfThreadsEnd = useRef<HTMLDivElement>(null)
  const refOfThreadsEndVisible = useRef(true)

  const refOfLatestStepEmitTime = useRef<ProtobufWkt.Timestamp | null>(null)
  const refOfPreLoadScrollHeight = useRef<number | null>(null)

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
      const now = new Date()
      refOfLatestStepEmitTime.current = ProtobufWkt.timestampFromDate(now)
      const filter = `emit_time < timestamp("${now.toISOString()}")`
      setHistoryFilter(filter)
      const firstPage = await dictateService.ListBrainSteps(personId, {
        filter,
        orderBy: "emit_time desc",
      })
      if (cancelled()) {
        return
      }
      for (const step of firstPage.brainSteps) {
        showStep(step)
      }
      setHistoryPageToken(firstPage.nextPageToken)
      const personTopic = Protobuf.create(PersonTopicSchema, {
        personId,
        topic: "",
      })

      // Keep the live feed alive across drops: consume the subscription, and
      // when it ends or errors reconnect until this effect is cancelled. Each
      // subscribe passes the newest step we've shown as `since`, so the server
      // backfills the steps emitted while we were disconnected (and the gap
      // between the history page above and the first subscribe). A stream that
      // delivered a step is healthy, so the retry delay resets on delivery and
      // only applies to reconnects that deliver nothing.
      let retryDelayMs = 0
      while (!cancelled()) {
        if (retryDelayMs > 0) {
          await new Promise((resolve) => setTimeout(resolve, retryDelayMs))
          if (cancelled()) {
            return
          }
        }
        let delivered = false
        const since = refOfLatestStepEmitTime.current ?? undefined
        const stream = dictateService.SubscribeBrainStep([personTopic], {
          since,
        })
        try {
          for await (const step of stream) {
            if (cancelled()) {
              return
            }
            delivered = true
            showStep(step)
            if (
              step.emitTime &&
              (!refOfLatestStepEmitTime.current ||
                ProtobufWkt.timestampMs(step.emitTime) >
                  ProtobufWkt.timestampMs(refOfLatestStepEmitTime.current))
            ) {
              refOfLatestStepEmitTime.current = step.emitTime
            }
          }
        } catch (err: unknown) {
          if (cancelled()) {
            return
          }
          console.error("SubscribeBrainStep stream error", err)
        }
        retryDelayMs = delivered ? 0 : RECONNECT_RETRY_DELAY_MS
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

  // Load older threads whenever the top sentinel scrolls into view. The first
  // page is loaded by reload; this fetches successively older pages, advancing
  // `historyPageToken` and re-subscribing each time so the next intersection
  // gets the following page. It unobserves the moment it fires so a single page
  // is in flight at a time. An empty `historyPageToken` means there is nothing
  // to fetch — the first page hasn't loaded yet, or history is exhausted — so
  // the effect stays idle. ListBrainSteps returns steps newest-first (order_by);
  // showStep dedups against live steps.
  useEffect(() => {
    const startEl = refOfThreadsStart.current
    if (
      !startEl ||
      !dictateService ||
      !personId ||
      !historyFilter ||
      !historyPageToken
    ) {
      return
    }
    const observer = new IntersectionObserver(([entry]) => {
      if (!entry?.isIntersecting) {
        return
      }
      // Single-flight: stop observing until the cursor advances and this effect
      // re-subscribes for the next page. If the loaded page doesn't push the
      // sentinel out of view (e.g. many system events collapse into one
      // counter, or steps dedup away), the recreated observer intersects again
      // and fetches the next page, so history keeps loading until enough
      // content fills the viewport above the sentinel or the cursor is
      // exhausted — rather than stalling until the user scrolls.
      observer.unobserve(startEl)
      void (async () => {
        const response = await dictateService.ListBrainSteps(personId, {
          pageToken: historyPageToken,
          filter: historyFilter,
          orderBy: "emit_time desc",
        })
        // Remember the pre-prepend height so the layout effect can hold the
        // viewport steady over the older threads about to be inserted above.
        refOfPreLoadScrollHeight.current =
          refOfThreadsContainer.current?.scrollHeight ?? null
        for (const step of response.brainSteps) {
          showStep(step)
        }
        // "" marks exhaustion; any other token is the next page's cursor.
        setHistoryPageToken(response.nextPageToken)
      })()
    })
    observer.observe(startEl)
    return () => {
      observer.disconnect()
    }
  }, [dictateService, personId, historyFilter, historyPageToken, showStep])

  // Track whether the tail is on screen so live steps only auto-scroll when the
  // user is already at the bottom, and never while they are reading history up
  // top.
  useEffect(() => {
    const endEl = refOfThreadsEnd.current
    if (!endEl) {
      return
    }
    const observer = new IntersectionObserver(([entry]) => {
      refOfThreadsEndVisible.current = entry?.isIntersecting ?? false
    })
    observer.observe(endEl)
    return () => {
      observer.disconnect()
    }
  }, [])

  // After an older page is prepended, restore the scroll offset by the amount
  // the content grew, so the threads the user was reading stay in place instead
  // of jumping. Runs before paint to avoid a visible flicker.
  useLayoutEffect(() => {
    const container = refOfThreadsContainer.current
    const preLoadScrollHeight = refOfPreLoadScrollHeight.current
    if (!container || preLoadScrollHeight === null) {
      return
    }
    container.scrollTop += container.scrollHeight - preLoadScrollHeight
    refOfPreLoadScrollHeight.current = null
  }, [threads])

  useEffect(() => {
    if (refOfThreadsEndVisible.current) {
      refOfThreadsEnd.current?.scrollIntoView({ behavior: "smooth" })
    }
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
      <div
        ref={refOfThreadsContainer}
        className={tw({ layout: "min-h-0 flex-1 overflow-auto px-7 py-6" })}
      >
        <div
          className={tw({ layout: "mx-auto flex max-w-3xl flex-col gap-4" })}
        >
          <div ref={refOfThreadsStart} />
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
