package webhandler

import "time"

// global constants
const (
	KeyUserID         = "user_id"
	KeyServiceTokenID = "svc_token_id"
	KeyFullName       = "fullname"
	KeyAuthorized     = "authorized"
	KeyExpiresAt      = "expires_at"

	EnvJwtAccessSecret = "JWT_ACCESS_SECRET" // environment variable to set secret for JWT token

	HeaderAuthorization = "Authorization"
	HeaderUserID        = "mc_userid"

	AccessToken = "access_token"

	DefaultTokenExpiration = time.Hour * 24 // 24 hours
	AUTH_COOKIE_NAME       = "__mc_auth"
	SIGNOUT_PATH           = "/api/mc_auth/sign_out"

	SecureShareDirWebHandlerPath   = "/secure_share"
	InsecureShareDirWebHandlerPath = "/insecure_share"
)

// UserLogin struct
type UserLogin struct {
	Username  string `json:"username" yaml:"username"`
	Password  string `json:"password" yaml:"password"`
	SvcToken  string `json:"token" yaml:"token"`
	ExpiresIn string `json:"expiresIn" yaml:"expiresIn"`
}

// JwtToken struct
type JwtToken struct {
	ID    string `json:"id" yaml:"id"`
	Email string `json:"email" yaml:"email"`
}

// JwtTokenResponse struct
type JwtTokenResponse struct {
	ID       string `json:"id" yaml:"id"`
	Username string `json:"username" yaml:"username"`
	FullName string `json:"fullName" yaml:"fullName"`
	Email    string `json:"email" yaml:"email"`
	Token    string `json:"token" yaml:"token"`
}

// Response struct
type Response struct {
	Success bool        `json:"success" yaml:"success"`
	Message string      `json:"message" yaml:"message"`
	Data    interface{} `json:"data" yaml:"data"`
}

// ActionConfig struct
type ActionConfig struct {
	Resource string `json:"resource" yaml:"resource"`
	KayPath  string `json:"keyPath" yaml:"keyPath"`
	Payload  string `json:"payload" yaml:"payload"`
}
