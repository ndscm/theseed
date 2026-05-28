import { SnackbarProvider } from "notistack"
import React from "react"
import { Outlet } from "react-router"

import CssBaseline from "@mui/material/CssBaseline"
import { ThemeProvider } from "@mui/material/styles"

import { LoginServiceProvider } from "../../../../../cloud/login/client/tsx/LoginServiceContext"
import Gotcha from "../../../../gotcha/tsx/Gotcha"
import { GolinkServiceProvider } from "../../../client/tsx/GolinkServiceContext"
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
