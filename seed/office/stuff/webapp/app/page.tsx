import "@univerjs/design/lib/index.css"
import "@univerjs/docs-ui/lib/index.css"
import "@univerjs/sheets-formula-ui/lib/index.css"
import "@univerjs/sheets-numfmt-ui/lib/index.css"
import "@univerjs/sheets-ui/lib/index.css"
import "@univerjs/ui/lib/index.css"

import * as Protobuf from "@bufbuild/protobuf"
import {
  type CellValue,
  CommandType,
  type ICellData,
  type IObjectMatrixPrimitiveType,
  type IWorkbookData,
  LocaleType,
  type Nullable,
  Univer,
  UniverInstanceType,
  mergeLocales,
} from "@univerjs/core"
import { FUniver } from "@univerjs/core/facade"
import DesignEnUS from "@univerjs/design/locale/en-US"
import { UniverDocsPlugin } from "@univerjs/docs"
import { UniverDocsUIPlugin } from "@univerjs/docs-ui"
import DocsUIEnUS from "@univerjs/docs-ui/locale/en-US"
import { UniverFormulaEnginePlugin } from "@univerjs/engine-formula"
import { UniverRenderEnginePlugin } from "@univerjs/engine-render"
import {
  COMMAND_LISTENER_SKELETON_CHANGE,
  COMMAND_LISTENER_VALUE_CHANGE,
  type CommandListenerSkeletonChange,
  type CommandListenerValueChange,
  SheetSkeletonChangeType,
  SheetValueChangeType,
  UniverSheetsPlugin,
} from "@univerjs/sheets"
import { UniverSheetsFormulaUIPlugin } from "@univerjs/sheets-formula-ui"
import SheetsFormulaUIEnUS from "@univerjs/sheets-formula-ui/locale/en-US"
import { UniverSheetsNumfmtUIPlugin } from "@univerjs/sheets-numfmt-ui"
import SheetsNumfmtUIEnUS from "@univerjs/sheets-numfmt-ui/locale/en-US"
import { UniverSheetsUIPlugin } from "@univerjs/sheets-ui"
import SheetsUIEnUS from "@univerjs/sheets-ui/locale/en-US"
import { type FWorksheet } from "@univerjs/sheets/facade"
import SheetsEnUS from "@univerjs/sheets/locale/en-US"
import { UniverUIPlugin } from "@univerjs/ui"
import UIEnUS from "@univerjs/ui/locale/en-US"
import * as Fractional from "fractional-indexing"
import React, { useCallback, useEffect, useRef, useState } from "react"
import { useTranslation } from "react-i18next"

import Box from "@mui/material/Box"
import CircularProgress from "@mui/material/CircularProgress"
import Toolbar from "@mui/material/Toolbar"
import Typography from "@mui/material/Typography"

import { StuffSchema } from "../../proto/stuff_pb"
import { useStuffService } from "../../client/tsx/stuff-service-context"

import "@univerjs/sheets/facade"
import "@univerjs/sheets-ui/facade"

const BASIC_COLUMNS = [
  "identifier",
  "primaryCategory",
  "secondaryCategory",
  "fixedAsset",
  "displayName",
  "model",
  "brand",
  "purchaseDate",
  "source",
  "originalUsage",
  "notes",
]

