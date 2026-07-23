import React, { useCallback, useState } from "react"
import { useTranslation } from "react-i18next"

import { CornerDownLeftIcon, SendIcon } from "lucide-react"

import tw from "../../../../devprod/ts/grouping-tailwind"

// BrainThreadInput is the compact reply box rendered at the tail of a thread
// panel. It is a controlled input: the draft text lives in the parent panel,
// which also binds the topic and thread uuid before dispatching, so a reply
// lands under the same thread.
const BrainThreadInput: React.FC<{
  value?: string
  onChange?: React.ChangeEventHandler<HTMLTextAreaElement>
  onSend?: () => void
}> = ({ value, onChange, onSend }) => {
  const { t } = useTranslation("person")
  const [isMultilineEnabled, setIsMultilineEnabled] = useState(true)

  const sendThreadInput = useCallback(() => {
    const trimmed = value ? value.trim() : ""
    if (!trimmed) {
      return
    }
    onSend?.()
  }, [value, onSend])

  const handleKeyDown = useCallback(
    (e: React.KeyboardEvent<HTMLTextAreaElement>) => {
      if (e.key !== "Enter") {
        return
      }
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
        sendThreadInput()
        return
      }
      if (!isMultilineEnabled) {
        // Plain Enter sends when multi-line mode is off; otherwise it inserts
        // a carriage return via the default behavior.
        e.preventDefault()
        sendThreadInput()
      }
    },
    [sendThreadInput, isMultilineEnabled],
  )

  return (
    <div
      className={tw({
        layout: "mt-3 flex flex-col items-stretch pt-3",
        appearance: "border-base-200 border-t",
      })}
    >
      <div className={tw({ layout: "flex" })}>
        <textarea
          className={tw({
            component: "textarea textarea-ghost",
            layout:
              "field-sizing-content max-h-48 min-h-0 flex-1 resize-none overflow-y-auto",
            state: "focus:outline-none",
          })}
          rows={2}
          placeholder={t(
            "brain.threadInputPlaceholder",
            "Additional Brain Input",
          )}
          value={value}
          onChange={onChange}
          onKeyDown={handleKeyDown}
        />
      </div>
      <div className={tw({ layout: "flex justify-end gap-1" })}>
        <button
          type="button"
          className={tw(
            {
              component: "btn btn-square btn-ghost",
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
              : t("brain.multilineOffTooltip", "Use <Enter> to send messages")
          }
        >
          <CornerDownLeftIcon />
        </button>
        <button
          type="button"
          className={tw({
            component: "btn btn-square btn-primary btn-ghost",
            layout: "shrink-0",
          })}
          onClick={sendThreadInput}
          disabled={!value || !value.trim()}
        >
          <SendIcon />
        </button>
      </div>
    </div>
  )
}

export default BrainThreadInput
