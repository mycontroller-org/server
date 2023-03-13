package auth

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/gorilla/mux"
	entityAPI "github.com/mycontroller-org/server/v2/pkg/api/entities"
	middleware "github.com/mycontroller-org/server/v2/pkg/http_router/middleware"
	userTY "github.com/mycontroller-org/server/v2/pkg/types/user"
	handlerTY "github.com/mycontroller-org/server/v2/pkg/types/web_handler"
	"github.com/mycontroller-org/server/v2/pkg/utils/hashed"
	handlerUtils "github.com/mycontroller-org/server/v2/pkg/utils/http_handler"
	"go.uber.org/zap"
)

type OAuthRoutes struct {
	logger *zap.Logger
	api    *entityAPI.API
	router *mux.Router
}

func NewOAuthRoutes(logger *zap.Logger, api *entityAPI.API, router *mux.Router) *OAuthRoutes {
	return &OAuthRoutes{
		logger: logger,
		api:    api,
		router: router,
	}
}

// RegisterAuthRoutes registers auth api
func (oa *OAuthRoutes) RegisterRoutes() {
	oa.router.HandleFunc("/api/oauth/login", oa.landing).Methods(http.MethodGet)
	oa.router.HandleFunc("/api/oauth/login", oa.login).Methods(http.MethodPost)
	oa.router.HandleFunc("/api/oauth/token", oa.token).Methods(http.MethodPost)
	oa.router.HandleFunc("/api/oauth/token-alexa", oa.tokenAlexa).Methods(http.MethodPost)
}

func (oa *OAuthRoutes) landing(w http.ResponseWriter, r *http.Request) {
	handlerUtils.WriteResponse(w, []byte(OAuthLoginPageHTML))
}

