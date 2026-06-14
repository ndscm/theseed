import React from "react"
import { Outlet } from "react-router"

import { LoginServiceProvider } from "../../../../cloud/login/client/tsx/LoginServiceContext"
import Gotcha from "../../../../devprod/gotcha/tsx/Gotcha"
import { HooinDictateServiceProvider } from "../../../hooin/dictate/client/tsx/HooinDictateServiceContext"
import { HooinRosterServiceProvider } from "../../../hooin/roster/client/tsx/HooinRosterServiceContext"

const RootLayout: React.FC = () => {
  return (
    <LoginServiceProvider>
      <HooinDictateServiceProvider>
        <HooinRosterServiceProvider>
          <Gotcha />
          <Outlet />
        </HooinRosterServiceProvider>
      </HooinDictateServiceProvider>
    </LoginServiceProvider>
  )
}

export default RootLayout
