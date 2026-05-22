import React from "react"
import {
  Link as RouterLink,
  type LinkProps as RouterLinkProps,
} from "react-router"

import { type LinkProps } from "@mui/material/Link"
import { createTheme } from "@mui/material/styles"

const LinkBehavior = React.forwardRef<
  HTMLAnchorElement,
  Omit<RouterLinkProps, "to"> & { href: RouterLinkProps["to"] }
>((props, ref) => {
  const { href, ...other } = props
  return <RouterLink ref={ref} to={href} {...other} />
})

const StuffMuiTheme = createTheme({
  colorSchemes: {
    dark: true,
  },
  components: {
    MuiLink: {
      defaultProps: {
        component: LinkBehavior,
      } as LinkProps,
    },
    MuiButtonBase: {
      defaultProps: {
        LinkComponent: LinkBehavior,
      },
    },
  },
  palette: {
    primary: {
      main: "#e2732d",
    },
    secondary: {
      main: "#cc3554",
    },
  },
  typography: {
    h1: {
      fontSize: "3rem",
      fontWeight: 500,
      letterSpacing: "0",
    },
    h2: {
      fontSize: "2rem",
      fontWeight: 500,
      letterSpacing: "0",
    },
    h3: {
      fontSize: "1.5rem",
      fontWeight: 500,
      letterSpacing: "0",
    },
    h4: {
      fontSize: "1.2rem",
      fontWeight: 500,
      letterSpacing: "0",
    },
    h5: {
      fontSize: "1.2rem",
      fontWeight: 400,
      letterSpacing: "0",
    },
    h6: {
      fontSize: "1.2rem",
      fontWeight: 300,
      letterSpacing: "0",
    },
  },
})

export default StuffMuiTheme
