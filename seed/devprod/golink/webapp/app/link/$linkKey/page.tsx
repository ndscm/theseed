import React from "react"

import Box from "@mui/material/Box"
import Container from "@mui/material/Container"
import Toolbar from "@mui/material/Toolbar"

const LinkPage: React.FC<{ params: { linkKey: string } }> = ({ params }) => {
  const { linkKey } = params

  return (
    <Box>
      <Toolbar />
      <Container component="main">{linkKey}</Container>
    </Box>
  )
}

export default LinkPage
