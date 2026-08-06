import * as Protobuf from "@bufbuild/protobuf"
import * as ProtobufWkt from "@bufbuild/protobuf/wkt"
import React, { useEffect, useMemo, useRef } from "react"
import { useTranslation } from "react-i18next"

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
  type BrainStep,
  BrainStepSchema,
} from "../../../gajetto/proto/brain_pb"
import BrainThreadInput from "./BrainThreadInput"
import KurisuPanel from "./KurisuPanel"

export type BrainThread = {
  threadUuid: string
  emitTime: ProtobufWkt.Timestamp
  topic: string
  steps: { [stepUuid: string]: BrainStep | undefined }
}

const getStepEmitTime = (step: BrainStep): number => {
  const emitTime = step.emitTime
  return emitTime ? ProtobufWkt.timestampMs(emitTime) : 0
}

// A thread whose last non-result step is older than this is treated as no
// longer streaming: an active run emits steps continuously, so a trailing
// non-result step this stale is almost certainly an interrupted run rather
// than a live one.
const STREAMING_STALE_MS = 10 * 60 * 1000

const sortStepsByEmitTime = (steps: {
  [stepUuid: string]: BrainStep | undefined
}): BrainStep[] =>
  Object.values(steps)
    .filter((step): step is BrainStep => step !== undefined)
    .sort((a, b) => getStepEmitTime(a) - getStepEmitTime(b))

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

const AssistantToolUseBash: React.FC<{
  description: string
  command: string
  unknown?: Record<string, unknown>
}> = ({ description, command, unknown }) => {
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
        {description}
      </span>
      <pre
        className={tw({
          layout: "m-0 mt-2 overflow-x-auto",
          appearance: "text-neutral font-mono text-xs",
        })}
      >
        {command}
      </pre>
      {unknown && Object.keys(unknown).length > 0 && (
        <pre
          className={tw({
            layout: "m-0 mt-2 overflow-x-auto",
            appearance: "text-neutral font-mono text-xs",
          })}
        >
          {JSON.stringify(unknown ?? {}, null, 2)}
        </pre>
      )}
    </div>
  )
}

const AssistantToolUseRead: React.FC<{
  filePath: string
  unknown?: Record<string, unknown>
}> = ({ filePath, unknown }) => {
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
        {filePath}
      </span>
      {unknown && Object.keys(unknown).length > 0 && (
        <pre
          className={tw({
            layout: "m-0 mt-2 overflow-x-auto",
            appearance: "text-neutral font-mono text-xs",
          })}
        >
          {JSON.stringify(unknown ?? {}, null, 2)}
        </pre>
      )}
    </div>
  )
}

const AssistantToolUseWrite: React.FC<{
  filePath: string
  content: string
  unknown?: Record<string, unknown>
}> = ({ filePath, content, unknown }) => {
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
        {filePath}
      </span>
      <pre
        className={tw({
          layout: "m-0 mt-2 overflow-x-auto",
          appearance: "text-neutral font-mono text-xs",
        })}
      >
        {content}
      </pre>
      {unknown && Object.keys(unknown).length > 0 && (
        <pre
          className={tw({
            layout: "m-0 mt-2 overflow-x-auto",
            appearance: "text-neutral font-mono text-xs",
          })}
        >
          {JSON.stringify(unknown ?? {}, null, 2)}
        </pre>
      )}
    </div>
  )
}

