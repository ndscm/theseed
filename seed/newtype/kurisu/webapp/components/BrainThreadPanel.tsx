import * as Protobuf from "@bufbuild/protobuf"
import React, { useCallback, useEffect, useMemo, useRef } from "react"

import {
  CircleCheckIcon,
  CircleDotIcon,
  CircleEllipsisIcon,
  CircleIcon,
  CirclePlusIcon,
  LoaderCircleIcon,
  ZapIcon,
} from "lucide-react"

import tw from "../../../../devprod/ts/grouping-tailwind"
import MarkdownView from "../../../../visual/markdown/tsx/MarkdownView"
import ClaudePayload, {
  type StreamOutputMessage,
} from "../../../gajetto/payload/ts/claude-payload"
import {
  type BrainInput,
  type BrainStep,
  BrainStepSchema,
} from "../../../gajetto/proto/brain_pb"
import BrainThreadInput from "./BrainThreadInput"
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

const AssistantTextMessage: React.FC<{ text: string }> = ({ text }) => {
  return (
    <div
      className={tw({
        layout: "mt-3 p-3",
        appearance:
          "border-base-200 text-base-content rounded-lg border text-sm",
      })}
    >
      <MarkdownView
        content={text}
        className={tw({
          layout: "[&>*:first-child]:mt-0 [&>*:last-child]:mb-0",
          appearance: "break-words",
        })}
        htmlClassNames={markdownClassNames}
      />
    </div>
  )
}

const AssistantThinkingMessage: React.FC<{ thinking: string }> = ({
  thinking,
}) => {
  return (
    <div
      className={tw({
        layout: "mt-3 p-3",
        appearance: "border-base-200 text-neutral rounded-lg border text-sm",
      })}
    >
      {thinking}
    </div>
  )
}

const AssistantToolUseMessage: React.FC<{ name: string; input: unknown }> = ({
  name,
  input,
}) => {
  return (
    <div
      className={tw({
        layout: "mt-3 p-3",
        appearance: "bg-base-200 rounded-lg",
      })}
    >
      <span
        className={tw({
          appearance: "text-base-content font-mono text-xs font-medium",
        })}
      >
        {name}
      </span>
      <pre
        className={tw({
          layout: "m-0 mt-2 overflow-x-auto",
          appearance: "text-neutral font-mono text-xs",
        })}
      >
        {JSON.stringify(input ?? {}, null, 2)}
      </pre>
    </div>
  )
}

const UserTextMessage: React.FC<{ text: string }> = ({ text }) => {
  return (
    <div
      className={tw({
        layout: "mt-3",
        appearance: "bg-base-200 text-base-content rounded-lg p-3 text-sm",
      })}
    >
      <MarkdownView
        content={text}
        className={tw({
          layout: "[&>*:first-child]:mt-0 [&>*:last-child]:mb-0",
          appearance: "break-words",
        })}
        htmlClassNames={markdownClassNames}
      />
    </div>
  )
}

const UserToolResultMessage: React.FC<{
  content: unknown
  isError: boolean
}> = ({ content, isError }) => {
  const text =
    typeof content === "string"
      ? content
      : JSON.stringify(content ?? {}, null, 2)
  return (
    <div
      className={tw({
        layout: "mt-3 p-3",
        appearance: isError
          ? "bg-error/10 text-error rounded-lg"
          : "bg-base-200 text-neutral rounded-lg",
      })}
    >
      <pre
        className={tw({
          layout: "m-0 overflow-x-auto",
          appearance: "font-mono text-xs",
        })}
      >
        {text}
      </pre>
    </div>
  )
}

