import { useSnackbar } from "notistack"
import React, { useEffect } from "react"

const Gotcha: React.FC<{}> = () => {
  const snackbar = useSnackbar()

  useEffect(() => {
    const errorHandler = (ev: ErrorEvent) => {
      snackbar.enqueueSnackbar(ev.message, { variant: "error" })
    }
    const promiseRejectionHandler = (ev: PromiseRejectionEvent) => {
      const err = ev.reason
      if (err instanceof Error) {
        snackbar.enqueueSnackbar(err.message, { variant: "error" })
      }
    }
    if (typeof window != "undefined") {
      window.addEventListener("error", errorHandler)
      window.addEventListener("unhandledrejection", promiseRejectionHandler)
    }
    return () => {
      if (typeof window != "undefined") {
        window.removeEventListener("error", errorHandler)
        window.removeEventListener(
          "unhandledrejection",
          promiseRejectionHandler,
        )
      }
    }
  }, [snackbar])

  return null
}

export default Gotcha
