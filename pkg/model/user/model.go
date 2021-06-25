package user

import (
	"time"

	json "github.com/mycontroller-org/server/v2/pkg/json"
	"github.com/mycontroller-org/server/v2/pkg/model/cmap"
)

// User model
type User struct {
	ID         string               `json:"id"`
	Username   string               `json:"username"`
	Email      string               `json:"email"`
	Password   string               `json:"password"`
	FullName   string               `json:"fullName"`
	Labels     cmap.CustomStringMap `json:"labels"`
	ModifiedOn time.Time            `json:"modifiedOn"`
}

// MarshalJSON implementation
func (u *User) MarshalJSON() ([]byte, error) {
	type user User // prevent recursion
	x := user(*u)
	x.Password = ""
	return json.Marshal(x)
}

// UserWithPassword used to keep the password on json export
type UserWithPassword User

// UserProfileUpdate model, used to update user profile
type UserProfileUpdate struct {
	ID              string               `json:"id"`
	Username        string               `json:"username"`
	Email           string               `json:"email"`
	CurrentPassword string               `json:"currentPassword"`
	NewPassword     string               `json:"newPassword"`
	ConfirmPassword string               `json:"confirmPassword"`
	FullName        string               `json:"fullName"`
	Labels          cmap.CustomStringMap `json:"labels"`
}
