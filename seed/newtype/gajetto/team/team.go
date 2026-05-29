package team

// Person is a member of a [Team].
//
// Each person is identified by a personId, which may be either a UUID or a
// handle. When it is a handle, it must match the value returned by
// [Person.GetHandle]. When the team is linked to an IdP tenant, the UUID
// must correspond to the user's sub claim.
type Person interface {
	// GetPersonId reports the person's unique identifier.
	GetPersonId() string
	// GetHandle reports the person's handle.
	GetHandle() string
	// GetDisplayName reports the person's human-readable display name.
	GetDisplayName() string

	// Auth verifies that the given token is valid for this person.
	Auth(token string) error
}

// Team is a named group of [Person] members.
type Team interface {
	// GetTeamUuid reports the team's UUID. It returns "" if the team is not
	// linked to an IdP tenant.
	GetTeamUuid() string
	// GetHandle reports the team's unique handle.
	GetHandle() string
	// GetDisplayName reports the team's human-readable display name.
	GetDisplayName() string

	// GetMember looks up a member by personId. It returns the person and true
	// if found, or nil and false otherwise.
	GetMember(personId string) (Person, bool)

	// Auth validates the given token and returns the personId of the
	// authenticated member.
	Auth(token string) (personId string, err error)
}
