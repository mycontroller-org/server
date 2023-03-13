package user

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	types "github.com/mycontroller-org/server/v2/pkg/types"
	userTY "github.com/mycontroller-org/server/v2/pkg/types/user"
	"github.com/mycontroller-org/server/v2/pkg/utils"
	"github.com/mycontroller-org/server/v2/pkg/utils/hashed"
	storageTY "github.com/mycontroller-org/server/v2/plugin/database/storage/types"
	"go.uber.org/zap"
)

type UserAPI struct {
	ctx     context.Context
	logger  *zap.Logger
	storage storageTY.Plugin
}

func New(ctx context.Context, logger *zap.Logger, storage storageTY.Plugin) *UserAPI {
	return &UserAPI{
		ctx:     ctx,
		logger:  logger.Named("user_api"),
		storage: storage,
	}
}

// List by filter and pagination
func (u *UserAPI) List(filters []storageTY.Filter, pagination *storageTY.Pagination) (*storageTY.Result, error) {
	result := make([]userTY.User, 0)
	return u.storage.Find(types.EntityUser, &result, filters, pagination)
}

// Get returns a item
func (u *UserAPI) Get(filters []storageTY.Filter) (userTY.User, error) {
	result := userTY.User{}
	err := u.storage.FindOne(types.EntityUser, &result, filters)
	return result, err
}

// GetByID returns a item
func (u *UserAPI) GetByID(ID string) (userTY.User, error) {
	result := userTY.User{}
	filters := []storageTY.Filter{
		{Key: types.KeyID, Value: ID},
	}
	err := u.storage.FindOne(types.EntityUser, &result, filters)
	return result, err
}

// GetByUsername returns a item
func (u *UserAPI) GetByUsername(username string) (userTY.User, error) {
	result := userTY.User{}
	filters := []storageTY.Filter{
		{Key: types.KeyUsername, Value: username},
	}
	err := u.storage.FindOne(types.EntityUser, &result, filters)
	return result, err
}

// GetByEmail returns a item
func (u *UserAPI) GetByEmail(email string) (userTY.User, error) {
	result := userTY.User{}
	filters := []storageTY.Filter{
		{Key: types.KeyEmail, Value: email},
	}
	err := u.storage.FindOne(types.EntityUser, &result, filters)
	return result, err
}

// Save config into disk
func (u *UserAPI) Save(user *userTY.User) error {
	if user.ID == "" {
		user.ID = utils.RandUUID()
	}
	filters := []storageTY.Filter{
		{Key: types.KeyID, Value: user.ID},
	}
	user.ModifiedOn = time.Now()

	return u.storage.Upsert(types.EntityUser, user, filters)
}

// Delete items
func (u *UserAPI) Delete(IDs []string) (int64, error) {
	filters := []storageTY.Filter{{Key: types.KeyID, Operator: storageTY.OperatorIn, Value: IDs}}
	return u.storage.Delete(types.EntityUser, filters)
}

// UpdateProfile updates the user profile
func (u *UserAPI) UpdateProfile(userData *userTY.UserProfileUpdate) error {
	if userData.ID == "" {
		return errors.New("user id can not be empty")
	}
	user, err := u.GetByID(userData.ID)
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

	return u.Save(&user)
}

func (u *UserAPI) Import(data interface{}) error {
	input, ok := data.(userTY.User)
	if !ok {
		return fmt.Errorf("invalid type:%T", data)
	}
	if input.ID == "" {
		input.ID = utils.RandUUID()
	}

	filters := []storageTY.Filter{
		{Key: types.KeyID, Value: input.ID},
	}
	return u.storage.Upsert(types.EntityUser, &input, filters)
}

func (u *UserAPI) GetEntityInterface() interface{} {
	return userTY.User{}
}