const HomePage: React.FC = () => {
  const { t } = useTranslation("common")
  const stuffService = useStuffService()
  const univerDivRef = useRef<HTMLDivElement>(null)
  const univerInstanceRef = useRef<Univer>(null)
  const univerApiRef = useRef<FUniver>(null)
  const [selectedCellMeta, setSelectedCellMeta] = useState<{
    uuid?: string
    order?: string
    field?: string
  }>({})

  const assignOrders = useCallback(
    (
      sheet: FWorksheet,
      cells: Nullable<CellValue>[][],
      affectedRows: number[],
    ): { [rowIndex: string]: string } => {
      const sortedRows = [...affectedRows].sort((a, b) => a - b)
      const result: { [rowIndex: string]: string } = {}

      // Group affected rows into contiguous runs so we can generate
      // fractional keys between the boundaries of each run.
      const groups: number[][] = []
      for (const rowIndex of sortedRows) {
        const last = groups[groups.length - 1]
        if (last && rowIndex === last[last.length - 1] + 1) {
          last.push(rowIndex)
        } else {
          groups.push([rowIndex])
        }
      }

      for (const group of groups) {
        const firstRow = group[0]
        const lastRow = group[group.length - 1]

        // Find lower bound: nearest row before the group with an existing order
        let lower: string | null = null
        for (let i = firstRow - 1; i >= 0; i--) {
          const order = sheet.getRowCustomMetadata(i)?.order
          if (typeof order === "string" && order) {
            lower = order
            break
          }
        }

        // Find upper bound: nearest row after the group with an existing order
        let upper: string | null = null
        for (let i = lastRow + 1; i <= cells.length - 1; i++) {
          const order = sheet.getRowCustomMetadata(i)?.order
          if (typeof order === "string" && order) {
            upper = order
            break
          }
        }

        // Generate N keys between lower and upper for the group
        const keys = Fractional.generateNKeysBetween(lower, upper, group.length)
        for (let i = 0; i < group.length; i++) {
          result[group[i]] = keys[i]
        }
      }

      return result
    },
    [],
  )

  const createOrUpdateRow = useCallback(
    async (
      sheet: FWorksheet,
      cells: Nullable<CellValue>[][],
      rowIndex: number,
      newOrder: string,
    ) => {
      if (!stuffService) {
        return
      }
      const metadata = sheet.getRowCustomMetadata(rowIndex)
      const uuid = metadata?.uuid
      const dataObject: { [field: string]: string } = {}
      for (const [colKey, cell] of Object.entries(cells[rowIndex] || {})) {
        const colIndex = Number(colKey)
        const colMetadata = sheet.getColumnCustomMetadata(colIndex)
        const field = colMetadata?.field
        if (typeof field === "string" && field) {
          switch (typeof cell) {
            case "string":
              dataObject[field] = cell
              break
            case "number":
              dataObject[field] = cell.toString()
              break
            case "boolean":
              dataObject[field] = cell ? "true" : "false"
              break
            default:
              break
          }
        }
      }

      if (uuid) {
        // Update existing stuff
        console.info(
          "Updating stuff with uuid:",
          uuid,
          "order:",
          newOrder,
          "data:",
          dataObject,
        )
        const stuff = Protobuf.create(StuffSchema, {
          uuid: String(uuid),
          order: newOrder,
          data: JSON.stringify(dataObject),
        })
        const updated = await stuffService.UpdateStuff(stuff)
        sheet.setRowCustomMetadata(rowIndex, {
          uuid: String(uuid),
          order: updated.order,
        })
      } else {
        // Create new stuff
        console.info(
          "Creating new stuff with order:",
          newOrder,
          "data:",
          dataObject,
        )
        const stuff = Protobuf.create(StuffSchema, {
          order: newOrder,
          data: JSON.stringify(dataObject),
        })
        const created = await stuffService.CreateStuff(stuff)
        sheet.setRowCustomMetadata(rowIndex, {
          uuid: created.uuid,
          order: created.order,
        })
      }
    },
    [stuffService],
  )

  const removeStuff = useCallback(
    async (sheet: FWorksheet, rowIndex: number) => {
      if (!stuffService) {
        return
      }
      const uuid = sheet.getRowCustomMetadata(rowIndex)?.uuid
      if (typeof uuid === "string" && uuid) {
        console.info("Deleting stuff with uuid:", uuid)
        await stuffService.DeleteStuff(uuid)
      }
    },
    [stuffService],
  )

  useEffect(() => {
    let disposed = false
    void (async () => {
      if (!univerDivRef.current) {
        return
      }
      if (!stuffService) {
        return
      }
      const response = await stuffService.ListStuff()

      // Sort by order field ascending so row order is stable
      const sortedStuff = [...response.stuff].sort((a, b) => {
        return a.order < b.order ? -1 : a.order > b.order ? 1 : 0
      })
      console.info("Load stuff:", sortedStuff)

      // Build initial cellData and rowData for the workbook
      const cellData: IObjectMatrixPrimitiveType<ICellData> = {}
      const rowData: Record<
        number,
        { custom: { uuid: string; order: string } }
      > = {}
      const columnData: Record<number, { custom: { field: string } }> = {}

      const columns = [...BASIC_COLUMNS]
      for (const [colIndex, field] of columns.entries()) {
        columnData[colIndex] = { custom: { field } }
      }
      const columnMapping: Record<string, number> = {}
      columns.forEach((field, index) => {
        columnMapping[field] = index
      })

      sortedStuff.forEach((stuff, rowIndex) => {
        const dataObject: { [field: string]: string } = {}
        if (stuff.data) {
          try {
            const parsed = JSON.parse(stuff.data)
            Object.entries(parsed).forEach(([field, value]) => {
              dataObject[field] = String(value)
            })
          } catch (err) {
            console.error("Failed to parse data:", stuff.data, err)
          }
        }
        rowData[rowIndex] = {
          custom: { uuid: stuff.uuid, order: stuff.order },
        }
        Object.entries(dataObject).forEach(([field, value]) => {
          if (typeof columnMapping[field] !== "number") {
            columnData[columns.length] = { custom: { field } }
            columnMapping[field] = columns.length
            columns.push(field)
          }
          const colIndex = columnMapping[field]
          if (!cellData[rowIndex]) {
            cellData[rowIndex] = {}
          }
          cellData[rowIndex][colIndex] = { v: value }
        })
      })

      if (disposed) {
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

      const workbookData: Partial<IWorkbookData> = {
        sheets: {
          stuff: {
            id: "stuff",
            name: "stuff",
            cellData,
            rowData,
            columnData,
            columnCount: columns.length,
          },
        },
      }

      univer.createUnit(UniverInstanceType.UNIVER_SHEET, workbookData)

      univerInstanceRef.current = univer
      const univerApi = FUniver.newAPI(univer)
      univerApiRef.current = univerApi

      setTimeout(() => {
        const workbook = univerApi.getActiveWorkbook()
        if (workbook) {
          const sheet = workbook.getActiveSheet()
          if (sheet) {
            const columnsCfg: Record<number, string> = {}
            columns.forEach((field, colIndex) => {
              switch (field) {
                case "identifier": {
                  columnsCfg[colIndex] = t(`stuff.identifier`, "ID")
                  break
                }
                case "primaryCategory": {
                  columnsCfg[colIndex] = t(
                    `stuff.primaryCategory`,
                    "Primary Category",
                  )
                  break
                }
                case "secondaryCategory": {
                  columnsCfg[colIndex] = t(
                    `stuff.secondaryCategory`,
                    "Secondary Category",
                  )
                  break
                }
                case "fixedAsset": {
                  columnsCfg[colIndex] = t(`stuff.fixedAsset`, "Fixed Asset")
                  break
                }
                case "displayName": {
                  columnsCfg[colIndex] = t(`stuff.displayName`, "Name")
                  break
                }
                case "model": {
                  columnsCfg[colIndex] = t(`stuff.model`, "Model")
                  break
                }
                case "brand": {
                  columnsCfg[colIndex] = t(`stuff.brand`, "Brand")
                  break
                }
                case "purchaseDate": {
                  columnsCfg[colIndex] = t(
                    `stuff.purchaseDate`,
                    "Purchase Date",
                  )
                  break
                }
                case "source": {
                  columnsCfg[colIndex] = t(`stuff.source`, "Source")
                  break
                }
                case "originalUsage": {
                  columnsCfg[colIndex] = t(
                    `stuff.originalUsage`,
                    "Original Usage",
                  )
                  break
                }
                case "notes": {
                  columnsCfg[colIndex] = t(`stuff.notes`, "Notes")
                  break
                }
                default: {
                  columnsCfg[colIndex] = field
                  break
                }
              }
            })
            sheet.customizeColumnHeader({ columnsCfg })
          }
        }
      }, 0)

      univerApi.addEvent(univerApi.Event.SelectionChanged, () => {
        const workbook = univerApi.getActiveWorkbook()
        if (!workbook) {
          return
        }
        const sheet = workbook.getActiveSheet()
        if (!sheet) {
          return
        }
        const selection = sheet.getSelection()
        if (!selection) {
          return
        }
        const range = selection.getActiveRange()
        if (!range) {
          return
        }
        const rowIndex = range.getRow()
        const rowMetadata = sheet.getRowCustomMetadata(rowIndex)
        const colIndex = range.getColumn()
        const colMetadata = sheet.getColumnCustomMetadata(colIndex)
        setSelectedCellMeta({
          uuid: rowMetadata?.uuid ? String(rowMetadata.uuid) : "-",
          order: rowMetadata?.order ? String(rowMetadata.order) : "-",
          field: colMetadata?.field ? String(colMetadata.field) : "-",
        })
      })

      univerApi.addEvent(univerApi.Event.BeforeCommandExecute, (event) => {
        if (!univerApiRef.current) {
          return
        }
        if (COMMAND_LISTENER_SKELETON_CHANGE.includes(event.id)) {
          const ev = event as CommandListenerSkeletonChange
          switch (ev.id) {
            case SheetSkeletonChangeType.REMOVE_ROW: {
              const { range } = ev.params
              if (!range) {
                return
              }
              const workbook = univerApiRef.current.getActiveWorkbook()
              if (!workbook) {
                return
              }
              const sheet = workbook.getActiveSheet()
              if (!sheet) {
                return
              }

              void (async () => {
                for (
                  let rowIndex = range.startRow;
                  rowIndex <= range.endRow;
                  rowIndex++
                ) {
                  console.info("Processing row:", rowIndex)
                  await removeStuff(sheet, rowIndex)
                }
              })()

              break
            }
          }
        }
      })

      univerApi.addEvent(univerApi.Event.CommandExecuted, (event) => {
        if (!univerApiRef.current) {
          return
        }
        if (event.type !== CommandType.MUTATION) {
          return
        }
        if (COMMAND_LISTENER_VALUE_CHANGE.includes(event.id)) {
          const ev = event as CommandListenerValueChange
          switch (ev.id) {
            case SheetValueChangeType.SET_RANGE_VALUES: {
              const workbook = univerApiRef.current.getActiveWorkbook()
              if (!workbook) {
                return
              }
              const sheet = workbook.getActiveSheet()
              if (!sheet) {
                return
              }
              const cells = sheet
                .getRange(0, 0, sheet.getMaxRows(), sheet.getMaxColumns())
                .getValues()
              if (!cells) {
                return
              }

              void (async () => {
                const affectedRows = Object.keys(ev.params.cellValue || {})
                  .map(Number)
                  .sort((a, b) => a - b)
                const assignedOrders = assignOrders(sheet, cells, affectedRows)
                for (const rowIndex of affectedRows) {
                  console.info("Processing row:", rowIndex)
                  await createOrUpdateRow(
                    sheet,
                    cells,
                    rowIndex,
                    assignedOrders[rowIndex],
                  )
                }
              })()

              break
            }
          }
        }
      })
    })()

    return () => {
      disposed = true
      if (univerInstanceRef.current) {
        univerInstanceRef.current.dispose()
        univerInstanceRef.current = null
      }
      univerApiRef.current = null
    }
  }, [stuffService])

  if (!stuffService) {
    return (
      <Box
        sx={{
          height: "100vh",
          display: "flex",
          alignItems: "center",
          justifyContent: "center",
        }}
      >
        <CircularProgress />
      </Box>
    )
  }

  return (
    <Box sx={{ height: "100vh", display: "flex", flexDirection: "column" }}>
      <Toolbar />
      <Box sx={{ padding: 1, display: "flex", alignItems: "center" }}>
        <Typography
          variant="body2"
          color="text.secondary"
          sx={{ marginRight: 2 }}
        >
          <strong>uuid:</strong> {selectedCellMeta?.uuid ?? "—"}
        </Typography>
        <Typography
          variant="body2"
          color="text.secondary"
          sx={{ marginRight: 2 }}
        >
          <strong>order:</strong> {selectedCellMeta?.order ?? "—"}
        </Typography>
        <Typography
          variant="body2"
          color="text.secondary"
          sx={{ marginRight: 2 }}
        >
          <strong>field:</strong> {selectedCellMeta?.field ?? "—"}
        </Typography>
      </Box>
      <Box ref={univerDivRef} sx={{ flexGrow: 1 }} />
    </Box>
  )
}

export default HomePage
