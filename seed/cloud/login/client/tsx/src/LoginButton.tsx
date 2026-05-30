import React from "react"

import { useLoginService } from "./LoginServiceContext"

const LoginButton: React.FC<
  {
    anonymous: React.ReactNode
    loading?: React.ReactNode
  } & React.ButtonHTMLAttributes<HTMLButtonElement>
> = ({ anonymous, loading, disabled, children, onClick, ...props }) => {
  const loginService = useLoginService()
  const [redirecting, setRedirecting] = React.useState(false)

  const onLoginClick = async () => {
    if (redirecting) {
      return
    }
    setRedirecting(true)
    window.location.href = `${window.location.origin}/auth/login?return=${encodeURIComponent(window.location.href)}`
  }

  return (
    <button
      disabled={loginService?.loading || redirecting || disabled}
      onClick={loginService?.current?.userUuid ? onClick : onLoginClick}
      {...props}
    >
      {loginService?.loading || redirecting
        ? loading
        : loginService?.current?.userUuid
          ? children
          : anonymous}
    </button>
  )
}

export default LoginButton
