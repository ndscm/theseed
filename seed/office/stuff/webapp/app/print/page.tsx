import React, { useEffect, useState } from "react"
import { useTranslation } from "react-i18next"
import { useSearchParams } from "react-router"

import Box from "@mui/material/Box"
import CircularProgress from "@mui/material/CircularProgress"

import { type Stuff } from "../../../proto/stuff_pb"
import { useStuffService } from "../../../client/tsx/stuff-service-context"


const TemplateMore: React.FC<{ stuff: Stuff }> = ({ stuff }) => {
  const { t } = useTranslation("common")
  const parsed = JSON.parse(stuff.data)

  return (
    <div
      style={{
        flexGrow: 1,
        padding: "15mm",
        fontSize: "10pt",
        display: "flex",
        flexDirection: "column",
      }}
    >
      <div style={{ marginBottom: "10mm", display: "flex" }}>
        <div
          style={{
            width: "20mm",
            height: "20mm",
            display: "flex",
            alignItems: "center",
            justifyContent: "center",
          }}
        >

        </div>
        <div
          style={{
            flexGrow: 1,
            display: "flex",
            flexDirection: "column",
            alignItems: "center",
            justifyContent: "center",
          }}
        >
          <div style={{ marginBottom: "2mm", fontSize: "12pt" }}>
            <strong>{parsed.primaryCategory}</strong>
          </div>
          <div style={{ fontSize: "10pt" }}>
            <strong>{parsed.secondaryCategory}</strong>
          </div>
        </div>
      </div>
      <div style={{ marginBottom: "2mm", display: "flex" }}>
        <div style={{ width: "20mm", flexShrink: 0 }}>
          <strong>{t("stuff.identifier")}: </strong>
        </div>
        <div>{parsed.identifier}</div>
      </div>
      <div style={{ marginBottom: "2mm", display: "flex" }}>
        <div style={{ width: "20mm", flexShrink: 0 }}>
          <strong>{t("stuff.displayName")}: </strong>
        </div>
        <div>{parsed.displayName}</div>
      </div>
      <div style={{ marginBottom: "2mm", display: "flex" }}>
        <div style={{ width: "20mm", flexShrink: 0 }}>
          <strong>{t("stuff.fixedAsset")}: </strong>
        </div>
        <div>{parsed.fixedAsset}</div>
      </div>
      <div style={{ marginBottom: "2mm", display: "flex" }}>
        <div style={{ width: "20mm", flexShrink: 0 }}>
          <strong>{t("stuff.model")}: </strong>
        </div>
        <div>{parsed.model}</div>
      </div>
      <div style={{ marginBottom: "2mm", display: "flex" }}>
        <div style={{ width: "20mm", flexShrink: 0 }}>
          <strong>{t("stuff.brand")}: </strong>
        </div>
        <div>{parsed.brand}</div>
      </div>
      <div style={{ marginBottom: "2mm", display: "flex" }}>
        <div style={{ width: "20mm", flexShrink: 0 }}>
          <strong>{t("stuff.purchaseDate")}: </strong>
        </div>
        <div>{parsed.purchaseDate}</div>
      </div>
      <div style={{ marginBottom: "2mm", display: "flex" }}>
        <div style={{ width: "20mm", flexShrink: 0 }}>
          <strong>{t("stuff.source")}: </strong>
        </div>
        <div>{parsed.source}</div>
      </div>
      <div style={{ marginBottom: "2mm", display: "flex" }}>
        <div style={{ width: "20mm", flexShrink: 0 }}>
          <strong>{t("stuff.originalUsage")}: </strong>
        </div>
        <div>{parsed.originalUsage}</div>
      </div>
      <div style={{ flexGrow: 1, display: "flex" }}>
        <div style={{ width: "20mm", flexShrink: 0 }}>
          <strong>{t("stuff.notes")}: </strong>
        </div>
        <div>{parsed.notes}</div>
      </div>
      <div
        style={{
          color: "#cccccc",
          fontSize: "8pt",
          display: "flex",
          justifyContent: "flex-end",
        }}
      >
        {new Date().toISOString().split("T")[0]}
      </div>
    </div>
  )
}

const getDimensions = (
  paperSize: string,
  landscape: boolean = false,
): { width: number; height: number } => {
  switch (paperSize) {
    case "a4": {
      return landscape
        ? { width: 297, height: 210 }
        : { width: 210, height: 297 }
    }
    case "a5": {
      return landscape
        ? { width: 210, height: 148 }
        : { width: 148, height: 210 }
    }
    case "a6": {
      return landscape
        ? { width: 148, height: 105 }
        : { width: 105, height: 148 }
    }
    default: {
      throw new Error(`paperSize is unsupported: ${paperSize}`)
    }
  }
}

const PrintPage: React.FC = () => {
  const { t } = useTranslation("common")
  const [searchParams] = useSearchParams()
  const stuffService = useStuffService()
  const [items, setItems] = useState<Stuff[]>([])
  const [isLoading, setIsLoading] = useState(true)

  const paperSize = searchParams.get("paperSize")
  const labelSize = searchParams.get("labelSize")
  const template = searchParams.get("template")
  const uuids = searchParams.get("uuids")

  if (!paperSize) {
    throw new Error("paperSize is required")
  }
  if (!labelSize) {
    throw new Error("labelSize is required")
  }
  if (!template) {
    throw new Error("template is required")
  }
  if (!uuids) {
    throw new Error("uuids is required")
  }

  const paperDimensions = getDimensions(paperSize)
  const labelDimensions = getDimensions(labelSize)

  useEffect(() => {
    document.title = t("page.print", "Print")
  }, [t])

  useEffect(() => {
    const style = document.createElement("style")
    style.textContent = `@page { size: ${paperDimensions.width}mm ${paperDimensions.height}mm; margin: 0; }`
    document.head.appendChild(style)
    return () => {
      document.head.removeChild(style)
    }
  }, [paperDimensions])

  useEffect(() => {
    if (!stuffService) {
      return
    }

    const uuidList = uuids.split(",").filter(Boolean)
    if (uuidList.length === 0) {
      throw new Error("uuids is required")
    }

    void (async () => {
      const results: Stuff[] = []
      for (const uuid of uuidList) {
        const stuff = await stuffService.GetStuff(uuid)
        results.push(stuff)
      }
      setItems(results)
      setIsLoading(false)
    })()
  }, [stuffService, searchParams])

  useEffect(() => {
    if (items.length > 0) {
      const debounce = setTimeout(() => {
        window.print()
      }, 1000)
      return () => clearTimeout(debounce)
    }
  }, [items])

  if (!stuffService || isLoading) {
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
    <Box
      sx={{
        display: "flex",
        flexWrap: "wrap",
      }}
    >
      {items.map((item) => (
        <Box
          key={item.uuid}
          sx={{
            width: `${labelDimensions.width}mm`,
            height: `${labelDimensions.height}mm`,
            border: "1px solid black",
            borderColor: "#cccccc",
            boxSizing: "border-box",
            display: "flex",
          }}
        >
          {(() => {
            switch (template) {
              case "more":
                return <TemplateMore stuff={item} />
              default:
                throw new Error(`template is unsupported: ${template}`)
            }
          })()}
        </Box>
      ))}
    </Box>
  )
}

export default PrintPage
