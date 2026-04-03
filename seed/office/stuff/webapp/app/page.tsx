import "@univerjs/design/lib/index.css"
import "@univerjs/docs-ui/lib/index.css"
import "@univerjs/sheets-formula-ui/lib/index.css"
import "@univerjs/sheets-numfmt-ui/lib/index.css"
import "@univerjs/sheets-ui/lib/index.css"
import "@univerjs/ui/lib/index.css"

import {
  LocaleType,
  Univer,
  UniverInstanceType,
  mergeLocales,
} from "@univerjs/core"
import DesignEnUS from "@univerjs/design/locale/en-US"
import { UniverDocsPlugin } from "@univerjs/docs"
import { UniverDocsUIPlugin } from "@univerjs/docs-ui"
import DocsUIEnUS from "@univerjs/docs-ui/locale/en-US"
import { UniverFormulaEnginePlugin } from "@univerjs/engine-formula"
import { UniverRenderEnginePlugin } from "@univerjs/engine-render"
import { UniverSheetsPlugin } from "@univerjs/sheets"
import { UniverSheetsFormulaUIPlugin } from "@univerjs/sheets-formula-ui"
import SheetsFormulaUIEnUS from "@univerjs/sheets-formula-ui/locale/en-US"
import { UniverSheetsNumfmtUIPlugin } from "@univerjs/sheets-numfmt-ui"
import SheetsNumfmtUIEnUS from "@univerjs/sheets-numfmt-ui/locale/en-US"
import { UniverSheetsUIPlugin } from "@univerjs/sheets-ui"
import SheetsUIEnUS from "@univerjs/sheets-ui/locale/en-US"
import SheetsEnUS from "@univerjs/sheets/locale/en-US"
import { UniverUIPlugin } from "@univerjs/ui"
import UIEnUS from "@univerjs/ui/locale/en-US"
import React, { useEffect, useRef } from "react"
import { useTranslation } from "react-i18next"

import Box from "@mui/material/Box"
import Toolbar from "@mui/material/Toolbar"

const HomePage: React.FC = () => {
  const { t } = useTranslation("common")
  const univerDivRef = useRef<HTMLDivElement>(null)
  const univerInstanceRef = useRef<Univer>(null)

  useEffect(() => {
    void (async () => {
      if (!univerDivRef.current) {
        return
      }
      const univer = new Univer({
        locale: LocaleType.EN_US,
        locales: {
          [LocaleType.EN_US]: mergeLocales(
            DesignEnUS,
            UIEnUS,
            DocsUIEnUS,
            SheetsEnUS,
            SheetsUIEnUS,
            SheetsFormulaUIEnUS,
            SheetsNumfmtUIEnUS,
          ),
        },
      })

      univer.registerPlugin(UniverRenderEnginePlugin)
      univer.registerPlugin(UniverFormulaEnginePlugin)
      univer.registerPlugin(UniverUIPlugin, {
        container: univerDivRef.current,
      })
      univer.registerPlugin(UniverDocsPlugin)
      univer.registerPlugin(UniverDocsUIPlugin)
      univer.registerPlugin(UniverSheetsPlugin)
      univer.registerPlugin(UniverSheetsUIPlugin)
      univer.registerPlugin(UniverSheetsFormulaUIPlugin)
      univer.registerPlugin(UniverSheetsNumfmtUIPlugin)

      univer.createUnit(UniverInstanceType.UNIVER_SHEET, {})

      univerInstanceRef.current = univer
    })()

    return () => {
      if (univerInstanceRef.current) {
        univerInstanceRef.current.dispose()
        univerInstanceRef.current = null
      }
    }
  }, [])

  return (
    <Box sx={{ height: "100vh", display: "flex", flexDirection: "column" }}>
      <Toolbar />
      <Box ref={univerDivRef} sx={{ flexGrow: 1 }} />
    </Box>
  )
}

export default HomePage
