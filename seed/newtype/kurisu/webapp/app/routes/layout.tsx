import React from "react"
import { Outlet } from "react-router"

import tw from "../../../../../devprod/ts/grouping-tailwind"
import KurisuSideBar from "../../components/KurisuSideBar"

const HomeLayout: React.FC = () => {
  return (
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
  )
}

export default HomeLayout
