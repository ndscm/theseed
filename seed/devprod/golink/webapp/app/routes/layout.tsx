import { SnackbarProvider } from "notistack"
import React from "react"
import { Outlet } from "react-router"

import CssBaseline from "@mui/material/CssBaseline"
import { ThemeProvider } from "@mui/material/styles"

import { LoginServiceProvider } from "@//seed/cloud/login/client/tsx/LoginServiceContext"
import { GolinkServiceProvider } from "@//seed/devprod/golink/client/tsx/golink-service-context"
import Gotcha from "@//seed/devprod/gotcha/tsx/Gotcha"

import GolinkAppBar from "../../components/GolinkAppBar"
import theme from "./theme"

const RootLayout: React.FC = () => {
  return (
    <ThemeProvider theme={theme}>
      <CssBaseline />
      <SnackbarProvider>
        <LoginServiceProvider>
          <GolinkServiceProvider>
            <Gotcha />
            <GolinkAppBar />
            <Outlet />
          </GolinkServiceProvider>
        </LoginServiceProvider>
      </SnackbarProvider>
    </ThemeProvider>
  )
}

export default RootLayout
