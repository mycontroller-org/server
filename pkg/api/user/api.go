package user

import (
	"errors"
	"strings"
	"time"

	"github.com/mycontroller-org/server/v2/pkg/service/configuration"
	"github.com/mycontroller-org/server/v2/pkg/store"
	types "github.com/mycontroller-org/server/v2/pkg/types"
	userTY "github.com/mycontroller-org/server/v2/pkg/types/user"
	"github.com/mycontroller-org/server/v2/pkg/utils"
	"github.com/mycontroller-org/server/v2/pkg/utils/hashed"
	storageTY "github.com/mycontroller-org/server/v2/plugin/database/storage/type"
)

// List by filter and pagination
func List(filters []storageTY.Filter, pagination *storageTY.Pagination) (*storageTY.Result, error) {
	result := make([]userTY.User, 0)
	return store.STORAGE.Find(types.EntityUser, &result, filters, pagination)
}

// Get returns a item
func Get(filters []storageTY.Filter) (userTY.User, error) {
	result := userTY.User{}
	err := store.STORAGE.FindOne(types.EntityUser, &result, filters)
	return result, err
}

// GetByID returns a item
func GetByID(ID string) (userTY.User, error) {
	result := userTY.User{}
	filters := []storageTY.Filter{
		{Key: types.KeyID, Value: ID},
	}
	err := store.STORAGE.FindOne(types.EntityUser, &result, filters)
	return result, err
}

// GetByUsername returns a item
func GetByUsername(username string) (userTY.User, error) {
	result := userTY.User{}
	filters := []storageTY.Filter{
		{Key: types.KeyUsername, Value: username},
	}
	err := store.STORAGE.FindOne(types.EntityUser, &result, filters)
	return result, err
}

// GetByEmail returns a item
func GetByEmail(email string) (userTY.User, error) {
	result := userTY.User{}
	filters := []storageTY.Filter{
		{Key: types.KeyEmail, Value: email},
	}
	err := store.STORAGE.FindOne(types.EntityUser, &result, filters)
	return result, err
}

// Save config into disk
func Save(user *userTY.User) error {
	if user.ID == "" {
		user.ID = utils.RandUUID()
	}
	filters := []storageTY.Filter{
		{Key: types.KeyID, Value: user.ID},
	}
	if !configuration.PauseModifiedOnUpdate.IsSet() {
		user.ModifiedOn = time.Now()
	}
	return store.STORAGE.Upsert(types.EntityUser, user, filters)
}

// Delete items
func Delete(IDs []string) (int64, error) {
	filters := []storageTY.Filter{{Key: types.KeyID, Operator: storageTY.OperatorIn, Value: IDs}}
	return store.STORAGE.Delete(types.EntityUser, filters)
}

// UpdateProfile updates the user profile
func UpdateProfile(userData *userTY.UserProfileUpdate) error {
	if userData.ID == "" {
		return errors.New("user id can not be empty")
	}
	user, err := GetByID(userData.ID)
	if err != nil {
		return err
	}

	if userData.CurrentPassword == "" || !hashed.IsValidPassword(user.Password, userData.CurrentPassword) {
		return errors.New("invalid current password")
	}

	newPassword := strings.TrimSpace(userData.NewPassword)
	hashedPassword := user.Password

	if newPassword != "" {
		if newPassword != userData.ConfirmPassword {
			return errors.New("new password and confirm password are not matching")
		}
		hashedNewPassword, err := hashed.GenerateHash(newPassword)
		if err != nil {
			return err
		}
		hashedPassword = hashedNewPassword
	}

	user.Password = hashedPassword

	if userData.Username != "" {
		user.Username = userData.Username
	}
	if userData.Email != "" {
		user.Email = userData.Email
	}

	user.FullName = userData.FullName
	user.Labels = userData.Labels

	return Save(&user)
}
