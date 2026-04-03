import React from "react"
import { Outlet } from "react-router"

import StuffAppBar from "../../containers/StuffAppBar"

const StuffAppBarLayout: React.FC = () => {
  return (
    <>
      <StuffAppBar />
      <Outlet />
    </>
  )
}

export default StuffAppBarLayout
