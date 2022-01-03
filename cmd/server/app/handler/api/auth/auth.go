package auth

import (
	"net/http"
	"time"

	"github.com/gorilla/mux"
	middleware "github.com/mycontroller-org/server/v2/cmd/server/app/handler/middleware"
	handlerUtils "github.com/mycontroller-org/server/v2/cmd/server/app/handler/utils"
	userAPI "github.com/mycontroller-org/server/v2/pkg/api/user"
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

	// get user details
	userInDB, err := userAPI.GetByUsername(login.Username)
	if err != nil {
		handlerUtils.PostErrorResponse(w, "Invalid credentails", http.StatusUnauthorized)
		return
	}

	// compare the user from the request, with the one we have in database
	if login.Username != userInDB.Username || !hashed.IsValidPassword(userInDB.Password, login.Password) {
		handlerUtils.PostErrorResponse(w, "Invalid credentails", http.StatusUnauthorized)
		return
	}
	token, err := middleware.CreateToken(userInDB, login.ExpiresIn)
	if err != nil {
		handlerUtils.PostErrorResponse(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// set authorization cookie
	// cookie validity
	cookieValidity, err := time.ParseDuration(login.ExpiresIn)
	if err != nil { // take default validity
		zap.L().Error("invalid validity", zap.String("input", login.ExpiresIn), zap.Error(err))
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
		handlerUtils.PostErrorResponse(w, "UserID missing in the request", http.StatusBadRequest)
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
		handlerUtils.PostErrorResponse(w, "UserID missing in the request", http.StatusBadRequest)
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
		http.Error(w, "You can not change ID", http.StatusBadRequest)
		return
	}

	err = userAPI.UpdateProfile(entity)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
