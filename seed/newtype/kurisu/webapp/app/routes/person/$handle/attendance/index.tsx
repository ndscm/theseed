import React, { useCallback, useEffect } from "react"
import { useTranslation } from "react-i18next"

import { ClipboardCopyIcon, UserKeyIcon, XIcon } from "lucide-react"

import tw from "../../../../../../../../devprod/ts/grouping-tailwind"
import { useKurisuService } from "../../../../../../client/tsx/KurisuServiceContext"

const PersonAttendancePage: React.FC<{ params: { handle: string } }> = ({
  params,
}) => {
  const { handle } = params
  const { t } = useTranslation("person")
  const kurisuService = useKurisuService()
  const [refreshToken, setRefreshToken] = React.useState<string>("")
  const [isRegenerating, setIsRegenerating] = React.useState<boolean>(false)
  const [isCopied, setIsCopied] = React.useState<boolean>(false)

  useEffect(() => {
    if (!refreshToken) {
      setIsCopied(false)
    }
  }, [refreshToken])

  const personHandle = handle
    .trim()
    .replace(/^@/, "")
    .replace(/@$/, "")
    .toLowerCase()
    .trim()

  const regenerateToken = useCallback(async () => {
    if (!kurisuService) {
      return
    }
    setRefreshToken("")
    setIsRegenerating(true)
    try {
      const response = await kurisuService.CreateSiliconJwt(personHandle)
      setRefreshToken(response.refreshToken)
    } finally {
      setIsRegenerating(false)
    }
  }, [kurisuService, personHandle])

  const copyRefreshToken = useCallback(async () => {
    if (refreshToken) {
      await navigator.clipboard.writeText(refreshToken)
      setIsCopied(true)
    }
  }, [refreshToken])

  return (
    <main className={tw({ layout: "min-h-0 flex-1 overflow-auto px-7 py-6" })}>
      <div className={tw({ layout: "mx-2 max-w-2xl" })}>
        <button
          className={tw({ component: "btn btn-primary btn-sm" })}
          onClick={regenerateToken}
          disabled={isRegenerating}
        >
          {isRegenerating ? (
            <span
              className={tw({
                component: "loading loading-spinner loading-xs",
              })}
            />
          ) : (
            <UserKeyIcon className={tw({ layout: "size-4" })} />
          )}
          {t("attendance.regenerateTokenButton", "Regenerate Token")}
        </button>
        {refreshToken && (
          <div
            className={tw({
              layout: "mt-4 flex flex-col",
              appearance: "bg-base-100 rounded p-4",
            })}
          >
            <div
              className={tw({
                layout: "font-mono break-all",
                appearance: "text-sm",
              })}
            >
              {refreshToken}
            </div>
            <div
              className={tw({
                layout: "mt-2 flex items-center justify-end gap-2",
              })}
            >
              {isCopied && (
                <span className={tw({ appearance: "text-neutral text-sm" })}>
                  {t("attendance.isCopiedHint", "Copied")}
                </span>
              )}
              <button
                className={tw({ component: "btn btn-ghost btn-sm" })}
                onClick={copyRefreshToken}
              >
                <ClipboardCopyIcon className={tw({ layout: "size-4" })} />
              </button>
              <button
                className={tw({ component: "btn btn-ghost btn-sm" })}
                onClick={() => setRefreshToken("")}
              >
                <XIcon className={tw({ layout: "size-4" })} />
              </button>
            </div>
          </div>
        )}
      </div>
    </main>
  )
}

export default PersonAttendancePage
