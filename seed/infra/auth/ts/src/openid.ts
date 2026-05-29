// See: https://openid.net/specs/openid-connect-discovery-1_0.html#ProviderMetadata
export type OpenidConfiguration = {
  issuer: string
  authorization_endpoint: string
  token_endpoint: string
  userinfo_endpoint: string
  jwks_uri: string
  scopes_supported: string[]
  response_types_supported: string[]
  grant_types_supported: string[]
  subject_types_supported: string[]
  claims_supported: string[]

  // See: https://datatracker.ietf.org/doc/html/rfc8628#section-7.4
  device_authorization_endpoint?: string
}

export default {}
