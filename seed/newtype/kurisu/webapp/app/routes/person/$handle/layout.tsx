import React from "react"
import { Outlet } from "react-router"

import PersonHeader from "../../../../components/PersonHeader"

const PersonLayout: React.FC = () => {
  return (
    <>
      <PersonHeader />
      <Outlet />
    </>
  )
}

export default PersonLayout
