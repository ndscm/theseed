import { SnackbarProvider } from "notistack"
import React from "react"
import { Outlet } from "react-router"

import CssBaseline from "@mui/material/CssBaseline"
import { ThemeProvider } from "@mui/material/styles"

import { LoginServiceProvider } from "@//seed/cloud/login/client/tsx/LoginServiceContext"
import Gotcha from "@//seed/devprod/gotcha/tsx/Gotcha"
import { StuffServiceProvider } from "@//seed/office/stuff/client/tsx/StuffServiceContext"

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
