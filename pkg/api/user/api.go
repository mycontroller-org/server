package user

import (
	"errors"
	"strings"
	"time"

	"github.com/mycontroller-org/server/v2/pkg/model"
	userML "github.com/mycontroller-org/server/v2/pkg/model/user"
	stg "github.com/mycontroller-org/server/v2/pkg/service/database/storage"
	"github.com/mycontroller-org/server/v2/pkg/utils"
	"github.com/mycontroller-org/server/v2/pkg/utils/hashed"
	stgML "github.com/mycontroller-org/server/v2/plugin/database/storage"
)

// List by filter and pagination
func List(filters []stgML.Filter, pagination *stgML.Pagination) (*stgML.Result, error) {
	result := make([]userML.User, 0)
	return stg.SVC.Find(model.EntityUser, &result, filters, pagination)
}

// Get returns a item
func Get(filters []stgML.Filter) (userML.User, error) {
	result := userML.User{}
	err := stg.SVC.FindOne(model.EntityUser, &result, filters)
	return result, err
}

// GetByID returns a item
func GetByID(ID string) (userML.User, error) {
	result := userML.User{}
	filters := []stgML.Filter{
		{Key: model.KeyID, Value: ID},
	}
	err := stg.SVC.FindOne(model.EntityUser, &result, filters)
	return result, err
}

// GetByUsername returns a item
func GetByUsername(username string) (userML.User, error) {
	result := userML.User{}
	filters := []stgML.Filter{
		{Key: model.KeyUsername, Value: username},
	}
	err := stg.SVC.FindOne(model.EntityUser, &result, filters)
	return result, err
}

// GetByEmail returns a item
func GetByEmail(email string) (userML.User, error) {
	result := userML.User{}
	filters := []stgML.Filter{
		{Key: model.KeyEmail, Value: email},
	}
	err := stg.SVC.FindOne(model.EntityUser, &result, filters)
	return result, err
}

// Save config into disk
func Save(user *userML.User) error {
	if user.ID == "" {
		user.ID = utils.RandUUID()
	}
	filters := []stgML.Filter{
		{Key: model.KeyID, Value: user.ID},
	}
	user.ModifiedOn = time.Now()
	return stg.SVC.Upsert(model.EntityUser, user, filters)
}

// Delete items
func Delete(IDs []string) (int64, error) {
	filters := []stgML.Filter{{Key: model.KeyID, Operator: stgML.OperatorIn, Value: IDs}}
	return stg.SVC.Delete(model.EntityUser, filters)
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

	// if user.Username != userData.Username {
	// 	return errors.New("username not matching")
	// }

	if userData.CurrentPassword == "" || !hashed.IsValidPassword(user.Password, userData.CurrentPassword) {
		return errors.New("invalid current password")
	}

	newPassword := strings.TrimSpace(userData.NewPassword)
	hashedPassword := user.Password

	if newPassword != "" {
		if newPassword != userData.ConfirmPassword {
			return errors.New("new password and confirm password not matching")
		}
		hashedNewPassword, err := hashed.GenerateHash(newPassword)
		if err != nil {
			return err
		}
		hashedPassword = hashedNewPassword
	}

	user.Password = hashedPassword
	user.Email = userData.Email
	user.FullName = userData.FullName
	user.Labels = userData.Labels

	return Save(&user)
}
