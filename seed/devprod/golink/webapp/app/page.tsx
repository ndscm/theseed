import * as Protobuf from "@bufbuild/protobuf"
import React, { useCallback, useEffect, useRef, useState } from "react"
import { useTranslation } from "react-i18next"

import Box from "@mui/material/Box"
import Button from "@mui/material/Button"
import Container from "@mui/material/Container"
import IconButton from "@mui/material/IconButton"
import InputAdornment from "@mui/material/InputAdornment"
import MuiLink from "@mui/material/Link"
import TextField from "@mui/material/TextField"
import Toolbar from "@mui/material/Toolbar"
import Typography from "@mui/material/Typography"

import ArrowForwardOutlinedIcon from "@mui/icons-material/ArrowForwardOutlined"
import CloseOutlinedIcon from "@mui/icons-material/CloseOutlined"

import { useGolinkService } from "../../client/tsx/golink-service-context"
import { type Link, LinkSchema } from "../../proto/golink_pb"

const PAGE_SIZE = 10

const LinkCreator: React.FC<{
  defaultKey?: string
  onCreateLink: (key: string, target: string) => Promise<void>
}> = ({ defaultKey, onCreateLink }) => {
  const { t } = useTranslation("common")
  const [inputLinkKey, setInputLinkKey] = useState("")
  const [inputLinkTarget, setInputLinkTarget] = useState("")
  const [isSubmitting, setIsSubmitting] = useState(false)

  useEffect(() => {
    if (defaultKey) {
      setInputLinkKey(defaultKey)
    }
  }, [defaultKey])

  const handleSubmit = useCallback(async () => {
    if (!inputLinkKey || !inputLinkTarget) {
      return
    }
    setIsSubmitting(true)
    try {
      await onCreateLink(inputLinkKey, inputLinkTarget)
    } finally {
      setIsSubmitting(false)
    }
    setInputLinkKey("")
    setInputLinkTarget("")
  }, [inputLinkKey, inputLinkTarget, onCreateLink])

  return (
    <Box sx={{ marginTop: 2, marginBottom: 2 }}>
      <Typography variant="h4">
        {t("link.createTitle", "Create Link")}
      </Typography>
      <Box
        sx={{
          display: "flex",
          flexDirection: "row",
          alignItems: "center",
          flexWrap: "wrap",
        }}
      >
        <Box
          sx={{
            marginTop: 1,
            marginBottom: 1,
            display: "flex",
            flexDirection: "row",
            alignItems: "center",
          }}
        >
          <TextField
            size="small"
            fullWidth
            slotProps={{
              input: {
                startAdornment: (
                  <InputAdornment position="start">http://go/</InputAdornment>
                ),
              },
            }}
            placeholder="short-link"
            value={inputLinkKey}
            onChange={(e) => setInputLinkKey(e.target.value)}
          />
          <ArrowForwardOutlinedIcon sx={{ marginLeft: 1, marginRight: 1 }} />
        </Box>
        <Box
          sx={{
            flexGrow: 1,
            marginTop: 1,
            marginBottom: 1,
            display: "flex",
            flexDirection: "row",
            alignItems: "center",
          }}
        >
          <TextField
            size="small"
            fullWidth
            value={inputLinkTarget}
            onChange={(e) => setInputLinkTarget(e.target.value)}
            placeholder="https://example.com"
          />
          <Button
            variant="contained"
            sx={{ marginLeft: 2 }}
            onClick={handleSubmit}
            disabled={isSubmitting || !inputLinkKey || !inputLinkTarget}
          >
            {t("link.createButton", "Create")}
          </Button>
        </Box>
      </Box>
    </Box>
  )
}

interface LinkListProps {
  links: Link[]
  isLoading: boolean
  onRemoveLink: (key: string, etag: string) => Promise<void>
}

