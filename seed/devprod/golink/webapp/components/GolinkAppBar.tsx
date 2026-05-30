import React, { useEffect } from "react"
import { useTranslation } from "react-i18next"

import AppBar from "@mui/material/AppBar"
import Box from "@mui/material/Box"
import Button from "@mui/material/Button"
import CircularProgress from "@mui/material/CircularProgress"
import Link from "@mui/material/Link"
import Toolbar from "@mui/material/Toolbar"
import Typography from "@mui/material/Typography"

import LoginButton from "../../../../cloud/login/client/tsx/LoginButton"
import { useLoginService } from "../../../../cloud/login/client/tsx/LoginServiceContext"
import { type LoginStatus } from "../../../../cloud/login/proto/login_pb"

const GolinkAppBar: React.FC = () => {
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
        <Box>
          <Button
            component={(props) => (
              <LoginButton
                anonymous={t("auth.loginButton", "Login")}
                loading={
                  <CircularProgress color="inherit" aria-label="Loading…" />
                }
                {...props}
              />
            )}
            color="inherit"
            sx={{
              display: "flex",
              flexDirection: "column",
              alignItems: "start",
              textTransform: "none",
            }}
            onClick={() => {
              console.info("Login status: ", login)
            }}
          >
            <Typography variant="body1">{login?.displayName}</Typography>
            <Typography variant="body2">{login?.userHandle}@</Typography>
          </Button>
        </Box>
      </Toolbar>
    </AppBar>
  )
}

export default GolinkAppBar
