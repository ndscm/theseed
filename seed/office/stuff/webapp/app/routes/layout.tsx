import { SnackbarProvider } from "notistack"
import React from "react"
import { Outlet } from "react-router"

import CssBaseline from "@mui/material/CssBaseline"
import { ThemeProvider } from "@mui/material/styles"

import { LoginServiceProvider } from "../../../../../cloud/login/client/tsx/LoginServiceContext.js"
import Gotcha from "../../../../../devprod/gotcha/tsx/Gotcha.js"
import { StuffServiceProvider } from "../../../client/tsx/StuffServiceContext.js"
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