func (oa *OAuthRoutes) login(w http.ResponseWriter, r *http.Request) {
	userInput, err := io.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		handlerUtils.PostErrorResponse(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// validate user details
	credentials, err := url.ParseQuery(string(userInput))
	if err != nil {
		handlerUtils.PostErrorResponse(w, "bad payload format", http.StatusBadRequest)
		return
	}

	userLogin := handlerTY.UserLogin{
		Username:  credentials.Get("username"),
		Password:  credentials.Get("password"),
		SvcToken:  credentials.Get("token"),
		ExpiresIn: "168h", // 7 days
	}

	var userInDB userTY.User
	var svcTokenID string

	// if token available, it is token based authentication
	if userLogin.SvcToken != "" {
		// get hashed token
		hashedToken, err := hashed.GenerateHash(userLogin.SvcToken)
		if err != nil {
			handlerUtils.PostErrorResponse(w, "invalid token", http.StatusUnauthorized)
			return
		}

		// verify token
		svcToken, err := oa.api.ServiceToken().GetByTokenID(hashedToken)
		if err != nil {
			handlerUtils.PostErrorResponse(w, "invalid token", http.StatusUnauthorized)
			return
		}

		// verify validity
		if svcToken.ExpiresOn.After(time.Now()) {
			handlerUtils.PostErrorResponse(w, "invalid token", http.StatusUnauthorized)
			return
		}

		// get user details
		_userInDB, err := oa.api.User().GetByID(svcToken.UserID)
		if err != nil {
			handlerUtils.PostErrorResponse(w, "invalid token", http.StatusUnauthorized)
			return
		}
		userInDB = _userInDB
		svcTokenID = svcToken.ID
	} else { // user based authentication
		// get user details
		_userInDB, err := oa.api.User().GetByUsername(userLogin.Username)
		if err != nil {
			handlerUtils.PostErrorResponse(w, "invalid user or password", http.StatusUnauthorized)
			return
		}

		//compare the user from the request, with the one we defined:
		if userLogin.Username != _userInDB.Username || !hashed.IsValidPassword(_userInDB.Password, userLogin.Password) {
			handlerUtils.PostErrorResponse(w, "please provide valid login details", http.StatusUnauthorized)
			return
		}
		userInDB = _userInDB
	}

	accessToken, err := middleware.CreateToken(userInDB, userLogin.ExpiresIn, svcTokenID)
	if err != nil {
		handlerUtils.PostErrorResponse(w, err.Error(), http.StatusInternalServerError)
		return
	}

	responseMap := map[string]string{}
	responseMap["access_token"] = accessToken

	// create response payload
	inputQuery := r.URL.Query()

	responseType := ""
	redirectUri := ""

	for key := range inputQuery {
		switch key {
		case "response_type":
			responseType = inputQuery.Get(key)

		case "redirect_uri":
			redirectUri = inputQuery.Get(key)

		case "state":
			responseMap[key] = inputQuery.Get(key)

		default:
			// noop
		}
	}

	if redirectUri == "" {
		http.Error(w, "redirect_uri not found", http.StatusInternalServerError)
		return
	}

	if responseType == "" {
		responseType = "bearer"
	}

	responseMap[responseType] = accessToken

	redirectAddress := fmt.Sprintf("%s?", redirectUri)

	delete(responseMap, "scope")

	for key, value := range responseMap {
		redirectAddress += fmt.Sprintf("%s=%s&", key, value)
	}

	redirectAddress = redirectAddress[:len(redirectAddress)-1]
	http.Redirect(w, r, redirectAddress, http.StatusSeeOther)

}

func (oa *OAuthRoutes) token(w http.ResponseWriter, r *http.Request) {
	inputBytes, err := io.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		oa.logger.Error("error on getting data", zap.Error(err))
	}
	// request body sample
	// grant_type=authorization_code&code=hellos-test-token&
	// redirect_uri=https://oauth-redirect.googleusercontent.com/r/test&
	// client_id=https://oauth-redirect.googleusercontent.com/r/test&client_secret=12345

	requestQuery, err := url.ParseQuery(string(inputBytes))
	if err != nil {
		oa.logger.Info("error on parsing oauth input", zap.Error(err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	redirectUri := requestQuery.Get("redirect_uri")
	if redirectUri == "" {
		oa.logger.Info("redirect_uri not found")
		http.Error(w, "redirect_uri not found", http.StatusBadRequest)
		return
	}

	accessToken := requestQuery.Get("code")
	if accessToken == "" {
		oa.logger.Info("code not found in the request")
		http.Error(w, "code not found", http.StatusBadRequest)
		return
	}

	r.Header.Set(handlerTY.HeaderAuthorization, accessToken)
	err, _ = middleware.IsValidToken(r)
	if err != nil {
		oa.logger.Info("invalid token", zap.Error(err))
		http.Error(w, "invalid token", http.StatusUnauthorized)
		return
	}

	userId := r.Header.Get(handlerTY.HeaderUserID)
	if userId == "" {
		oa.logger.Info("userId not found")
		http.Error(w, "invalid user", http.StatusUnauthorized)
		return
	}

	userInDB, err := oa.api.User().GetByID(userId)
	if err != nil {
		oa.logger.Info("user not in database", zap.Error(err))
		http.Error(w, "invalid token", http.StatusUnauthorized)
		return
	}

	validity := time.Hour * 24 * 7 // 7 days

	refreshToken, err := middleware.CreateToken(userInDB, validity.String(), "")
	if err != nil {
		oa.logger.Info("error on creating token", zap.Error(err))
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"access_token":  accessToken,
		"token_type":    "jwt",
		"expires_in":    validity.Seconds(),
		"refresh_token": refreshToken,
	}

	tokensBytes, err := json.Marshal(response)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	handlerUtils.WriteResponse(w, tokensBytes)
}

func (oa *OAuthRoutes) tokenAlexa(w http.ResponseWriter, r *http.Request) {
	err, _ := middleware.IsValidToken(r)
	if err != nil {
		oa.logger.Info("invalid token", zap.Error(err))
		http.Error(w, "invalid token", http.StatusUnauthorized)
		return
	}

	userId := r.Header.Get(handlerTY.HeaderUserID)
	if userId == "" {
		oa.logger.Info("userId not found")
		http.Error(w, "invalid user", http.StatusUnauthorized)
		return
	}

	userInDB, err := oa.api.User().GetByID(userId)
	if err != nil {
		oa.logger.Info("user not in database", zap.Error(err))
		http.Error(w, "invalid token", http.StatusUnauthorized)
		return
	}

	validity := time.Hour * 24 * 7 // 7 days

	refreshToken, err := middleware.CreateToken(userInDB, validity.String(), "")
	if err != nil {
		oa.logger.Info("error on creating token", zap.Error(err))
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"access_token":  "abc",
		"token_type":    "jwt",
		"expires_in":    validity.Seconds(),
		"refresh_token": refreshToken,
	}

	tokensBytes, err := json.Marshal(response)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	handlerUtils.WriteResponse(w, tokensBytes)
}
