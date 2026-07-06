import * as Protobuf from "@bufbuild/protobuf"
import React, { useCallback, useEffect, useRef } from "react"

import tw from "../../../../devprod/ts/grouping-tailwind/dist"
import {
  type BrainInput,
  type BrainStep,
  BrainStepSchema,
} from "../../../gajetto/proto/brain_pb"
import KurisuPanel from "./KurisuPanel"

export type BrainThread = {
  personId: string
  input: BrainInput
  steps: BrainStep[]
  live: AsyncIterable<BrainStep>
}

const BrainStepItem: React.FC<{ step: BrainStep }> = ({ step }) => {
  const [expanded, setExpanded] = React.useState(false)

  return (
    <div
      className={tw({
        layout: "mt-3",
        appearance: "border-base-200 border-t pt-3",
      })}
    >
      <div
        className={tw({
          layout: "flex cursor-pointer items-center gap-2",
          appearance: "text-neutral text-xs font-medium",
          state: "hover:text-base-content",
        })}
        onClick={() => setExpanded((prev) => !prev)}
      >
        <span className={tw({ appearance: "font-mono" })}>{step.type}</span>
      </div>
      {expanded && (
        <pre
          className={tw({
            layout: "m-0 mt-2 overflow-x-auto",
            appearance:
              "bg-base-200 text-base-content rounded-lg p-3 font-mono text-xs",
          })}
        >
          {Protobuf.toJsonString(BrainStepSchema, step, { prettySpaces: 2 })}
        </pre>
      )}
    </div>
  )
}

const BrainThreadPanel: React.FC<{
  thread: BrainThread
}> = ({ thread }) => {
  const [steps, setSteps] = React.useState<BrainStep[]>([])
  const startedRef = useRef(false)

  useEffect(() => {
    setSteps(thread.steps)
  }, [thread])

  const start = useCallback(async () => {
    if (!thread.live || startedRef.current) {
      return
    }
    startedRef.current = true

    for await (const step of thread.live) {
      setSteps((prev) => [...prev, step])
    }
  }, [thread.live])

  useEffect(() => {
    start()
  }, [start])

  const lastStep = steps[steps.length - 1]

  return (
    <KurisuPanel title={thread.input.topic} subtitle={thread.input.taskUuid}>
      <div
        className={tw({
          appearance: "text-base-content text-sm",
        })}
      >
        {thread.input.text}
      </div>
      {steps.map((step, index) => (
        <BrainStepItem key={index} step={step} />
      ))}
      {!!lastStep && lastStep.type === "result" && (
        <div
          className={tw({
            layout: "mt-3",
            appearance:
              "bg-base-200 text-base-content rounded-lg p-3 text-sm font-medium",
          })}
        >
          {lastStep.data?.result as string}
        </div>
      )}
    </KurisuPanel>
  )
}

export default BrainThreadPanel
