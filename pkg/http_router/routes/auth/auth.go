package auth

import (
	"net/http"
	"time"

	"github.com/gorilla/mux"
	entityAPI "github.com/mycontroller-org/server/v2/pkg/api/entities"
	middleware "github.com/mycontroller-org/server/v2/pkg/http_router/middleware"
	svcTokenTY "github.com/mycontroller-org/server/v2/pkg/types/service_token"
	userTY "github.com/mycontroller-org/server/v2/pkg/types/user"
	handlerTY "github.com/mycontroller-org/server/v2/pkg/types/web_handler"
	"github.com/mycontroller-org/server/v2/pkg/utils/hashed"
	handlerUtils "github.com/mycontroller-org/server/v2/pkg/utils/http_handler"
	"go.uber.org/zap"
)

type AuthRoutes struct {
	logger *zap.Logger
	api    *entityAPI.API
	router *mux.Router
}

func NewAuthRoutes(logger *zap.Logger, api *entityAPI.API, router *mux.Router) *AuthRoutes {
	return &AuthRoutes{
		logger: logger,
		api:    api,
		router: router,
	}
}

// registers auth api routes
func (a *AuthRoutes) RegisterRoutes() {
	a.router.HandleFunc("/api/user/login", a.login).Methods(http.MethodPost)
	a.router.HandleFunc("/api/user/profile", a.profile).Methods(http.MethodGet)
	a.router.HandleFunc("/api/user/profile", a.updateProfile).Methods(http.MethodPost)
}

func (a *AuthRoutes) login(w http.ResponseWriter, r *http.Request) {
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
		actualToken, err := a.api.ServiceToken().GetByTokenID(parsedToken.ID)
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
		_userInDB, err := a.api.User().GetByID(actualToken.UserID)
		if err != nil {
			handlerUtils.PostErrorResponse(w, "invalid token", http.StatusUnauthorized)
			return
		}
		userInDB = _userInDB
		svcTokenID = parsedToken.ID
	} else { // user based authentication
		// get user details
		_userInDB, err := a.api.User().GetByUsername(login.Username)
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
		a.logger.Warn("invalid validity", zap.String("input", login.ExpiresIn), zap.Error(err))
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

func (a *AuthRoutes) profile(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	userID := middleware.GetUserID(r)
	if userID == "" {
		handlerUtils.PostErrorResponse(w, "userID missing in the request", http.StatusBadRequest)
		return
	}

	user, err := a.api.User().GetByID(userID)
	if err != nil {
		handlerUtils.PostErrorResponse(w, err.Error(), http.StatusBadRequest)
	}
	handlerUtils.PostSuccessResponse(w, &user)
}

func (a *AuthRoutes) updateProfile(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	userID := middleware.GetUserID(r)
	if userID == "" {
		handlerUtils.PostErrorResponse(w, "userID missing in the request", http.StatusBadRequest)
		return
	}

	user, err := a.api.User().GetByID(userID)
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

	err = a.api.User().UpdateProfile(entity)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
