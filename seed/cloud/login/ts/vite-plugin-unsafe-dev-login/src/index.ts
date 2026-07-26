import crypto from "node:crypto"

import { type Plugin, type ViteDevServer } from "vite"

import { type OpenidConfiguration } from "../../../../../infra/auth/ts/openid"

const parseCookies = (header: string): Record<string, string> => {
  const result: Record<string, string> = {}
  for (const pair of header.split(";")) {
    const idx = pair.indexOf("=")
    if (idx === -1) {
      continue
    }
    const key = pair.slice(0, idx).trim()
    const value = pair.slice(idx + 1).trim()
    result[key] = decodeURIComponent(value)
  }
  return result
}

const isSafeRedirect = (url: string): boolean => {
  return url.startsWith("/") && !url.startsWith("//")
}

interface LoginOptions {
  discoveryUrl?: string
  clientId: string
  clientSecret?: string
  // When set, exchange the access token for a Keycloak Requesting Party Token
  // before handing it out, so the bearer already carries the authorization
  // claim and services (e.g. hooin) need not fetch the RPT themselves. The
  // audience defaults to clientId, the resource server that owns the resources.
  uma?: boolean
  umaAudience?: string
}

// TokenResponse is the subset of a token endpoint reply we carry between
// requests, whether it came from a code exchange, a refresh, or a uma-ticket
// grant. An RPT is returned like any other token, with its own refresh token.
interface TokenResponse {
  access_token: string
  refresh_token?: string
  expires_in?: number
  refresh_expires_in?: number
}

