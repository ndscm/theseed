import * as Protobuf from "@bufbuild/protobuf"
import React, { useCallback, useEffect, useRef, useState } from "react"
import { useTranslation } from "react-i18next"

import { SendIcon } from "lucide-react"

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
  const [inputTopic, setInputTopic] = useState("")
  const [inputText, setInputText] = useState("")
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
      topic: inputTopic || "default",
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
  }, [inputText, dictateService, personId, inputTopic])

  const handleKeyDown = useCallback(
    (e: React.KeyboardEvent) => {
      if (e.key === "Enter" && e.ctrlKey) {
        e.preventDefault()
        sendMessage()
      }
    },
    [sendMessage],
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
          <div
            className={tw({
              layout: "flex",
            })}
          >
            <input
              type="text"
              list="builtin-topics"
              className={tw({
                component: "input input-ghost",
                layout: "flex-1",
                state: "focus:outline-none",
              })}
              placeholder={t("brain.topicPlaceholder", "Topic")}
              value={inputTopic}
              onChange={(e) => setInputTopic(e.target.value)}
            />
            <datalist id="builtin-topics">
              <option value="develop" />
              <option value="review" />
            </datalist>
          </div>
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
              placeholder={t("brain.brainInputPlaceholder", "Brain input")}
              value={inputText}
              onChange={(e) => setInputText(e.target.value)}
              onKeyDown={handleKeyDown}
            />
          </div>
          <div className={tw({ layout: "flex justify-end" })}>
            <button
              className={tw({
                component: "btn btn-primary btn-ghost",
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
