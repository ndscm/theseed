import * as Protobuf from "@bufbuild/protobuf"
import React, { useCallback, useEffect, useRef } from "react"

import tw from "../../../../devprod/ts/grouping-tailwind/dist"
import MarkdownView from "../../../../visual/markdown/tsx/MarkdownView"
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

const markdownClassNames = {
  a: tw({ appearance: "text-primary underline", state: "hover:opacity-80" }),
  blockquote: tw({
    layout: "my-2 pl-3",
    appearance: "border-base-300 text-neutral border-l-2 italic",
  }),
  code: tw({
    appearance:
      "bg-base-300 text-base-content rounded px-1 py-0.5 font-mono text-xs",
  }),
  em: tw({ appearance: "italic" }),
  h1: tw({
    appearance:
      "text-base-content mt-4 mb-2 text-lg font-semibold tracking-tight",
  }),
  h2: tw({
    appearance:
      "text-base-content mt-4 mb-2 text-base font-semibold tracking-tight",
  }),
  h3: tw({ appearance: "text-base-content mt-3 mb-1.5 text-sm font-semibold" }),
  h4: tw({ appearance: "text-base-content mt-3 mb-1.5 text-sm font-semibold" }),
  h5: tw({
    appearance: "text-neutral mt-2 mb-1 text-xs font-semibold uppercase",
  }),
  h6: tw({
    appearance: "text-neutral mt-2 mb-1 text-xs font-semibold uppercase",
  }),
  hr: tw({ appearance: "border-base-200 my-3" }),
  li: tw({ layout: "mt-1" }),
  ol: tw({ layout: "my-2 pl-5", appearance: "list-decimal" }),
  p: tw({ layout: "my-2", appearance: "leading-relaxed" }),
  pre: tw({
    layout: "my-2 overflow-x-auto",
    appearance:
      "bg-base-300 text-base-content rounded-lg p-3 font-mono text-xs leading-relaxed",
  }),
  strong: tw({ appearance: "font-semibold" }),
  table: tw({
    layout: "my-2 block w-full overflow-x-auto",
    appearance: "border-base-200 border-collapse border text-sm",
  }),
  thead: tw({ appearance: "bg-base-200" }),
  th: tw({
    layout: "px-3 py-1.5",
    appearance:
      "border-base-200 text-base-content border text-left font-medium",
  }),
  td: tw({
    layout: "px-3 py-1.5",
    appearance: "border-base-200 text-base-content border",
  }),
  ul: tw({ layout: "my-2 pl-5", appearance: "list-disc" }),
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
      {step.type === "result" && (
        <div
          className={tw({
            layout: "mt-3",
            appearance: "bg-base-200 text-base-content rounded-lg p-3 text-sm",
          })}
        >
          <MarkdownView
            content={step.data?.result?.toString() || ""}
            className={tw({
              layout: "[&>*:first-child]:mt-0 [&>*:last-child]:mb-0",
              appearance: "break-words",
            })}
            htmlClassNames={markdownClassNames}
          />
        </div>
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
    </KurisuPanel>
  )
}

export default BrainThreadPanel
