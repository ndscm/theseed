package team

type Person interface {
	GetHandle() string
	GetDisplayName() string

	Auth(token string) error
}

type Team interface {
	GetHandle() string
	GetDisplayName() string
	GetMember(personId string) (Person, bool)

	Auth(token string) (personId string, err error)
}
