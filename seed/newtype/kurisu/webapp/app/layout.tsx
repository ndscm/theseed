import React from "react"
import { Outlet } from "react-router"

import { LoginServiceProvider } from "../../../../cloud/login/client/tsx/LoginServiceContext.js"
import Gotcha from "../../../../devprod/gotcha/tsx/Gotcha.js"
import { HooinDictateServiceProvider } from "../../../hooin/dictate/client/tsx/HooinDictateServiceContext.js"
import { HooinInvadeServiceProvider } from "../../../hooin/invade/client/tsx/HooinInvadeServiceContext.js"
import { HooinRaidServiceProvider } from "../../../hooin/raid/client/tsx/HooinRaidServiceContext.js"
import { HooinRosterServiceProvider } from "../../../hooin/roster/client/tsx/HooinRosterServiceContext.js"
import { KurisuServiceProvider } from "../../client/tsx/KurisuServiceContext.js"

const RootLayout: React.FC = () => {
  return (
    <LoginServiceProvider>
      <HooinDictateServiceProvider>
        <HooinInvadeServiceProvider>
          <HooinRaidServiceProvider>
            <HooinRosterServiceProvider>
              <KurisuServiceProvider>
                <Gotcha />
                <Outlet />
              </KurisuServiceProvider>
            </HooinRosterServiceProvider>
          </HooinRaidServiceProvider>
        </HooinInvadeServiceProvider>
      </HooinDictateServiceProvider>
    </LoginServiceProvider>
  )
}

export default RootLayout