const LinkList: React.FC<LinkListProps> = ({
  links,
  isLoading,
  onRemoveLink,
}) => {
  const { t } = useTranslation("common")

  return (
    <Box sx={{ display: "flex", flexDirection: "column" }}>
      <Box
        sx={{
          display: "flex",
          flexDirection: "row",
          fontWeight: "bold",
          borderBottom: 1,
          borderColor: "divider",
          py: 1,
        }}
      >
        <Box sx={{ flex: 1 }}>{t("list.keyLabel", "Key")}</Box>
        <Box sx={{ flex: 2, marginLeft: 1 }}>
          {t("list.targetLabel", "Target")}
        </Box>
        <Box sx={{ flex: 1, marginLeft: 1 }}>
          {t("list.ownerLabel", "Owner")}
        </Box>
        <Box sx={{ width: 72, marginLeft: 1 }}>
          {t("list.hitsLabel", "Hits")}
        </Box>
        <Box sx={{ width: 48, marginLeft: 1 }}></Box>
      </Box>
      {links.map((link) => (
        <Box
          key={link.key}
          sx={{
            display: "flex",
            flexDirection: "row",
            borderBottom: 1,
            borderColor: "divider",
            paddingTop: 1,
            paddingBottom: 1,
          }}
        >
          <Box
            sx={{
              flex: 1,
              overflow: "hidden",
              textOverflow: "ellipsis",
              whiteSpace: "nowrap",
            }}
          >
            {link.key}
          </Box>
          <Box
            sx={{
              flex: 2,
              marginLeft: 1,
              overflow: "hidden",
              textOverflow: "ellipsis",
              whiteSpace: "nowrap",
            }}
          >
            <MuiLink href={link.target}>{link.target}</MuiLink>
          </Box>
          <Box
            sx={{
              flex: 1,
              marginLeft: 1,
              overflow: "hidden",
              textOverflow: "ellipsis",
              whiteSpace: "nowrap",
            }}
          >
            {link.owner}
          </Box>
          <Box
            sx={{
              width: 72,
              marginLeft: 1,
              overflow: "hidden",
              textOverflow: "ellipsis",
              whiteSpace: "nowrap",
            }}
          >
            {link.hitCount.toString()}
          </Box>
          <Box
            sx={{
              width: 48,
              marginLeft: 1,
              overflow: "hidden",
              textOverflow: "ellipsis",
              whiteSpace: "nowrap",
              textAlign: "right",
            }}
          >
            <IconButton
              size="small"
              onClick={() => onRemoveLink(link.key, link.etag)}
            >
              <CloseOutlinedIcon fontSize="small" color="error" />
            </IconButton>
          </Box>
        </Box>
      ))}
      {links.length === 0 && !isLoading && (
        <Box sx={{ py: 2, textAlign: "center" }}>
          {t("list.noLinksFound", "No links found")}
        </Box>
      )}
    </Box>
  )
}

const HomePage: React.FC<{ params: { linkKey?: string } }> = ({ params }) => {
  const { linkKey } = params
  const golinkService = useGolinkService()
  const [links, setLinks] = useState<Link[]>([])
  const [isLoading, setIsLoading] = useState(false)
  const [totalSize, setTotalSize] = useState<bigint>(BigInt(0))
  const [nextPageToken, setNextPageToken] = useState<string>("")
  const refOfListEnd = useRef<HTMLDivElement>(null)

  const fetchLinks = useCallback(async () => {
    if (!golinkService) {
      return
    }
    setIsLoading(true)
    try {
      const response = await golinkService.ListLinks({
        pageSize: PAGE_SIZE,
      })
      setLinks(response.links)
      setTotalSize(response.totalSize)
      setNextPageToken(response.nextPageToken)
    } finally {
      setIsLoading(false)
    }
  }, [golinkService])

  const handleCreateLink = useCallback(
    async (key: string, target: string) => {
      if (!golinkService) {
        return
      }
      const link: Link = Protobuf.create(LinkSchema, {
        key,
        target,
      })
      await golinkService.CreateLink(link)
      await fetchLinks()
    },
    [golinkService, fetchLinks],
  )

  const handleRemoveLink = useCallback(
    async (key: string, etag: string) => {
      if (!golinkService) {
        return
      }
      await golinkService.DeleteLink(key, { etag })
      await fetchLinks()
    },
    [golinkService, fetchLinks],
  )

  useEffect(() => {
    fetchLinks()
  }, [fetchLinks])

  // Infinite scroll observer
  useEffect(() => {
    const divOfListEnd = refOfListEnd.current
    const observer = new IntersectionObserver(
      async (entries) => {
        if (entries[0].isIntersecting) {
          observer.unobserve(entries[0].target)
          if (!golinkService) {
            return
          }
          if (nextPageToken) {
            const response = await golinkService.ListLinks({
              pageSize: PAGE_SIZE,
              pageToken: nextPageToken,
            })
            setLinks((prev) => [...prev, ...response.links])
            setTotalSize(response.totalSize)
            setNextPageToken(response.nextPageToken)
          }
        }
      },
      { threshold: 0.1 },
    )

    if (divOfListEnd) {
      observer.observe(divOfListEnd)
    }
    return () => {
      if (divOfListEnd) {
        observer.unobserve(divOfListEnd)
      }
    }
  }, [golinkService, nextPageToken])

  return (
    <Box>
      <Toolbar />
      <Container component="main">
        <LinkCreator defaultKey={linkKey} onCreateLink={handleCreateLink} />
        <LinkList
          links={links}
          isLoading={isLoading}
          onRemoveLink={handleRemoveLink}
        />
        <Box ref={refOfListEnd} sx={{ height: "1px" }} />
      </Container>
    </Box>
  )
}

export default HomePage
