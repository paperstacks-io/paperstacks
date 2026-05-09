package auth

type Session struct {
	Email          string
	Token          string
	ExpirationTime string
	IsValid        bool
}
