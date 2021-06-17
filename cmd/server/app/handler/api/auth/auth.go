package auth

import (
	"io/ioutil"
	"net/http"

	"github.com/gorilla/mux"
	middleware "github.com/mycontroller-org/backend/v2/cmd/server/app/handler/middleware"
	handlerUtils "github.com/mycontroller-org/backend/v2/cmd/server/app/handler/utils"
	userAPI "github.com/mycontroller-org/backend/v2/pkg/api/user"
	json "github.com/mycontroller-org/backend/v2/pkg/json"
	userML "github.com/mycontroller-org/backend/v2/pkg/model/user"
	handlerML "github.com/mycontroller-org/backend/v2/pkg/model/web_handler"
	"github.com/mycontroller-org/backend/v2/pkg/utils/hashed"
)

// RegisterAuthRoutes registers auth api
func RegisterAuthRoutes(router *mux.Router) {
	router.HandleFunc("/api/user/login", login).Methods(http.MethodPost)
	router.HandleFunc("/api/user/profile", profile).Methods(http.MethodGet)
	router.HandleFunc("/api/user/profile", updateProfile).Methods(http.MethodPost)
}

func login(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	d, err := ioutil.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		handlerUtils.PostErrorResponse(w, err.Error(), 500)
		return
	}

	userLogin := handlerML.UserLogin{}

	err = json.Unmarshal(d, &userLogin)
	if err != nil {
		handlerUtils.PostErrorResponse(w, err.Error(), 500)
		return
	}

	// get user details
	userDB, err := userAPI.GetByUsername(userLogin.Username)
	if err != nil {
		handlerUtils.PostErrorResponse(w, "Invalid user or password!", 401)
		return
	}

	//compare the user from the request, with the one we defined:
	if userLogin.Username != userDB.Username || !hashed.IsValidPassword(userDB.Password, userLogin.Password) {
		handlerUtils.PostErrorResponse(w, "Please provide valid login details", 401)
		return
	}
	token, err := middleware.CreateToken(userDB, userLogin.ExpiresIn)
	if err != nil {
		handlerUtils.PostErrorResponse(w, err.Error(), 500)
		return
	}

	// update in cookies
	// expiration := time.Now().Add(7 * 24 * time.Hour)
	// tokenCookie := http.Cookie{Name: "authToken", Value: token, Expires: expiration, Path: "/"}
	// http.SetCookie(w, &tokenCookie)

	tokenResponse := &handlerML.JwtTokenResponse{
		ID:       userDB.ID,
		Username: userDB.Username,
		Email:    userDB.Email,
		FullName: userDB.FullName,
		Token:    token,
	}
	handlerUtils.PostSuccessResponse(w, tokenResponse)
}

func profile(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	userID := middleware.GetUserID(r)
	if userID == "" {
		handlerUtils.PostErrorResponse(w, "UserID missing in the request", 400)
		return
	}

	user, err := userAPI.GetByID(userID)
	if err != nil {
		handlerUtils.PostErrorResponse(w, err.Error(), 400)
	}
	handlerUtils.PostSuccessResponse(w, &user)
}

func updateProfile(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	userID := middleware.GetUserID(r)
	if userID == "" {
		handlerUtils.PostErrorResponse(w, "UserID missing in the request", 400)
		return
	}

	user, err := userAPI.GetByID(userID)
	if err != nil {
		handlerUtils.PostErrorResponse(w, err.Error(), 400)
	}

	entity := &userML.UserProfileUpdate{}
	err = handlerUtils.LoadEntity(w, r, entity)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	if user.ID != entity.ID {
		http.Error(w, "You can not change ID", 400)
		return
	}

	err = userAPI.UpdateProfile(entity)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
}
