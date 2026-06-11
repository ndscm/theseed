import React from "react"
import { Outlet } from "react-router"

import TeamHeader from "../../../components/TeamHeader"

const TeamLayout: React.FC = () => {
  return (
    <>
      <TeamHeader />
      <Outlet />
    </>
  )
}

export default TeamLayout
