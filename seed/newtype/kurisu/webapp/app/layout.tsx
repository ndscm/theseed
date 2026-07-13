import React from "react"
import { Outlet } from "react-router"

import { LoginServiceProvider } from "../../../../cloud/login/client/tsx/LoginServiceContext"
import Gotcha from "../../../../devprod/gotcha/tsx/Gotcha"
import { HooinDictateServiceProvider } from "../../../hooin/dictate/client/tsx/HooinDictateServiceContext"
import { HooinInvadeServiceProvider } from "../../../hooin/invade/client/tsx/HooinInvadeServiceContext"
import { HooinRaidServiceProvider } from "../../../hooin/raid/client/tsx/HooinRaidServiceContext"
import { HooinRosterServiceProvider } from "../../../hooin/roster/client/tsx/HooinRosterServiceContext"
import { KurisuServiceProvider } from "../../client/tsx/KurisuServiceContext"

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
