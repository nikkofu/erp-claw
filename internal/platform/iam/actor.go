package iam

// Actor represents the authenticated subject for a request.
type Actor struct {
	ID           string
	Roles        []string
	DepartmentID string
}

// SystemActor is used until IAM is wired and represents the runtime itself.
var SystemActor = Actor{
	ID:    "system",
	Roles: []string{"platform_admin"},
}

func (a Actor) HasRole(role string) bool {
	for _, candidate := range a.Roles {
		if candidate == role {
			return true
		}
	}
	return false
}
