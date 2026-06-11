import React, { useEffect } from "react"
import { useNavigate } from "react-router"

const PersonPage: React.FC<{ params: { handle: string } }> = ({ params }) => {
  const { handle } = params
  const navigate = useNavigate()

  useEffect(() => {
    navigate(`/person/${handle}/chat`)
  }, [])

  return <div />
}

export default PersonPage