const AssistantToolUseMessage: React.FC<{ name: string; input: unknown }> = ({
  name,
  input,
}) => {
  const inputObject =
    typeof input === "object" && input !== null
      ? (input as Record<string, unknown>)
      : {}

  switch (name.trim()) {
    case "Bash": {
      const { description, command, ...unknown } = inputObject as {
        description?: string
        command?: string
        [key: string]: unknown
      }
      return (
        <AssistantToolUseBash
          description={description ?? "Bash"}
          command={command ?? ""}
          unknown={unknown}
        />
      )
    }
    case "Read": {
      const { file_path, ...unknown } = inputObject as {
        file_path?: string
        [key: string]: unknown
      }
      return (
        <AssistantToolUseRead
          filePath={file_path ?? "Read"}
          unknown={unknown}
        />
      )
    }
    case "Write": {
      const { file_path, content, ...unknown } = inputObject as {
        file_path?: string
        content?: string
        [key: string]: unknown
      }
      return (
        <AssistantToolUseWrite
          filePath={file_path ?? "Write"}
          content={content ?? ""}
          unknown={unknown}
        />
      )
    }
  }

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
      {children}
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
    ></BrainStepItem>
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
        layout: "flex items-center justify-start gap-2 pt-3 pb-3",
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
}> = ({ thread, onSend }) => {
  const { t } = useTranslation("person")
  const { steps } = thread
  const [inputText, setInputText] = React.useState("")
  const inputRef = useRef<HTMLDivElement>(null)
  const inputVisibleRef = useRef(true)

  // steps is a uuid-keyed map; render in emit-time order.
  const sortedSteps = useMemo(() => sortStepsByEmitTime(steps), [steps])

  // NOTE: There is no actual liveness signal available here — this panel
  // only sees the steps it has been handed, not whether the thread is
  // still receiving from the live subscription. So "streaming" is inferred
  // purely from the last step's type: anything other than a terminal
  // result is treated as still in progress. This is a heuristic, and it is
  // wrong for a thread whose last known step legitimately isn't a result —
  // e.g. an interrupted/abandoned run, or a history thread whose trailing
  // result page hasn't loaded yet — where it shows a perpetual streaming
  // spinner (and keeps the input in tail-follow mode). Fixing this
  // properly requires threading real liveness state down from the live
  // subscription instead of guessing from the last step type.
  //
  // As a bound on the false positives above, also treat the thread as not
  // streaming once its last step is older than STREAMING_STALE_MS: a live
  // run emits steps continuously, so a long-idle trailing non-result step
  // is far more likely an interrupted/abandoned run than an active one.
  const streaming = useMemo(() => {
    const lastStep = sortedSteps[sortedSteps.length - 1]
    if (!lastStep) {
      return false
    }
    if (lastStep.type === "result" || lastStep.type === "claudecli-result") {
      return false
    }
    return Date.now() - getStepEmitTime(lastStep) < STREAMING_STALE_MS
  }, [sortedSteps])

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
  }, [sortedSteps, streaming])

  const chain: React.ReactNode[] = useMemo(() => {
    const result: React.ReactNode[] = []
    for (let stepIndex = 0; stepIndex < sortedSteps.length; stepIndex++) {
      const step = sortedSteps[stepIndex]
      switch (step.type) {
        case "input": {
          result.push(<BrainInputView key={step.uuid} step={step} />)
          break
        }
        case "claudecli-input": {
          result.push(<BrainInputStepItem key={step.uuid} step={step} />)
          break
        }
        case "claudecli-system":
        case "system": {
          let nextStepIndex = stepIndex + 1
          while (
            nextStepIndex < sortedSteps.length &&
            (sortedSteps[nextStepIndex].type === "system" ||
              sortedSteps[nextStepIndex].type === "claudecli-system")
          ) {
            nextStepIndex++
          }
          const group = sortedSteps.slice(stepIndex, nextStepIndex)
          result.push(<BrainSystemSteps key={group[0].uuid} steps={group} />)
          stepIndex = nextStepIndex - 1
          break
        }
        case "claudecli-assistant":
        case "assistant": {
          result.push(<BrainAssistantStepItem key={step.uuid} step={step} />)
          break
        }
        case "claudecli-user":
        case "user": {
          result.push(<BrainUserStepItem key={step.uuid} step={step} />)
          break
        }
        case "claudecli-result":
        case "result": {
          result.push(<BrainResultStepItem key={step.uuid} step={step} />)
          break
        }
        default: {
          result.push(<BrainUnknownStepItem key={step.uuid} step={step} />)
          break
        }
      }
    }
    return result
  }, [sortedSteps])

  return (
    <KurisuPanel title={thread.topic} subtitle={thread.threadUuid}>
      <div
        className={tw({
          layout: "pb-3",
          appearance: "text-neutral text-xs font-semibold uppercase",
        })}
      >
        {t("brain.threadStepsTitle", "Steps")}
      </div>
      {chain}
      {streaming && <BrainStreamingStepItem />}
      {!!(streaming || inputText) && (
        <BrainThreadInput
          ref={inputRef}
          value={inputText}
          onChange={(e) => setInputText(e.target.value)}
          onSend={() => {
            onSend?.(thread.topic, thread.threadUuid, {
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
