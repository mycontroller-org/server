package auth

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"

	"github.com/gorilla/mux"
	middleware "github.com/mycontroller-org/server/v2/cmd/server/app/handler/middleware"
	handlerUtils "github.com/mycontroller-org/server/v2/cmd/server/app/handler/utils"
	userAPI "github.com/mycontroller-org/server/v2/pkg/api/user"
	handlerType "github.com/mycontroller-org/server/v2/pkg/model/web_handler"
	"github.com/mycontroller-org/server/v2/pkg/utils/hashed"
	"go.uber.org/zap"
)

// RegisterAuthRoutes registers auth api
func RegisterOAuthRoutes(router *mux.Router) {
	router.HandleFunc("/api/oauth/login", oAuthLanding).Methods(http.MethodGet)
	router.HandleFunc("/api/oauth/login", oAuthLogin).Methods(http.MethodPost)
	router.HandleFunc("/api/oauth/token", oAuthToken).Methods(http.MethodPost)
}

func oAuthLanding(w http.ResponseWriter, r *http.Request) {
	handlerUtils.WriteResponse(w, []byte(OAuthLoginPageHTML))
}

func oAuthLogin(w http.ResponseWriter, r *http.Request) {
	userInput, err := ioutil.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		handlerUtils.PostErrorResponse(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// validate user details
	credentials, err := url.ParseQuery(string(userInput))
	if err != nil {
		handlerUtils.PostErrorResponse(w, "bad payload format", 400)
		return
	}

	userLogin := handlerType.UserLogin{
		Username:  credentials.Get("username"),
		Password:  credentials.Get("password"),
		ExpiresIn: "168h", // 7 days
	}

	// get user details
	userInDB, err := userAPI.GetByUsername(userLogin.Username)
	if err != nil {
		handlerUtils.PostErrorResponse(w, "invalid user or password!", 401)
		return
	}

	//compare the user from the request, with the one we defined:
	if userLogin.Username != userInDB.Username || !hashed.IsValidPassword(userInDB.Password, userLogin.Password) {
		handlerUtils.PostErrorResponse(w, "Please provide valid login details", 401)
		return
	}
	accessToken, err := middleware.CreateToken(userInDB, userLogin.ExpiresIn)
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

func oAuthToken(w http.ResponseWriter, r *http.Request) {
	inputBytes, err := ioutil.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		zap.L().Error("error on getting data", zap.Error(err))
	}
	// request body sample
	// grant_type=authorization_code&code=hellos-test-token&
	// redirect_uri=https://oauth-redirect.googleusercontent.com/r/test&
	// client_id=https://oauth-redirect.googleusercontent.com/r/test&client_secret=12345

	requestQuery, err := url.ParseQuery(string(inputBytes))
	if err != nil {
		zap.L().Info("error on parsing oauth input", zap.Error(err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	redirectUri := requestQuery.Get("redirect_uri")
	if redirectUri == "" {
		zap.L().Info("redirect_uri not found")
		http.Error(w, "redirect_uri not found", http.StatusBadRequest)
		return
	}

	accessToken := requestQuery.Get("code")
	if accessToken == "" {
		zap.L().Info("code not found in the request")
		http.Error(w, "code not found", http.StatusBadRequest)
		return
	}

	r.Header.Set(handlerType.HeaderAuthorization, accessToken)
	err = middleware.IsValidToken(r)
	if err != nil {
		zap.L().Info("invalid token", zap.Error(err))
		http.Error(w, "invalid token", http.StatusUnauthorized)
		return
	}

	userId := r.Header.Get(handlerType.HeaderUserID)
	if userId == "" {
		zap.L().Info("userId not found")
		http.Error(w, "invalid user", http.StatusUnauthorized)
		return
	}

	userInDB, err := userAPI.GetByID(userId)
	if err != nil {
		zap.L().Info("user not in database", zap.Error(err))
		http.Error(w, "invalid token", http.StatusUnauthorized)
		return
	}

	validity := time.Hour * 24 * 7 // 7 days

	refreshToken, err := middleware.CreateToken(userInDB, validity.String())
	if err != nil {
		zap.L().Info("error on creating token", zap.Error(err))
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
