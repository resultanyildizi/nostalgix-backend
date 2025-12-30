package entity

type AuthMethod string

const (
	AuthMethodAnonymous AuthMethod = "anonymous"
	AuthMethodGoogle    AuthMethod = "google"
	AuthMethodApple     AuthMethod = "apple"
)
