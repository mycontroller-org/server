package service_token

import "time"

type ServiceToken struct {
	ID          string    `json:"id" yaml:"id"`
	UserID      string    `json:"userId" yaml:"userId"`
	Description string    `json:"description" yaml:"description"`
	Token       string    `json:"token" yaml:"token"` // keeps hashed token, not the actual token
	NeverExpire bool      `json:"neverExpire" yaml:"neverExpire"`
	ExpiresAt   time.Time `json:"expiresAt" yaml:"expiresAt"`
	CreatedAt   time.Time `json:"createdAt" yaml:"createdAt"`
}

type CreateTokenResponse struct {
	ID    string `json:"id" yaml:"id"`
	Token string `json:"token" yaml:"token"`
}
