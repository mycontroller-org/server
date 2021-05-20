package webhandler

import "time"

// global constants
const (
	KeyUserID     = "user_id"
	KeyFullName   = "fullname"
	KeyAuthorized = "authorized"
	KeyExpiresAt  = "expires_at"

	EnvJwtAccessSecret = "JWT_ACCESS_SECRET" // environment variable to set secret for JWT token

	HeaderAuthorization = "Authorization"
	HeaderUserID        = "mc_userid"

	AccessToken = "access_token"

	DefaultExpiration = time.Hour * 24 * 7 // 24 * 7 days
)

// UserLogin struct
type UserLogin struct {
	Username  string `json:"username"`
	Password  string `json:"password"`
	ExpiresIn string `json:"expiresIn"`
}

// JwtToken struct
type JwtToken struct {
	ID    string `json:"id"`
	Email string `json:"email"`
}

// JwtTokenResponse struct
type JwtTokenResponse struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	FullName string `json:"fullName"`
	Email    string `json:"email"`
	Token    string `json:"token"`
}

// Response struct
type Response struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}
