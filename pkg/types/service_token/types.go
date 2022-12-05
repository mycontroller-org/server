package service_token

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/mycontroller-org/server/v2/pkg/types/cmap"
	dateTimeTY "github.com/mycontroller-org/server/v2/pkg/types/cusom_datetime"
	"github.com/mycontroller-org/server/v2/pkg/utils"
)

type ServiceToken struct {
	ID          string                `json:"id" yaml:"id"`
	UserID      string                `json:"userId" yaml:"userId"`
	Description string                `json:"description" yaml:"description"`
	Token       Token                 `json:"token" yaml:"token"` // keeps hashed token, not the actual token
	NeverExpire bool                  `json:"neverExpire" yaml:"neverExpire"`
	ExpiresOn   dateTimeTY.CustomDate `json:"expiresOn" yaml:"expiresOn"`
	Labels      cmap.CustomStringMap  `json:"labels" yaml:"labels"`
	CreatedOn   time.Time             `json:"createdOn" yaml:"createdOn"`
}

type CreateTokenResponse struct {
	ID    string `json:"id" yaml:"id"`
	Token string `json:"token" yaml:"token"`
}

type Token struct {
	ID    string `json:"id" yaml:"id"`
	Token string `json:"token" yaml:"token"`
}

// returns new token
func GetNewToken() Token {
	return Token{
		ID:    fmt.Sprintf("%d%s", time.Now().UnixMilli(), utils.RandIDWithLength(9)),
		Token: utils.RandIDWithLength(32),
	}
}

func (t *Token) GetTokenWithID() string {
	return fmt.Sprintf("%s_%s", t.ID, t.Token)
}

func ParseToken(rawToken string) (*Token, error) {
	rawTokens := strings.SplitAfterN(rawToken, "_", 2)
	if len(rawTokens) != 2 {
		return nil, errors.New("invalid token")
	}

	return &Token{ID: strings.TrimSuffix(rawTokens[0], "_"), Token: rawTokens[1]}, nil
}
