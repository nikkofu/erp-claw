package iam

// Actor represents the authenticated subject for a request.
type Actor struct {
	ID string
}

// SystemActor is used until IAM is wired and represents the runtime itself.
var SystemActor = Actor{ID: "system"}
