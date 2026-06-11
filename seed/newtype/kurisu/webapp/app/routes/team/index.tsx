import React, { useEffect } from "react"
import { useNavigate } from "react-router"

const TeamPage: React.FC<{}> = ({}) => {
  let navigate = useNavigate()

  useEffect(() => {
    navigate("/team/members")
  })

  return <div />
}

export default TeamPage
