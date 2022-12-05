package handler

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	handlerUtils "github.com/mycontroller-org/server/v2/cmd/server/app/handler/utils"
	"github.com/mycontroller-org/server/v2/pkg/types"
	"github.com/mycontroller-org/server/v2/pkg/types/user"
	handlerTY "github.com/mycontroller-org/server/v2/pkg/types/web_handler"
	"github.com/mycontroller-org/server/v2/pkg/utils/convertor"
	"github.com/mycontroller-org/server/v2/pkg/version"
	"go.uber.org/zap"

	jwt "github.com/golang-jwt/jwt/v4"
)

var (
	// middler check vars
	verifyPrefixes = []string{
		"/api/",                                // all api
		handlerTY.SecureShareDirWebHandlerPath, // web file secure share api
	}
	nonRestrictedAPIs = []string{
		"/api/status",                            // reports mycontroller server status
		"/api/user/registration",                 // register new user. TODO: this api not used. verify and remove this
		"/api/user/login",                        // login api
		handlerTY.InsecureShareDirWebHandlerPath, // web file insecure share api
		"/api/oauth/login",                       // oauth login api
		"/api/oauth/token",                       // oauth token api
		"/api/plugin/gateway",                    // gateway plugin api
	}
)

// MiddlewareAuthenticationVerification verifies user auth details
func MiddlewareAuthenticationVerification(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		isSecurePrefix := false

		// if signout requested, do signout
		if path == handlerTY.SIGNOUT_PATH {
			doSignOut(w, r)
			handlerUtils.WriteResponse(w, []byte("Signout success"))
			return
		}

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
			// authentication required
			if err, mcApiContext := IsValidToken(r); err == nil {

				// include user details as context
				ctx := context.WithValue(r.Context(), types.MC_API_CONTEXT, mcApiContext)
				reqWithCtx := r.WithContext(ctx)

				next.ServeHTTP(w, reqWithCtx)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			handlerUtils.PostErrorResponse(w, "401 Unauthorized", http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)

	})
}

// steps to verify the authentication
// 1. Verify the token in header
// 2. Verify the token in cookie
func IsValidToken(r *http.Request) (error, *types.McApiContext) {
	token, claims, err := getJwtToken(r)
	if err != nil {
		return err, nil
	}
	if !token.Valid {
		return errors.New("invalid token"), nil
	}

	// verify the validity
	expiresAt := convertor.ToInteger(claims[handlerTY.KeyExpiresAt])
	if time.Now().Unix() >= expiresAt {
		return errors.New("expired token"), nil
	}

	// clear userID header, might be injected from external
	// add userID into request header from here
	r.Header.Del(handlerTY.HeaderUserID)
	if userID, ok := claims[handlerTY.KeyUserID]; ok {
		id, ok := userID.(string)
		if ok {
			r.Header.Set(handlerTY.HeaderUserID, id)
		}
	}

	mcApiContext := types.McApiContext{
		Tenant: "",
		UserID: r.Header.Get(handlerTY.HeaderUserID),
	}

	return nil, &mcApiContext
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

// steps to extract the authentication token
// 1. verify the token in request
// 2. verify the token in header
// 3. verify the token in cookie
func extractJwtToken(r *http.Request) string {
	// 1. verify the token in request
	accessToken := r.URL.Query().Get(handlerTY.AccessToken)
	if accessToken != "" {
		return accessToken
	}

	// 2. verify the token in header
	accessToken = r.Header.Get(handlerTY.HeaderAuthorization)
	if accessToken != "" {
		// normally, Authorization: 'Bearer the_token_xxx'
		if !strings.Contains(accessToken, " ") {
			return accessToken
		}
		strArr := strings.Split(accessToken, " ")
		if len(strArr) == 2 {
			return strArr[1]
		}
	}

	// 3. verify the token in cookie
	authCookie, err := r.Cookie(handlerTY.AUTH_COOKIE_NAME)
	if err == nil {
		return authCookie.Value
	}
	return ""
}

// CreateToken creates a token for a user
func CreateToken(user user.User, expiresIn, svcTokenID string) (string, error) {
	atClaims := jwt.MapClaims{}
	atClaims[handlerTY.KeyAuthorized] = true
	atClaims[handlerTY.KeyUserID] = user.ID
	atClaims[handlerTY.KeyFullName] = user.FullName
	atClaims[handlerTY.KeyServiceTokenID] = svcTokenID

	expiresInDuration := handlerTY.DefaultTokenExpiration

	if expiresIn != "" {
		expiresInReceived, err := time.ParseDuration(expiresIn)
		if err == nil {
			expiresInDuration = expiresInReceived
		} else {
			zap.L().Error("error on parse the duration", zap.String("expiration", expiresIn), zap.Error(err))
		}
	}

	atClaims[handlerTY.KeyExpiresAt] = time.Now().Add(expiresInDuration).Unix()
	at := jwt.NewWithClaims(jwt.SigningMethodHS256, atClaims)
	token, err := at.SignedString(getJwtSecret())
	if err != nil {
		return "", err
	}
	return token, nil
}

// GetUserID returns the logged in user details
func GetUserID(r *http.Request) string {
	return r.Header.Get(handlerTY.HeaderUserID)
}

func getJwtSecret() []byte {
	return []byte(fmt.Sprintf("%s_%s", os.Getenv(handlerTY.EnvJwtAccessSecret), version.Get().HostID))
}

// doSignout clears the cookies
func doSignOut(w http.ResponseWriter, r *http.Request) {
	// remove cookie and redirect to authentication page
	clearCookie := &http.Cookie{
		Name:   handlerTY.AUTH_COOKIE_NAME,
		Path:   "/",
		Domain: handlerUtils.ExtractHost(r.Host),
		MaxAge: -1,
		Value:  "",
	}
	http.SetCookie(w, clearCookie)
}