export const unsafeDevLogin = (options: LoginOptions): Plugin => {
  const discoveryUrl =
    options.discoveryUrl || process.env.SEED_OPENID_DISCOVERY_URL || ""
  const clientId = options.clientId
  const clientSecret = options.clientSecret
  const uma = options.uma || false
  const umaAudience = options.umaAudience || clientId

  let openidConfiguration: OpenidConfiguration | null = null

  const getOpenidConfiguration = async (): Promise<OpenidConfiguration> => {
    if (openidConfiguration) {
      return openidConfiguration
    }
    if (!discoveryUrl) {
      throw new Error("login openid discovery url is required.")
    }
    const response = await fetch(discoveryUrl)
    if (!response.ok) {
      throw new Error(
        `openid discovery failed: ${response.status} ${response.statusText}`,
      )
    }
    openidConfiguration = await response.json()
    return openidConfiguration!
  }

  // umaGrant turns a token response into a Requesting Party Token so
  // the bearer already carries the authorization claim. It requests no specific
  // permission, so Keycloak returns every grant the subject holds on the
  // audience's resources. A subject with no grant is denied the exchange (403),
  // in which case the plain token response is kept so unprotected routes still
  // work — the protected ones are refused later all the same.
  //
  // The whole RPT response is returned, including its own refresh token, so the
  // session can refresh the RPT directly rather than re-running this exchange.
  const umaGrant = async (
    authnToken: TokenResponse,
  ): Promise<TokenResponse> => {
    const config = await getOpenidConfiguration()
    const body = new URLSearchParams({
      grant_type: "urn:ietf:params:oauth:grant-type:uma-ticket",
      audience: umaAudience,
    })
    const response = await fetch(config.token_endpoint, {
      method: "POST",
      headers: {
        "Content-Type": "application/x-www-form-urlencoded",
        Authorization: `Bearer ${authnToken.access_token}`,
      },
      body: body.toString(),
    })
    if (!response.ok) {
      const text = await response.text()
      console.warn(
        `party token request failed, using plain access token: ${response.status} ${text}`,
      )
      return authnToken
    }
    const authzToken: TokenResponse = await response.json()
    return authzToken
  }

  const configureServer = (server: ViteDevServer) => {
    if (!clientId) {
      throw new Error("login openid client id is required.")
    }

    server.middlewares.use("/auth/login", async (req, res) => {
      const url = new URL(req.url || "/", `http://${req.headers.host}`)
      const rawReturnUrl = url.searchParams.get("return") || "/"
      const returnUrl = isSafeRedirect(rawReturnUrl) ? rawReturnUrl : "/"

      const openidConfiguration = await getOpenidConfiguration()
      const authorizationEndpoint = openidConfiguration.authorization_endpoint

      const codeVerifier = crypto.randomBytes(64).toString("hex")
      const codeChallenge = crypto
        .createHash("sha256")
        .update(codeVerifier)
        .digest("base64url")

      const params = new URLSearchParams({
        response_type: "code",
        client_id: clientId,
        redirect_uri: `${url.protocol}//${url.host}/auth/callback`,
        scope: "openid profile email",
        code_challenge: codeChallenge,
        code_challenge_method: "S256",
        state: returnUrl,
      })

      res.setHeader("Set-Cookie", [
        `pkce_code_verifier=${codeVerifier}; Path=/; HttpOnly; SameSite=Lax`,
      ])
      res.statusCode = 307
      res.setHeader("Location", `${authorizationEndpoint}?${params.toString()}`)
      res.end()
    })

    server.middlewares.use("/auth/callback", async (req, res) => {
      const url = new URL(req.url || "/", `http://${req.headers.host}`)
      const code = url.searchParams.get("code")
      const state = url.searchParams.get("state")
      const errorParam = url.searchParams.get("error")

      if (errorParam) {
        res.statusCode = 400
        res.end(
          `OAuth error: ${errorParam}: ${url.searchParams.get("error_description") || ""}`,
        )
        return
      }

      if (!code) {
        res.statusCode = 400
        res.end("Missing authorization code")
        return
      }

      const cookies = parseCookies(req.headers.cookie || "")
      const codeVerifier = cookies["pkce_code_verifier"]
      if (!codeVerifier) {
        res.statusCode = 400
        res.end("Missing PKCE code verifier")
        return
      }

      const openidConfiguration = await getOpenidConfiguration()
      const tokenEndpoint = openidConfiguration.token_endpoint

      const body = new URLSearchParams({
        grant_type: "authorization_code",
        client_id: clientId,
        code,
        redirect_uri: `${url.protocol}//${url.host}/auth/callback`,
        code_verifier: codeVerifier,
      })
      if (clientSecret) {
        body.set("client_secret", clientSecret)
      }

      const tokenResp = await fetch(tokenEndpoint, {
        method: "POST",
        headers: { "Content-Type": "application/x-www-form-urlencoded" },
        body: body.toString(),
      })

      if (!tokenResp.ok) {
        const text = await tokenResp.text()
        res.statusCode = 502
        res.end(`Token exchange failed: ${text}`)
        return
      }

      const authnToken: TokenResponse = await tokenResp.json()
      const authzToken = uma ? await umaGrant(authnToken) : authnToken
      const accessMaxAge = authzToken.expires_in || 3600
      const setCookies: string[] = [
        `access_token=${authzToken.access_token}; Path=/; HttpOnly; SameSite=Lax; Max-Age=${accessMaxAge}`,
        "pkce_code_verifier=; Path=/; Max-Age=0",
      ]
      if (authzToken.refresh_token) {
        const refreshMaxAge = authzToken.refresh_expires_in || 7 * 24 * 3600
        setCookies.push(
          `refresh_token=${authzToken.refresh_token}; Path=/; HttpOnly; SameSite=Lax; Max-Age=${refreshMaxAge}`,
        )
      }

      res.setHeader("Set-Cookie", setCookies)
      res.statusCode = 307
      const redirect = state && isSafeRedirect(state) ? state : "/"
      res.setHeader("Location", redirect)
      res.end()
    })

    server.middlewares.use(async (req, res, next) => {
      const cookies = parseCookies(req.headers.cookie || "")
      let accessToken = cookies["access_token"]

      if (!accessToken && cookies["refresh_token"]) {
        try {
          const config = await getOpenidConfiguration()
          const body = new URLSearchParams({
            grant_type: "refresh_token",
            client_id: clientId,
            refresh_token: cookies["refresh_token"],
          })
          if (clientSecret) {
            body.set("client_secret", clientSecret)
          }
          const refreshResp = await fetch(config.token_endpoint, {
            method: "POST",
            headers: { "Content-Type": "application/x-www-form-urlencoded" },
            body: body.toString(),
          })
          if (refreshResp.ok) {
            // The refresh token in the cookie is the RPT's own, so refreshing it
            // should hand back another RPT — that is what we are testing. We use
            // the result as is rather than re-running the uma-ticket exchange, so
            // a token that came back without the authorization claim would show
            // up as hooin fetching the RPT itself again.
            const authzToken: TokenResponse = await refreshResp.json()
            accessToken = authzToken.access_token
            const accessMaxAge = authzToken.expires_in || 3600
            const setCookies: string[] = [
              `access_token=${authzToken.access_token}; Path=/; HttpOnly; SameSite=Lax; Max-Age=${accessMaxAge}`,
            ]
            if (authzToken.refresh_token) {
              const refreshMaxAge =
                authzToken.refresh_expires_in || 7 * 24 * 3600
              setCookies.push(
                `refresh_token=${authzToken.refresh_token}; Path=/; HttpOnly; SameSite=Lax; Max-Age=${refreshMaxAge}`,
              )
            }
            res.setHeader("Set-Cookie", setCookies)
          } else {
            res.setHeader("Set-Cookie", [
              "access_token=; Path=/; Max-Age=0",
              "refresh_token=; Path=/; Max-Age=0",
            ])
          }
        } catch {
          res.setHeader("Set-Cookie", [
            "access_token=; Path=/; Max-Age=0",
            "refresh_token=; Path=/; Max-Age=0",
          ])
        }
      }

      if (accessToken) {
        req.headers["Authorization"] = `Bearer ${accessToken}`
      }
      next()
    })
  }

  return {
    name: "vite-plugin-unsafe-dev-login",
    configureServer,
  }
}

export default unsafeDevLogin
