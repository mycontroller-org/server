package handler

import (
	"io/ioutil"
	"net/http"

	"github.com/gorilla/mux"
	userAPI "github.com/mycontroller-org/backend/v2/pkg/api/user"
	json "github.com/mycontroller-org/backend/v2/pkg/json"
	userML "github.com/mycontroller-org/backend/v2/pkg/model/user"
	handlerML "github.com/mycontroller-org/backend/v2/pkg/model/web_handler"
	"github.com/mycontroller-org/backend/v2/pkg/utils/hashed"
)

func registerAuthRoutes(router *mux.Router) {
	router.HandleFunc("/api/user/login", login).Methods(http.MethodPost)
	router.HandleFunc("/api/user/profile", profile).Methods(http.MethodGet)
	router.HandleFunc("/api/user/profile", updateProfile).Methods(http.MethodPost)
}

func login(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	d, err := ioutil.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		postErrorResponse(w, err.Error(), 500)
		return
	}

	userLogin := handlerML.UserLogin{}

	err = json.Unmarshal(d, &userLogin)
	if err != nil {
		postErrorResponse(w, err.Error(), 500)
		return
	}

	// get user details
	userDB, err := userAPI.GetByUsername(userLogin.Username)
	if err != nil {
		postErrorResponse(w, "Invalid user or password!", 401)
		return
	}

	//compare the user from the request, with the one we defined:
	if userLogin.Username != userDB.Username || !hashed.IsValidPassword(userDB.Password, userLogin.Password) {
		postErrorResponse(w, "Please provide valid login details", 401)
		return
	}
	token, err := createToken(userDB, userLogin.Expiration)
	if err != nil {
		postErrorResponse(w, err.Error(), 500)
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
	postSuccessResponse(w, tokenResponse)
}

func profile(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	userID := getUserID(r)
	if userID == "" {
		postErrorResponse(w, "UserID missing in the request", 400)
		return
	}

	user, err := userAPI.GetByID(userID)
	if err != nil {
		postErrorResponse(w, err.Error(), 400)
	}
	postSuccessResponse(w, &user)
}

func updateProfile(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	userID := getUserID(r)
	if userID == "" {
		postErrorResponse(w, "UserID missing in the request", 400)
		return
	}

	user, err := userAPI.GetByID(userID)
	if err != nil {
		postErrorResponse(w, err.Error(), 400)
	}

	entity := &userML.UserProfileUpdate{}
	err = LoadEntity(w, r, entity)
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
