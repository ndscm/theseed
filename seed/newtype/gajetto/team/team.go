package team

import (
	"context"
)

// Person is a member of a [Team].
//
// Each person is identified by a personId, which may be either a UUID or a
// handle. When it is a handle, it must match the value returned by
// [Person.GetHandle]. When the team is linked to an IdP tenant, the UUID
// must correspond to the user's sub claim.
type Person interface {
	// GetPersonId reports the person's unique identifier.
	GetPersonId(ctx context.Context) (string, error)

	// GetHandle reports the person's handle.
	GetHandle(ctx context.Context) (string, error)

	// GetDisplayName reports the person's human-readable display name.
	GetDisplayName(ctx context.Context) (string, error)

	// GetOrganic reports the person's organic identifier.
	GetOrganic(ctx context.Context) (string, error)
}

// Team is a named group of [Person] members.
type Team interface {
	// GetTeamId reports the team's unique identifier. It returns "" if the team is not
	// linked to an IdP tenant.
	GetTeamId(ctx context.Context) (string, error)

	// GetHandle reports the team's unique handle.
	GetHandle(ctx context.Context) (string, error)

	// GetDisplayName reports the team's human-readable display name.
	GetDisplayName(ctx context.Context) (string, error)

	// GetMember looks up a member by personId. It returns the person and true
	// if found, or nil and false otherwise.
	GetMember(ctx context.Context, personId string) (Person, error)

	// GetMemberByHandle looks up a member by handle. It returns the person and true
	// if found, or nil and false otherwise.
	GetMemberByHandle(ctx context.Context, personHandle string) (Person, error)

	// ListMembers returns all members of the team, keyed by personId.
	ListMembers(ctx context.Context) (map[string]Person, error)

	// Auth validates the context authentication and returns the personId of the
	// authenticated member.
	Auth(ctx context.Context) (personId string, err error)
}
