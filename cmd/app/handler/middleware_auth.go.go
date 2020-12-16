package handler

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	handlerML "github.com/mycontroller-org/backend/v2/pkg/model/handler"
	"github.com/mycontroller-org/backend/v2/pkg/model/user"
	"go.uber.org/zap"

	"github.com/dgrijalva/jwt-go"
)

var (
	// middler check vars
	verifyPrefix      = "/api/"
	nonRestrictedAPIs = []string{"/api/status", "/api/user/registration", "/api/user/login"}
)

func middlewareAuthenticationVerification(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		if strings.HasPrefix(path, verifyPrefix) {
			for _, aPath := range nonRestrictedAPIs {
				if strings.HasPrefix(path, aPath) {
					next.ServeHTTP(w, r)
					return
				}
			}
			// verify token and allow
			if err := isValidToken(r); err == nil {
				next.ServeHTTP(w, r)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			postErrorResponse(w, "Authentication required", 401)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func isValidToken(r *http.Request) error {
	token, claims, err := getJwtToken(r)
	if err != nil {
		return err
	}
	if _, ok := token.Claims.(jwt.Claims); !ok && !token.Valid {
		return err
	}
	// add userID into request header
	// clear userID header, might be injected from external
	r.Header.Del(handlerML.HeaderUserID)
	if userID, ok := claims[handlerML.KeyUserID]; ok {
		id, ok := userID.(string)
		if ok {
			r.Header.Set(handlerML.HeaderUserID, id)
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
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		// Make sure that the token method conform to "SigningMethodHMAC"
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(os.Getenv(handlerML.EnvJwtAccessSecret)), nil
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

	bearerToken := r.Header.Get(handlerML.HeaderAuthorization)
	// normally, Authorization: Bearer the_token_xxx
	strArr := strings.Split(bearerToken, " ")
	if len(strArr) == 2 {
		return strArr[1]
	}
	return ""
}

func createToken(user user.User, expiration string) (string, error) {
	atClaims := jwt.MapClaims{}
	atClaims[handlerML.KeyAuthorized] = true
	atClaims[handlerML.KeyUserID] = user.ID
	atClaims[handlerML.KeyFullName] = user.FullName

	expirationDuration, err := time.ParseDuration(handlerML.DefaultExpiration)
	if err != nil {
		zap.L().Error("failed to parse", zap.Error(err))
	}
	if expiration != "" {
		exp, err := time.ParseDuration(expiration)
		if err != nil {
			zap.L().Error("failed to parse", zap.String("expiration", expiration), zap.Error(err))
		} else {
			expirationDuration = exp
		}
	}

	atClaims[handlerML.KeyExpiration] = time.Now().Add(expirationDuration).Unix()
	at := jwt.NewWithClaims(jwt.SigningMethodHS256, atClaims)
	token, err := at.SignedString([]byte(os.Getenv(handlerML.EnvJwtAccessSecret)))
	if err != nil {
		return "", err
	}
	return token, nil
}

func getUserID(r *http.Request) string {
	return r.Header.Get(handlerML.HeaderUserID)
}