const BrainStepItem: React.FC<{
  prefix?: React.ReactNode
  title?: React.ReactNode
  step: BrainStep
  children?: React.ReactNode
}> = ({ prefix: prefix, title, step, children }) => {
  const [showRaw, setShowRaw] = React.useState(false)

  return (
    <div
      className={tw({
        layout: "flex flex-col pt-3 pb-3",
        appearance: "border-base-200 border-t",
      })}
    >
      <button
        type="button"
        className={tw({
          layout: "flex items-center gap-2",
          appearance: "text-neutral cursor-pointer text-xs font-medium",
          state: "hover:text-base-content",
        })}
        onClick={() => setShowRaw((prev) => !prev)}
      >
        {prefix}
        {title && <span>{title}</span>}
      </button>
      {children}
      {showRaw && (
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

/**
 * Renders a locally-recorded input step (`type: "input"`).
 *
 * This is the frontend's own record of what the user submitted, appended
 * optimistically the moment they send — before any server round-trip. It is the
 * "record the input" half of the pair; {@link BrainInputStepItem} is the
 * "confirm the server received it" half, so both may show the same text by
 * design.
 *
 * The step's `data` is a plain `{ text }` object created on the client (not a
 * decoded Claude payload), so it is read directly rather than through
 * {@link ClaudePayload.DecodeStreamInput}.
 */
const BrainInputView: React.FC<{ step: BrainStep }> = ({ step }) => {
  if (typeof step.data !== "object") {
    return <>error: missing data</>
  }
  const data = step.data as { text?: string }

  return (
    <div
      className={tw({
        layout: "flex items-center gap-2 pt-3 pb-3",
        appearance: "border-base-200 text-base-content border-t text-sm",
      })}
    >
      <ZapIcon className={tw({ layout: "size-4 shrink-0" })} />
      <span>{data?.text ?? ""}</span>
    </div>
  )
}

/**
 * Renders a server-confirmed input step (`type: "claudecli-input"`).
 *
 * The brain echoes each input it actually wrote to the Claude CLI back over the
 * live stream (see `writeInput` in `topic_runner.go`), so this is the "confirm
 * the server received it" half of the pair — the counterpart to
 * {@link BrainInputView}'s optimistic local record. Both may show the same text
 * by design.
 *
 * The step's `data` is a Claude stream-json input envelope, so it is decoded via
 * {@link ClaudePayload.DecodeStreamInput} and narrowed to the `user` variant.
 */
const BrainInputStepItem: React.FC<{ step: BrainStep }> = ({ step }) => {
  const input = ClaudePayload.DecodeStreamInput(step.data)
  if (!input || input.type !== "user") {
    return <>error: wrong type</>
  }

  const content = input.message?.content ?? ""

  return (
    <BrainStepItem
      key={step.uuid}
      prefix={<CirclePlusIcon className={tw({ layout: "size-4" })} />}
      title={"Input"}
      step={step}
    >
      <UserTextMessage text={content} />
    </BrainStepItem>
  )
}

const BrainSystemStepItem: React.FC<{ step: BrainStep }> = ({ step }) => {
  const output = ClaudePayload.DecodeStreamOutput(step.data)
  if (!output || output.type !== "system") {
    return <>error: wrong type</>
  }
  return (
    <BrainStepItem
      key={step.uuid}
      prefix={<CircleIcon className={tw({ layout: "size-4" })} />}
      title={<>{output.subtype ?? "system"}</>}
      step={step}
    />
  )
}

const BrainSystemSteps: React.FC<{ steps: BrainStep[] }> = ({ steps }) => {
  const [expanded, setExpanded] = React.useState(false)

  return (
    <div className={tw({ layout: "flex flex-col" })}>
      <button
        type="button"
        className={tw({
          layout: "flex items-center justify-start gap-2 pt-3 pb-3",
          appearance:
            "border-base-200 text-neutral cursor-pointer border-t text-xs font-medium",
          state: "hover:text-base-content",
        })}
        onClick={() => setExpanded((prev) => !prev)}
      >
        <CircleIcon className={tw({ layout: "size-4" })} />
        {steps.length} system event{steps.length !== 1 ? "s" : ""}
      </button>
      {expanded &&
        steps.map((step) => (
          <BrainSystemStepItem key={step.uuid} step={step} />
        ))}
    </div>
  )
}

const summarizeAssistantMessage = (message: StreamOutputMessage) => {
  const results = (message.content || []).map((block) => {
    switch (block.type) {
      case "text":
        return "Step"
      case "thinking":
        return "Thinking"
      case "tool_use":
        return block.name ? `Tool: ${block.name}` : "Tool"
      default:
        return block.type
    }
  })
  return results.join("; ")
}

const BrainAssistantStepItem: React.FC<{ step: BrainStep }> = ({ step }) => {
  const output = ClaudePayload.DecodeStreamOutput(step.data)

  if (!output || output.type !== "assistant") {
    return <>error: wrong type</>
  }

  const message = output.message || {}

  return (
    <BrainStepItem
      key={step.uuid}
      prefix={<CircleDotIcon className={tw({ layout: "size-4" })} />}
      title={summarizeAssistantMessage(message)}
      step={step}
    >
      {message.content?.map((block, index) => (
        <React.Fragment key={index}>
          {block.type === "text" && (
            <AssistantTextMessage text={block.text ?? ""} />
          )}
          {block.type === "thinking" && (
            <AssistantThinkingMessage thinking={block.thinking ?? ""} />
          )}
          {block.type === "tool_use" && (
            <AssistantToolUseMessage
              name={block.name ?? ""}
              input={block.input}
            />
          )}
        </React.Fragment>
      ))}
    </BrainStepItem>
  )
}

const summarizeUserMessage = (message: StreamOutputMessage) => {
  const results = (message.content || []).map((block) => {
    switch (block.type) {
      case "text":
        return "Text"
      case "tool_result":
        return "Tool Result"
      default:
        return block.type
    }
  })
  return results.join("; ")
}

const BrainUserStepItem: React.FC<{ step: BrainStep }> = ({ step }) => {
  const output = ClaudePayload.DecodeStreamOutput(step.data)
  if (!output || output.type !== "user") {
    return <>error: wrong type</>
  }

  const message = output.message || {}

  return (
    <BrainStepItem
      key={step.uuid}
      prefix={<CircleEllipsisIcon className={tw({ layout: "size-4" })} />}
      title={summarizeUserMessage(message)}
      step={step}
    >
      {message.content?.map((block, index) => (
        <React.Fragment key={index}>
          {block.type === "text" && <UserTextMessage text={block.text ?? ""} />}
          {block.type === "tool_result" && (
            <UserToolResultMessage
              content={block.content}
              isError={block.is_error ?? false}
            />
          )}
        </React.Fragment>
      ))}
    </BrainStepItem>
  )
}

const BrainResultStepItem: React.FC<{ step: BrainStep }> = ({ step }) => {
  const output = ClaudePayload.DecodeStreamOutput(step.data)
  if (!output || output.type !== "result") {
    return <>error: wrong type</>
  }
  return (
    <BrainStepItem
      key={step.uuid}
      prefix={<CircleCheckIcon className={tw({ layout: "size-4" })} />}
      title={"Result"}
      step={step}
    >
      <div
        className={tw({
          layout: "mt-3 p-3",
          appearance:
            "border-base-200 text-base-content rounded-lg border text-sm",
        })}
      >
        <MarkdownView
          content={output.result ?? ""}
          className={tw({
            layout: "[&>*:first-child]:mt-0 [&>*:last-child]:mb-0",
            appearance: "break-words",
          })}
          htmlClassNames={markdownClassNames}
        />
      </div>
    </BrainStepItem>
  )
}

const BrainUnknownStepItem: React.FC<{ step: BrainStep }> = ({ step }) => {
  return (
    <BrainStepItem
      key={step.uuid}
      prefix={<CircleIcon className={tw({ layout: "size-4" })} />}
      title={step.type}
      step={step}
    />
  )
}

const BrainStreamingStepItem: React.FC = () => {
  return (
    <div
      className={tw({
        layout: "mt-3 flex items-center justify-start gap-2 pt-3",
        appearance: "border-base-200 text-neutral border-t text-xs font-medium",
      })}
    >
      <LoaderCircleIcon
        className={tw({ layout: "size-4", appearance: "animate-spin" })}
      />
    </div>
  )
}

const BrainThreadPanel: React.FC<{
  thread: BrainThread
  onSend?: (
    topic: string,
    threadUuid: string,
    content: { text?: string },
  ) => void
  onFinish?: (pendingInputText: string) => void
}> = ({ thread, onSend, onFinish }) => {
  const [steps, setSteps] = React.useState<BrainStep[]>([])
  const [streaming, setStreaming] = React.useState(false)
  const [inputText, setInputText] = React.useState("")
  const startedRef = useRef(false)
  const inputRef = useRef<HTMLDivElement>(null)
  const inputVisibleRef = useRef(true)

  useEffect(() => {
    const initialSteps = [
      Protobuf.create(BrainStepSchema, {
        type: "input",
        uuid: crypto.randomUUID(),
        data: { text: thread.input.text },
      }),
      ...thread.steps,
    ]
    setSteps(initialSteps)
  }, [thread])

  // Track whether the input container is on screen so we only auto-scroll when
  // the user is already following the tail (not when they've scrolled up). The
  // container only exists while streaming, so re-run on `streaming` to attach
  // when it mounts and disconnect when it unmounts.
  useEffect(() => {
    const inputContainer = inputRef.current
    if (!inputContainer) {
      return
    }
    const observer = new IntersectionObserver(([entry]) => {
      inputVisibleRef.current = entry?.isIntersecting ?? false
    })
    observer.observe(inputContainer)
    return () => {
      observer.disconnect()
    }
  }, [streaming])

  // While the thread is alive, keep its end in view as new steps arrive — but
  // only if the user was already following the tail.
  useEffect(() => {
    if (streaming && inputVisibleRef.current) {
      inputRef.current?.scrollIntoView({ behavior: "smooth", block: "end" })
    }
  }, [steps, streaming])

  const start = useCallback(async () => {
    if (!thread.live || startedRef.current) {
      return
    }
    startedRef.current = true

    setStreaming(true)
    try {
      for await (const step of thread.live) {
        setSteps((prev) => [...prev, step])
      }
    } finally {
      setStreaming(false)
    }
  }, [thread.live])

  useEffect(() => {
    start()
  }, [start])

  // When a thread stops streaming, hand its still-pending draft back so it is
  // not lost. `streaming` is the only dependency: it flips false exactly once,
  // when the live loop ends, so this fires a single time with the draft from
  // that render. `startedRef` skips the initial not-yet-streaming render, and
  // keeping onFinish/inputText out of the deps stops the fresh inline onFinish
  // from re-firing it on every unrelated render.
  useEffect(() => {
    if (!startedRef.current || streaming) {
      return
    }
    onFinish?.(inputText)
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [streaming])

  const chain: React.ReactNode[] = useMemo(() => {
    const result: React.ReactNode[] = []
    for (let stepIndex = 0; stepIndex < steps.length; stepIndex++) {
      const step = steps[stepIndex]
      switch (step.type) {
        case "input": {
          result.push(<BrainInputView key={step.uuid} step={step} />)
          break
        }
        case "claudecli-input": {
          result.push(<BrainInputStepItem key={step.uuid} step={step} />)
          break
        }
        case "system": {
          let nextStepIndex = stepIndex + 1
          while (
            nextStepIndex < steps.length &&
            steps[nextStepIndex].type === "system"
          ) {
            nextStepIndex++
          }
          const group = steps.slice(stepIndex, nextStepIndex)
          result.push(<BrainSystemSteps key={group[0].uuid} steps={group} />)
          stepIndex = nextStepIndex - 1
          break
        }
        case "assistant":
          result.push(<BrainAssistantStepItem key={step.uuid} step={step} />)
          break
        case "user":
          result.push(<BrainUserStepItem key={step.uuid} step={step} />)
          break
        case "result":
          result.push(<BrainResultStepItem key={step.uuid} step={step} />)
          break
        default:
          result.push(<BrainUnknownStepItem key={step.uuid} step={step} />)
      }
    }
    return result
  }, [steps])

  return (
    <KurisuPanel title={thread.input.topic} subtitle={thread.input.threadUuid}>
      {chain}
      {streaming && <BrainStreamingStepItem />}
      {streaming && (
        <BrainThreadInput
          ref={inputRef}
          value={inputText}
          onChange={(e) => setInputText(e.target.value)}
          onSend={() => {
            setSteps((prev) => [
              ...prev,
              Protobuf.create(BrainStepSchema, {
                type: "input",
                uuid: crypto.randomUUID(),
                data: { text: inputText },
              }),
            ])
            onSend?.(thread.input.topic, thread.input.threadUuid, {
              text: inputText,
            })
            setInputText("")
          }}
        />
      )}
    </KurisuPanel>
  )
}

export default BrainThreadPanel
