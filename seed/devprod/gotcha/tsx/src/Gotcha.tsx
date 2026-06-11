import React, { useEffect } from "react"

const ErrorBox: React.FC<{ error: Error; onClose: () => void }> = ({
  error,
  onClose,
}) => {
  return (
    <div
      style={{
        marginLeft: "8px",
        marginRight: "8px",
        marginBottom: "8px",
        padding: "8px",
        display: "flex",
        backgroundColor: "red",
        color: "white",
        pointerEvents: "auto",
      }}
    >
      <pre>{error.message}</pre>
      <button style={{ marginLeft: "8px" }} onClick={onClose}>
        X
      </button>
    </div>
  )
}

const Gotcha: React.FC<{}> = () => {
  const [errors, setErrors] = React.useState<Error[]>([])

  const pushError = (error: Error) => {
    setErrors((prev) => [...prev, error])
  }

  const removeError = (error: Error) => {
    setErrors((prev) => prev.filter((e) => e !== error))
  }

  useEffect(() => {
    const errorHandler = (ev: ErrorEvent) => {
      pushError(new Error(ev.message))
    }
    const promiseRejectionHandler = (ev: PromiseRejectionEvent) => {
      const err = ev.reason
      if (err instanceof Error) {
        pushError(err)
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
  }, [])

  return errors.length > 0 ? (
    <div
      style={{
        position: "fixed",
        bottom: 0,
        left: 0,
        right: 0,
        zIndex: 9999,
        display: "flex",
        flexDirection: "column",
        alignItems: "flex-start",
        pointerEvents: "none",
      }}
    >
      {errors.map((error, index) => (
        <ErrorBox
          key={index}
          error={error}
          onClose={() => removeError(error)}
        />
      ))}
      <button
        style={{
          margin: "8px",
          padding: "8px",
          backgroundColor: "gray",
          color: "white",
          pointerEvents: "auto",
        }}
        onClick={() => setErrors([])}
      >
        X
      </button>
    </div>
  ) : null
}

export default Gotcha
