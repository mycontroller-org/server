package handler

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/mycontroller-org/backend/v2/pkg/model/user"

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
	token, err := getJwtToken(r)
	if err != nil {
		return err
	}
	if _, ok := token.Claims.(jwt.Claims); !ok && !token.Valid {
		return err
	}
	return nil
}

func getJwtToken(r *http.Request) (*jwt.Token, error) {
	tokenString := extractJwtToken(r)
	if tokenString == "" {
		return nil, errors.New("token not supplied")
	}
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Make sure that the token method conform to "SigningMethodHMAC"
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(os.Getenv("ACCESS_SECRET")), nil
	})
	if err != nil {
		return nil, err
	}
	return token, nil
}

func extractJwtToken(r *http.Request) string {
	authCookie, err := r.Cookie("authToken")
	if err == nil {
		return authCookie.Value
	}

	bearToken := r.Header.Get("Authorization")
	//normally Authorization the_token_xxx
	strArr := strings.Split(bearToken, " ")
	if len(strArr) == 2 {
		return strArr[1]
	}
	return ""
}

func createToken(user user.User) (string, error) {
	var err error
	//Creating Access Token
	os.Setenv("ACCESS_SECRET", "jdnfksdmfksdabcd") //this should be in an env file
	atClaims := jwt.MapClaims{}
	atClaims["authorized"] = true
	atClaims["id"] = user.ID
	atClaims["fullname"] = user.FullName
	atClaims["exp"] = time.Now().Add(time.Hour * 72).Unix()
	at := jwt.NewWithClaims(jwt.SigningMethodHS256, atClaims)
	token, err := at.SignedString([]byte(os.Getenv("ACCESS_SECRET")))
	if err != nil {
		return "", err
	}
	return token, nil
}
