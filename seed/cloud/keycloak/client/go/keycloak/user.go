package keycloak

// UserRepresentation is the Keycloak admin REST representation of a realm user. Only
// the fields the seed uses are declared; Keycloak returns many more.
type UserRepresentation struct {
	Id string `json:"id"`

	Username string `json:"username"`

	FirstName string `json:"firstName,omitempty"`

	LastName string `json:"lastName,omitempty"`

	Email string `json:"email,omitempty"`

	// Attributes holds the user's custom attributes. Each attribute maps to a
	// list of values, mirroring the Keycloak representation.
	Attributes map[string][]string `json:"attributes,omitempty"`

	Enabled bool `json:"enabled"`
}
