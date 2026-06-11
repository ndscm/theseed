import React, { useEffect } from "react"
import { useNavigate } from "react-router"

const PersonListPage: React.FC<{}> = ({}) => {
  const navigate = useNavigate()

  useEffect(() => {
    navigate(`/team/members`)
  })

  return <div />
}

export default PersonListPage
