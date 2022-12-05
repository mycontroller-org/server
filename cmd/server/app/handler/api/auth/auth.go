package auth

import (
	"net/http"
	"time"

	"github.com/gorilla/mux"
	middleware "github.com/mycontroller-org/server/v2/cmd/server/app/handler/middleware"
	handlerUtils "github.com/mycontroller-org/server/v2/cmd/server/app/handler/utils"
	svcTokenAPI "github.com/mycontroller-org/server/v2/pkg/api/service_token"
	userAPI "github.com/mycontroller-org/server/v2/pkg/api/user"
	svcTokenTY "github.com/mycontroller-org/server/v2/pkg/types/service_token"
	userTY "github.com/mycontroller-org/server/v2/pkg/types/user"
	handlerTY "github.com/mycontroller-org/server/v2/pkg/types/web_handler"
	"github.com/mycontroller-org/server/v2/pkg/utils/hashed"
	"go.uber.org/zap"
)

// RegisterAuthRoutes registers auth api
func RegisterAuthRoutes(router *mux.Router) {
	router.HandleFunc("/api/user/login", login).Methods(http.MethodPost)
	router.HandleFunc("/api/user/profile", profile).Methods(http.MethodGet)
	router.HandleFunc("/api/user/profile", updateProfile).Methods(http.MethodPost)
}

func login(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	login := &handlerTY.UserLogin{}

	err := handlerUtils.LoadEntity(w, r, login)
	if err != nil {
		handlerUtils.PostErrorResponse(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var userInDB userTY.User
	var svcTokenID string

	// if token available, it is token based authentication
	if login.SvcToken != "" {
		parsedToken, err := svcTokenTY.ParseToken(login.SvcToken)
		if err != nil {
			handlerUtils.PostErrorResponse(w, "invalid token", http.StatusUnauthorized)
			return
		}

		// get actual token
		actualToken, err := svcTokenAPI.GetByTokenID(parsedToken.ID)
		if err != nil {
			handlerUtils.PostErrorResponse(w, "invalid token", http.StatusUnauthorized)
			return
		}

		// verify validity
		if !actualToken.NeverExpire {
			if actualToken.ExpiresOn.Before(time.Now()) {
				handlerUtils.PostErrorResponse(w, "invalid token", http.StatusUnauthorized)
				return
			}
		}

		// verify token
		if !hashed.IsValidPassword(actualToken.Token.Token, parsedToken.Token) {
			handlerUtils.PostErrorResponse(w, "invalid token", http.StatusUnauthorized)
			return
		}

		// get user details
		_userInDB, err := userAPI.GetByID(actualToken.UserID)
		if err != nil {
			handlerUtils.PostErrorResponse(w, "invalid token", http.StatusUnauthorized)
			return
		}
		userInDB = _userInDB
		svcTokenID = parsedToken.ID
	} else { // user based authentication
		// get user details
		_userInDB, err := userAPI.GetByUsername(login.Username)
		if err != nil {
			handlerUtils.PostErrorResponse(w, "invalid user or password", http.StatusUnauthorized)
			return
		}

		//compare the user from the request, with the one we defined:
		if login.Username != _userInDB.Username || !hashed.IsValidPassword(_userInDB.Password, login.Password) {
			handlerUtils.PostErrorResponse(w, "please provide valid login details", http.StatusUnauthorized)
			return
		}
		userInDB = _userInDB
	}

	token, err := middleware.CreateToken(userInDB, login.ExpiresIn, svcTokenID)
	if err != nil {
		handlerUtils.PostErrorResponse(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// set authorization cookie
	// cookie validity
	cookieValidity, err := time.ParseDuration(login.ExpiresIn)
	if err != nil { // take default validity
		zap.L().Warn("invalid validity", zap.String("input", login.ExpiresIn), zap.Error(err))
		cookieValidity = handlerTY.DefaultTokenExpiration
	}

	generatedCookie := &http.Cookie{
		Name:    handlerTY.AUTH_COOKIE_NAME,
		Path:    "/",
		Domain:  handlerUtils.ExtractHost(r.Host),
		Expires: time.Now().Add(cookieValidity),
		Value:   token,
	}
	http.SetCookie(w, generatedCookie)

	// token response in the response body
	tokenResponse := &handlerTY.JwtTokenResponse{
		ID:       userInDB.ID,
		Username: userInDB.Username,
		Email:    userInDB.Email,
		FullName: userInDB.FullName,
		Token:    token,
	}
	handlerUtils.PostSuccessResponse(w, tokenResponse)
}

func profile(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	userID := middleware.GetUserID(r)
	if userID == "" {
		handlerUtils.PostErrorResponse(w, "userID missing in the request", http.StatusBadRequest)
		return
	}

	user, err := userAPI.GetByID(userID)
	if err != nil {
		handlerUtils.PostErrorResponse(w, err.Error(), http.StatusBadRequest)
	}
	handlerUtils.PostSuccessResponse(w, &user)
}

func updateProfile(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	userID := middleware.GetUserID(r)
	if userID == "" {
		handlerUtils.PostErrorResponse(w, "userID missing in the request", http.StatusBadRequest)
		return
	}

	user, err := userAPI.GetByID(userID)
	if err != nil {
		handlerUtils.PostErrorResponse(w, err.Error(), http.StatusBadRequest)
	}

	entity := &userTY.UserProfileUpdate{}
	err = handlerUtils.LoadEntity(w, r, entity)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if user.ID != entity.ID {
		http.Error(w, "you can not change ID", http.StatusBadRequest)
		return
	}

	err = userAPI.UpdateProfile(entity)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
