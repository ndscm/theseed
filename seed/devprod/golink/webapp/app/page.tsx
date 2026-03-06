import React from "react"
import { useTranslation } from "react-i18next"

import Box from "@mui/material/Box"
import Container from "@mui/material/Container"
import Toolbar from "@mui/material/Toolbar"
import Typography from "@mui/material/Typography"

const HomePage: React.FC = () => {
  const { t } = useTranslation("common")

  return (
    <Box>
      <Toolbar />
      <Container component="main">
        <Typography variant="body1">
          {t("system.welcome", "A private short link service")}
        </Typography>
      </Container>
    </Box>
  )
}

export default HomePage
