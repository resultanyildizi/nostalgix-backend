package entity

type AuthMethod string

const (
	AuthMethodAnonymous AuthMethod = "anonymous"
	AuthMethodGoogle    AuthMethod = "google"
	AuthMethodApple     AuthMethod = "apple"
)

type AuthTokens struct {
	RefreshToken string `json:"refresh_token"`
	AccessToken  string `json:"access_token"`
}
