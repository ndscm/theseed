import React from "react"
import { useTranslation } from "react-i18next"

import AppBar from "@mui/material/AppBar"
import Box from "@mui/material/Box"
import Link from "@mui/material/Link"
import Toolbar from "@mui/material/Toolbar"
import Typography from "@mui/material/Typography"

const GolinkAppBar: React.FC = () => {
  const { t } = useTranslation("common")

  return (
    <AppBar position="fixed" elevation={1}>
      <Toolbar sx={{ display: "flex" }}>
        <Box sx={{ flex: 1, display: "flex", alignItems: "center" }}>
          <Link
            sx={{ display: "flex", alignItems: "center" }}
            underline="none"
            href="/"
          >
            <Typography
              variant="h6"
              component="span"
              color="white"
              sx={{ marginLeft: 1 }}
            >
              {t("system.brand", "go/link")}
            </Typography>
          </Link>
          <Typography
            variant="body1"
            sx={{
              marginLeft: 2,
              whiteSpace: "nowrap",
            }}
          >
            {t("system.welcome", "A private short link service")}
          </Typography>
        </Box>
      </Toolbar>
    </AppBar>
  )
}

export default GolinkAppBar
