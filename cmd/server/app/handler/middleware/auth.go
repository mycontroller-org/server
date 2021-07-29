package handler

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	handlerUtils "github.com/mycontroller-org/server/v2/cmd/server/app/handler/utils"
	"github.com/mycontroller-org/server/v2/pkg/model/user"
	handlerType "github.com/mycontroller-org/server/v2/pkg/model/web_handler"
	"github.com/mycontroller-org/server/v2/pkg/utils/convertor"
	"github.com/mycontroller-org/server/v2/pkg/version"
	"go.uber.org/zap"

	"github.com/golang-jwt/jwt"
)

var (
	// middler check vars
	verifyPrefixes    = []string{"/api/", handlerType.SecureShareDirWebHandlerPath}
	nonRestrictedAPIs = []string{
		"/api/status", "/api/user/registration", "/api/user/login",
		"/api/oauth/login", "/api/oauth/token"}
)

// MiddlewareAuthenticationVerification verifies user auth details
func MiddlewareAuthenticationVerification(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		isSecurePrefix := false
		for _, verifyPrefix := range verifyPrefixes {
			if strings.HasPrefix(path, verifyPrefix) {
				isSecurePrefix = true
				break
			}
		}

		if isSecurePrefix {
			for _, aPath := range nonRestrictedAPIs {
				if strings.HasPrefix(path, aPath) {
					next.ServeHTTP(w, r)
					return
				}
			}
			// verify token and allow
			if err := IsValidToken(r); err == nil {
				next.ServeHTTP(w, r)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			handlerUtils.PostErrorResponse(w, "Authentication required", 401)
			return
		}
		next.ServeHTTP(w, r)

	})
}

func IsValidToken(r *http.Request) error {
	token, claims, err := getJwtToken(r)
	if err != nil {
		return err
	}
	if !token.Valid {
		return errors.New("invalid token")
	}

	// verify the validity
	expiresAt := convertor.ToInteger(claims[handlerType.KeyExpiresAt])
	if time.Now().Unix() >= expiresAt {
		return errors.New("expired token")
	}

	// clear userID header, might be injected from external
	// add userID into request header from here
	r.Header.Del(handlerType.HeaderUserID)
	if userID, ok := claims[handlerType.KeyUserID]; ok {
		id, ok := userID.(string)
		if ok {
			r.Header.Set(handlerType.HeaderUserID, id)
		}
	}
	return nil
}

func getJwtToken(r *http.Request) (*jwt.Token, jwt.MapClaims, error) {
	tokenString := extractJwtToken(r)
	if tokenString == "" {
		return nil, nil, errors.New("token not supplied")
	}
	claims := jwt.MapClaims{}
	token, err := jwt.ParseWithClaims(tokenString, &claims, func(token *jwt.Token) (interface{}, error) {
		// Make sure that the token method conform to "SigningMethodHMAC"
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return getJwtSecret(), nil
	})
	if err != nil {
		return nil, nil, err
	}
	return token, claims, err
}

func extractJwtToken(r *http.Request) string {
	// authCookie, err := r.Cookie("authToken")
	// if err == nil {
	// 	return authCookie.Value
	// }

	accessToken := r.Header.Get(handlerType.HeaderAuthorization)
	// normally, Authorization: Bearer the_token_xxx
	if accessToken != "" {
		if !strings.Contains(accessToken, " ") {
			return accessToken
		}
		strArr := strings.Split(accessToken, " ")
		if len(strArr) == 2 {
			return strArr[1]
		}
	}

	// verify query param has authorization token
	accessToken = r.URL.Query().Get(handlerType.AccessToken)
	if accessToken != "" {
		return accessToken
	}

	return ""
}

// CreateToken creates a token for a user
func CreateToken(user user.User, expiresIn string) (string, error) {
	atClaims := jwt.MapClaims{}
	atClaims[handlerType.KeyAuthorized] = true
	atClaims[handlerType.KeyUserID] = user.ID
	atClaims[handlerType.KeyFullName] = user.FullName

	expiresInDuration := handlerType.DefaultExpiration

	if expiresIn != "" {
		expiresInReceived, err := time.ParseDuration(expiresIn)
		if err != nil {
			zap.L().Error("error on parse the duration", zap.String("expiration", expiresIn), zap.Error(err))
		} else if expiresInReceived > 2*time.Second { // minimum expiration duration is 2 minutes
			expiresInDuration = expiresInReceived
		}
	}

	atClaims[handlerType.KeyExpiresAt] = time.Now().Add(expiresInDuration).Unix()
	at := jwt.NewWithClaims(jwt.SigningMethodHS256, atClaims)
	token, err := at.SignedString(getJwtSecret())
	if err != nil {
		return "", err
	}
	return token, nil
}

// GetUserID returns the logged in user details
func GetUserID(r *http.Request) string {
	return r.Header.Get(handlerType.HeaderUserID)
}

func getJwtSecret() []byte {
	return []byte(fmt.Sprintf("%s_%s", os.Getenv(handlerType.EnvJwtAccessSecret), version.Get().HostID))
}
