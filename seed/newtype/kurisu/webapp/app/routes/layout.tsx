import React from "react"
import { Outlet } from "react-router"

import { LoginServiceProvider } from "../../../../../cloud/login/client/tsx/LoginServiceContext"
import Gotcha from "../../../../../devprod/gotcha/tsx/Gotcha"
import tw from "../../../../../devprod/ts/grouping-tailwind"
import { HooinDictateServiceProvider } from "../../../../hooin/dictate/client/tsx/HooinDictateServiceContext"
import { HooinRosterServiceProvider } from "../../../../hooin/roster/client/tsx/HooinRosterServiceContext"
import KurisuSideBar from "../../components/KurisuSideBar"

const RootLayout: React.FC = () => {
  return (
    <LoginServiceProvider>
      <HooinDictateServiceProvider>
        <HooinRosterServiceProvider>
          <Gotcha />
          <div
            className={tw({
              component: "drawer",
              state: "lg:drawer-open",
            })}
          >
            <input
              id="kurisu-sidebar-toggle"
              type="checkbox"
              className={tw({ component: "drawer-toggle" })}
            />
            <div
              className={tw({
                component: "drawer-side",
                state: "is-drawer-close:overflow-visible",
              })}
            >
              <label
                htmlFor="kurisu-sidebar-toggle"
                className={tw({ component: "drawer-overlay" })}
              />
              <KurisuSideBar />
            </div>
            <div
              className={tw({
                component: "drawer-content",
                layout: "flex h-dvh min-h-0 flex-col",
              })}
            >
              <Outlet />
            </div>
          </div>
        </HooinRosterServiceProvider>
      </HooinDictateServiceProvider>
    </LoginServiceProvider>
  )
}

export default RootLayout
