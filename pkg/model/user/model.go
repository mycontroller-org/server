package user

import (
	"time"

	"github.com/mycontroller-org/backend/v2/pkg/model/cmap"
)

// User model
type User struct {
	ID             string               `json:"id"`
	Username       string               `json:"username"`
	Email          string               `json:"email"`
	Password       string               `json:"password"`
	FullName       string               `json:"fullName"`
	Labels         cmap.CustomStringMap `json:"labels"`
	LastModifiedOn time.Time            `json:"lastModifiedOn"`
}
