import { SnackbarProvider } from "notistack"
import React from "react"
import { Outlet } from "react-router"

import CssBaseline from "@mui/material/CssBaseline"
import { ThemeProvider } from "@mui/material/styles"

import Gotcha from "../../../../../seed/devprod/gotcha/tsx/"
import { GolinkServiceProvider } from "../../client/tsx/golink-service-context"
import GolinkAppBar from "../containers/GolinkAppBar"
import theme from "./theme"

const RootLayout: React.FC = () => {
  return (
    <ThemeProvider theme={theme}>
      <CssBaseline />
      <SnackbarProvider>
        <GolinkServiceProvider>
          <Gotcha />
          <GolinkAppBar />
          <Outlet />
        </GolinkServiceProvider>
      </SnackbarProvider>
    </ThemeProvider>
  )
}

export default RootLayout
