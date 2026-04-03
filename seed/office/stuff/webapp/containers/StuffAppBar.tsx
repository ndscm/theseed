import React, { useEffect } from "react"
import { useTranslation } from "react-i18next"

import AppBar from "@mui/material/AppBar"
import Box from "@mui/material/Box"
import Button from "@mui/material/Button"
import Link from "@mui/material/Link"
import Toolbar from "@mui/material/Toolbar"
import Typography from "@mui/material/Typography"

import { type LoginStatus } from "../../../../../seed/cloud/login/proto/login_pb"
import { useLoginService } from "../../../../../seed/cloud/login/client/tsx/login-service-context"

const StuffAppBar: React.FC = () => {
  const { t } = useTranslation("common")
  const loginService = useLoginService()
  const [login, setLogin] = React.useState<LoginStatus>()

  useEffect(() => {
    void (async () => {
      if (!loginService) {
        return
      }
      const currentLogin = await loginService.GetLoginStatus()
      setLogin(currentLogin)
    })()
  }, [loginService])

  const onLoginClick = async () => {
    window.location.href = `${window.location.origin}/auth/login?return=${encodeURIComponent(window.location.href)}`
  }

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
              {t("system.brand", "Stuff")}
            </Typography>
          </Link>
          <Typography
            variant="body1"
            sx={{
              marginLeft: 2,
              whiteSpace: "nowrap",
            }}
          >
            {t("system.welcome", "Physical asset management")}
          </Typography>
        </Box>
        <Box>
          {login?.userUuid ? (
            <Box
              sx={{
                display: "flex",
                flexDirection: "column",
                alignItems: "start",
              }}
            >
              <Button
                color="inherit"
                sx={{
                  display: "flex",
                  flexDirection: "column",
                  alignItems: "start",
                  textTransform: "none",
                }}
              >
                <Typography variant="body1">{login.displayName}</Typography>
                <Typography variant="body2">@{login.userHandle}</Typography>
              </Button>
            </Box>
          ) : (
            <Button
              color="inherit"
              onClick={async () => {
                await onLoginClick()
              }}
            >
              {t("auth.loginButton", "Login")}
            </Button>
          )}
        </Box>
      </Toolbar>
    </AppBar>
  )
}

export default StuffAppBar
