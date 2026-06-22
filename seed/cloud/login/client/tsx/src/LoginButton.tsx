import React from "react"

import { useLoginService } from "./LoginServiceContext"

const LoginButton: React.FC<
  {
    anonymous: React.ReactNode
    loading?: React.ReactNode
  } & React.ButtonHTMLAttributes<HTMLButtonElement>
> = ({ anonymous, loading, disabled, children, onClick, ...props }) => {
  const loginService = useLoginService()

  const onLoginClick = async () => {
    if (!loginService) {
      return
    }
    await loginService.login()
  }

  return (
    <button
      disabled={loginService?.loading || disabled}
      onClick={loginService?.current?.userUuid ? onClick : onLoginClick}
      {...props}
    >
      {loginService?.loading
        ? loading
        : loginService?.current?.userUuid
          ? children
          : anonymous}
    </button>
  )
}

export default LoginButton
