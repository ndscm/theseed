import { SnackbarProvider } from "notistack"
import React from "react"
import { Outlet } from "react-router"

import CssBaseline from "@mui/material/CssBaseline"
import { ThemeProvider } from "@mui/material/styles"

import { LoginServiceProvider } from "../../../../../seed/cloud/login/client/tsx/login-service-context"
import Gotcha from "../../../../../seed/devprod/gotcha/tsx"
import { StuffServiceProvider } from "../../client/tsx/stuff-service-context"
import theme from "./theme"

const RootLayout: React.FC = () => {
  return (
    <ThemeProvider theme={theme}>
      <CssBaseline />
      <SnackbarProvider>
        <Gotcha />
        <LoginServiceProvider>
          <StuffServiceProvider>
            <Outlet />
          </StuffServiceProvider>
        </LoginServiceProvider>
      </SnackbarProvider>
    </ThemeProvider>
  )
}

export default RootLayout
