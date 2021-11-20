package user

import (
	"errors"
	"strings"
	"time"

	"github.com/mycontroller-org/server/v2/pkg/model"
	userML "github.com/mycontroller-org/server/v2/pkg/model/user"
	"github.com/mycontroller-org/server/v2/pkg/service/configuration"
	"github.com/mycontroller-org/server/v2/pkg/store"
	"github.com/mycontroller-org/server/v2/pkg/utils"
	"github.com/mycontroller-org/server/v2/pkg/utils/hashed"
	stgType "github.com/mycontroller-org/server/v2/plugin/database/storage/type"
)

// List by filter and pagination
func List(filters []stgType.Filter, pagination *stgType.Pagination) (*stgType.Result, error) {
	result := make([]userML.User, 0)
	return store.STORAGE.Find(model.EntityUser, &result, filters, pagination)
}

// Get returns a item
func Get(filters []stgType.Filter) (userML.User, error) {
	result := userML.User{}
	err := store.STORAGE.FindOne(model.EntityUser, &result, filters)
	return result, err
}

// GetByID returns a item
func GetByID(ID string) (userML.User, error) {
	result := userML.User{}
	filters := []stgType.Filter{
		{Key: model.KeyID, Value: ID},
	}
	err := store.STORAGE.FindOne(model.EntityUser, &result, filters)
	return result, err
}

// GetByUsername returns a item
func GetByUsername(username string) (userML.User, error) {
	result := userML.User{}
	filters := []stgType.Filter{
		{Key: model.KeyUsername, Value: username},
	}
	err := store.STORAGE.FindOne(model.EntityUser, &result, filters)
	return result, err
}

// GetByEmail returns a item
func GetByEmail(email string) (userML.User, error) {
	result := userML.User{}
	filters := []stgType.Filter{
		{Key: model.KeyEmail, Value: email},
	}
	err := store.STORAGE.FindOne(model.EntityUser, &result, filters)
	return result, err
}

// Save config into disk
func Save(user *userML.User) error {
	if user.ID == "" {
		user.ID = utils.RandUUID()
	}
	filters := []stgType.Filter{
		{Key: model.KeyID, Value: user.ID},
	}
	if !configuration.PauseModifiedOnUpdate.IsSet() {
		user.ModifiedOn = time.Now()
	}
	return store.STORAGE.Upsert(model.EntityUser, user, filters)
}

// Delete items
func Delete(IDs []string) (int64, error) {
	filters := []stgType.Filter{{Key: model.KeyID, Operator: stgType.OperatorIn, Value: IDs}}
	return store.STORAGE.Delete(model.EntityUser, filters)
}

// UpdateProfile updates the user profile
func UpdateProfile(userData *userML.UserProfileUpdate) error {
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
