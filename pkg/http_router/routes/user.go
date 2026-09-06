package routes

import (
	"errors"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
	userTY "github.com/mycontroller-org/server/v2/pkg/types/user"
	handlerUtils "github.com/mycontroller-org/server/v2/pkg/utils/http_handler"
	storageTY "github.com/mycontroller-org/server/v2/plugin/database/storage/types"
)

func (h *Routes) registerUserRoutes() {
	h.router.HandleFunc("/api/user", h.listUsers).Methods(http.MethodGet)
	// Skip reserved auth paths so they are not treated as user ids (e.g. "profile").
	// Match on the request path only - mux.Vars is not reliable inside MatcherFunc.
	h.router.HandleFunc("/api/user/{id}", h.getUser).Methods(http.MethodGet).
		MatcherFunc(func(r *http.Request, _ *mux.RouteMatch) bool {
			path := strings.TrimSuffix(r.URL.Path, "/")
			switch path {
			case "/api/user/profile", "/api/user/login", "/api/user/registration":
				return false
			default:
				return true
			}
		})
	h.router.HandleFunc("/api/user", h.updateUser).Methods(http.MethodPost)
	h.router.HandleFunc("/api/user", h.deleteUsers).Methods(http.MethodDelete)
}

func (h *Routes) listUsers(w http.ResponseWriter, r *http.Request) {
	entityFn := func(f []storageTY.Filter, p *storageTY.Pagination) (interface{}, error) {
		result, err := h.api.User().List(f, p)
		if err != nil {
			return nil, err
		}
		// never expose password hashes over the api
		if users, ok := result.Data.(*[]userTY.User); ok {
			for index := range *users {
				(*users)[index].Password = ""
			}
		}
		return result, nil
	}
	handlerUtils.LoadData(w, r, entityFn)
}

func (h *Routes) getUser(w http.ResponseWriter, r *http.Request) {
	id, ok := mux.Vars(r)["id"]
	if !ok || id == "" {
		http.Error(w, "id should not be an empty", http.StatusBadRequest)
		return
	}
	user, err := h.api.User().GetByID(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	// never expose the password hash over the api
	user.Password = ""
	handlerUtils.PostSuccessResponse(w, &user)
}

func (h *Routes) updateUser(w http.ResponseWriter, r *http.Request) {
	entity := &userTY.UserAdminUpdate{}
	err := handlerUtils.LoadEntity(w, r, entity)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if entity.ID == "" {
		// create new user from admin update payload
		disabled := false
		if entity.Disabled != nil {
			disabled = *entity.Disabled
		}
		user := &userTY.User{
			Username: entity.Username,
			Email:    entity.Email,
			FullName: entity.FullName,
			Disabled: disabled,
			Policies: entity.Policies,
			Labels:   entity.Labels,
		}
		if entity.Password == "" {
			http.Error(w, "password required for new user", http.StatusBadRequest)
			return
		}
		// hash via SaveAdmin path - create then set password
		if err := h.createUserWithPassword(user, entity.Password); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		return
	}
	err = h.api.User().SaveAdmin(entity)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (h *Routes) createUserWithPassword(user *userTY.User, plainPassword string) error {
	return h.api.User().Create(user, plainPassword)
}

func (h *Routes) deleteUsers(w http.ResponseWriter, r *http.Request) {
	IDs := make([]string, 0)
	updateFn := func(f []storageTY.Filter, p *storageTY.Pagination, d []byte) (interface{}, error) {
		if len(IDs) > 0 {
			count, err := h.api.User().Delete(IDs)
			if err != nil {
				return nil, err
			}
			return count, nil
		}
		return nil, errors.New("supply id(s)")
	}
	handlerUtils.UpdateData(w, r, &IDs, updateFn)
}
