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
      <Container component="main" sx={{ py: 2 }}>
        <Typography variant="h4">{t("system.brand", "Stuff")}</Typography>
      </Container>
    </Box>
  )
}

export default HomePage
