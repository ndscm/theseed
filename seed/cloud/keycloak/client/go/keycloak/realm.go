package keycloak

// RealmRepresentation is the Keycloak admin REST representation of a realm. Only the
// fields the seed uses are declared; Keycloak returns many more.
type RealmRepresentation struct {
	// Id is the realm's internal UUID, distinct from Realm.
	Id string `json:"id"`

	// Realm is the realm's name, the slug that appears in admin REST paths.
	Realm string `json:"realm"`

	DisplayName string `json:"displayName,omitempty"`
}
