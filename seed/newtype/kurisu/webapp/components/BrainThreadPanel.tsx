import * as Protobuf from "@bufbuild/protobuf"
import React, { useCallback, useEffect, useRef } from "react"

import tw from "../../../../devprod/ts/grouping-tailwind/dist"
import {
  type BrainInput,
  type BrainStep,
  BrainStepSchema,
} from "../../../gajetto/proto/brain_pb"
import { useHooinDictateService } from "../../../hooin/dictate/client/tsx/HooinDictateServiceContext"
import KurisuPanel from "./KurisuPanel"

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
  personId: string
  input: BrainInput
}> = ({ personId, input }) => {
  const dictateService = useHooinDictateService()
  const startedRef = useRef(false)
  const [steps, setSteps] = React.useState<BrainStep[]>([])

  const start = useCallback(async () => {
    if (!dictateService || !personId || !input?.text) {
      return
    }
    if (startedRef.current) {
      return
    }
    startedRef.current = true

    const stream = dictateService.SendBrainInputStreamBrainStep(personId, input)
    for await (const step of stream) {
      setSteps((prev) => [...prev, step])
    }
  }, [dictateService, personId, input])

  useEffect(() => {
    start()
  }, [start])

  const lastStep = steps[steps.length - 1]

  return (
    <KurisuPanel title={input.topic} subtitle={input.uuid}>
      <div
        className={tw({
          appearance: "text-base-content text-sm",
        })}
      >
        {input.text}
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
