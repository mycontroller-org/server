package user

import (
	"time"

	json "github.com/mycontroller-org/server/v2/pkg/json"
	"github.com/mycontroller-org/server/v2/pkg/types/cmap"
)

// User struct
type User struct {
	ID         string               `json:"id" yaml:"id"`
	Username   string               `json:"username" yaml:"username"`
	Email      string               `json:"email" yaml:"email"`
	Password   string               `json:"password" yaml:"password"`
	FullName   string               `json:"fullName" yaml:"fullName"`
	Labels     cmap.CustomStringMap `json:"labels" yaml:"labels"`
	ModifiedOn time.Time            `json:"modifiedOn" yaml:"modifiedOn"`
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

// UserProfileUpdate struct, used to update user profile
type UserProfileUpdate struct {
	ID              string               `json:"id" yaml:"id"`
	Username        string               `json:"username" yaml:"username"`
	Email           string               `json:"email" yaml:"email"`
	CurrentPassword string               `json:"currentPassword" yaml:"currentPassword"`
	NewPassword     string               `json:"newPassword" yaml:"newPassword"`
	ConfirmPassword string               `json:"confirmPassword" yaml:"confirmPassword"`
	FullName        string               `json:"fullName" yaml:"fullName"`
	Labels          cmap.CustomStringMap `json:"labels" yaml:"labels"`
}
