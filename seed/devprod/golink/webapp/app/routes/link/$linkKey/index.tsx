import { timestampDate } from "@bufbuild/protobuf/wkt"
import React, { useCallback, useEffect, useState } from "react"
import { useTranslation } from "react-i18next"

import Box from "@mui/material/Box"
import Chip from "@mui/material/Chip"
import Container from "@mui/material/Container"
import MuiLink from "@mui/material/Link"
import Toolbar from "@mui/material/Toolbar"
import Typography from "@mui/material/Typography"

import { useGolinkService } from "@//seed/devprod/golink/client/tsx/GolinkServiceContext"
import type { Link } from "@//seed/devprod/golink/proto/golink_pb"

const DetailRow: React.FC<{
  label: string
  children: React.ReactNode
}> = ({ label, children }) => (
  <Box sx={{ marginTop: 2, display: "flex", flexDirection: "row" }}>
    <Typography
      sx={{ width: 120, flexShrink: 0, color: "text.secondary" }}
      variant="body2"
    >
      {label}
    </Typography>
    <Typography variant="body2">{children}</Typography>
  </Box>
)

const LinkPage: React.FC<{ params: { linkKey: string } }> = ({ params }) => {
  const { linkKey } = params
  const { t } = useTranslation("link")
  const golinkService = useGolinkService()
  const [link, setLink] = useState<Link | null>(null)
  const [error, setError] = useState<string | null>(null)

  const fetchLink = useCallback(async () => {
    if (!golinkService) {
      return
    }
    try {
      const result = await golinkService.GetLink(linkKey)
      setLink(result)
    } catch (e) {
      setError(e instanceof Error ? e.message : String(e))
    }
  }, [golinkService, linkKey])

  useEffect(() => {
    fetchLink()
  }, [fetchLink])

  return (
    <Box>
      <Toolbar />
      <Container component="main">
        {!!error && (
          <Box>
            <Typography variant="h5" sx={{ marginTop: 2 }}>
              {t("error.title", "Error")}
            </Typography>
            <Typography color="error" sx={{ marginTop: 1 }}>
              {error}
            </Typography>
          </Box>
        )}
        {!error && !link && (
          <Typography sx={{ marginTop: 2 }}>
            {t("loading.loadingHint", "Loading...")}
          </Typography>
        )}
        {!error && !!link && (
          <Box sx={{ marginTop: 2 }}>
            <Box
              sx={{
                display: "flex",
                alignItems: "center",
                marginBottom: 2,
              }}
            >
              <Typography variant="h4">go/{link.key}</Typography>
              <Chip
                label={
                  link.public
                    ? t("link.publicOption", "Public")
                    : t("link.privateOption", "Private")
                }
                size="small"
                color={link.public ? "success" : "default"}
                variant="outlined"
              />
            </Box>
            <Box
              sx={{
                display: "flex",
                flexDirection: "column",
              }}
            >
              <DetailRow label={t("link.targetLabel", "Target")}>
                <MuiLink href={link.target} target="_blank" rel="noopener">
                  {link.target}
                </MuiLink>
              </DetailRow>
              <DetailRow label={t("link.ownerLabel", "Owner")}>
                {link.owner ?? "—"}
              </DetailRow>
              <DetailRow label={t("link.hitsLabel", "Hits")}>
                {link.hitCount.toString()}
              </DetailRow>
              {!!link.createdTime && (
                <DetailRow label={t("link.createdLabel", "Created")}>
                  {timestampDate(link.createdTime).toLocaleString()}
                </DetailRow>
              )}
              {!!link.updatedTime && (
                <DetailRow label={t("link.updatedLabel", "Updated")}>
                  {timestampDate(link.updatedTime).toLocaleString()}
                </DetailRow>
              )}
            </Box>
          </Box>
        )}
      </Container>
    </Box>
  )
}

export default LinkPage
