package user

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	policyAPI "github.com/mycontroller-org/server/v2/pkg/api/policy"
	types "github.com/mycontroller-org/server/v2/pkg/types"
	policyTY "github.com/mycontroller-org/server/v2/pkg/types/policy"
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

func (u *UserAPI) notifyCache(user *userTY.User) {
	if user == nil {
		return
	}
	// keep access-control cache in sync
	policyAPI.New(u.ctx, u.logger, u.storage).NotifyUserUpdated(user)
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
	if user.Policies == nil {
		user.Policies = []string{}
	}
	// default new users without policies get admin only when list is empty on first create -
	// callers should set policies; migration assigns admin for empty.
	filters := []storageTY.Filter{
		{Key: types.KeyID, Value: user.ID},
	}
	user.ModifiedOn = time.Now()

	if err := u.storage.Upsert(types.EntityUser, user, filters); err != nil {
		return err
	}
	u.notifyCache(user)
	return nil
}

// Delete items
func (u *UserAPI) Delete(IDs []string) (int64, error) {
	filters := []storageTY.Filter{{Key: types.KeyID, Operator: storageTY.OperatorIn, Value: IDs}}
	n, err := u.storage.Delete(types.EntityUser, filters)
	if err != nil {
		return n, err
	}
	pac := policyAPI.New(u.ctx, u.logger, u.storage)
	for _, id := range IDs {
		pac.NotifyUserDeleted(id)
	}
	return n, nil
}

// Create creates a new user with plain password and optional policies.
func (u *UserAPI) Create(user *userTY.User, plainPassword string) error {
	if user.Username == "" {
		return errors.New("username can not be empty")
	}
	if plainPassword == "" {
		return errors.New("password can not be empty")
	}
	hashedPassword, err := hashed.GenerateHash(plainPassword)
	if err != nil {
		return err
	}
	user.Password = hashedPassword
	if len(user.Policies) == 0 {
		user.Policies = []string{policyTY.PolicyReadOnly}
	}
	user.ID = ""
	return u.Save(user)
}

// SaveAdmin updates user including disabled flag and policies (admin path).
func (u *UserAPI) SaveAdmin(update *userTY.UserAdminUpdate) error {
	if update.ID == "" {
		return errors.New("user id can not be empty")
	}
	user, err := u.GetByID(update.ID)
	if err != nil {
		return err
	}
	if update.Username != "" {
		user.Username = update.Username
	}
	if update.Email != "" {
		user.Email = update.Email
	}
	user.FullName = update.FullName
	if update.Disabled != nil {
		user.Disabled = *update.Disabled
	}
	if update.Policies != nil {
		user.Policies = update.Policies
	}
	if update.Labels != nil {
		user.Labels = update.Labels
	}
	if strings.TrimSpace(update.Password) != "" {
		hashedPassword, err := hashed.GenerateHash(update.Password)
		if err != nil {
			return err
		}
		user.Password = hashedPassword
	}
	return u.Save(&user)
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
	// profile update does not change Disabled or Policies

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
	if input.Policies == nil {
		input.Policies = []string{}
	}

	filters := []storageTY.Filter{
		{Key: types.KeyID, Value: input.ID},
	}
	if err := u.storage.Upsert(types.EntityUser, &input, filters); err != nil {
		return err
	}
	u.notifyCache(&input)
	return nil
}

func (u *UserAPI) GetEntityInterface() interface{} {
	return userTY.User{}
}
