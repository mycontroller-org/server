package upgrade

import (
	"context"
	"errors"
	"fmt"
	"time"

	entitiesAPI "github.com/mycontroller-org/server/v2/pkg/api/entities"
	"github.com/mycontroller-org/server/v2/pkg/types"
	"github.com/mycontroller-org/server/v2/pkg/types/cmap"
	dateTimeTY "github.com/mycontroller-org/server/v2/pkg/types/cusom_datetime"
	policyTY "github.com/mycontroller-org/server/v2/pkg/types/policy"
	svcAccountTY "github.com/mycontroller-org/server/v2/pkg/types/service_account"
	userTY "github.com/mycontroller-org/server/v2/pkg/types/user"
	"github.com/mycontroller-org/server/v2/pkg/utils"
	backupTY "github.com/mycontroller-org/server/v2/plugin/database/storage/backup"
	storageTY "github.com/mycontroller-org/server/v2/plugin/database/storage/types"
	"go.uber.org/zap"
)

// from version 2.3.0, user and service account use enabled (same as other resources)
// instead of disabled. Handle restore from a previous backup and migrate stored rows.

type userOld_2_2_0 struct {
	ID         string               `json:"id" yaml:"id"`
	Username   string               `json:"username" yaml:"username"`
	Email      string               `json:"email" yaml:"email"`
	Password   string               `json:"password" yaml:"password"`
	FullName   string               `json:"fullName" yaml:"fullName"`
	Disabled   bool                 `json:"disabled" yaml:"disabled"`
	Policies   []string             `json:"policies" yaml:"policies"`
	Labels     cmap.CustomStringMap `json:"labels" yaml:"labels"`
	CreatedOn  time.Time            `json:"createdOn" yaml:"createdOn"`
	ModifiedOn time.Time            `json:"modifiedOn" yaml:"modifiedOn"`
}

type serviceAccountOld_2_2_0 struct {
	ID          string                `json:"id" yaml:"id"`
	UserID      string                `json:"userId" yaml:"userId"`
	Username    string                `json:"username" yaml:"username"`
	Name        string                `json:"name" yaml:"name"`
	Description string                `json:"description" yaml:"description"`
	Token       svcAccountTY.Token    `json:"token" yaml:"token"`
	Disabled    bool                  `json:"disabled" yaml:"disabled"`
	NeverExpire bool                  `json:"neverExpire" yaml:"neverExpire"`
	ExpiresOn   dateTimeTY.CustomDate `json:"expiresOn" yaml:"expiresOn"`
	Statements  []policyTY.Statement  `json:"statements" yaml:"statements"`
	Labels      cmap.CustomStringMap  `json:"labels" yaml:"labels"`
	CreatedOn   time.Time             `json:"createdOn" yaml:"createdOn"`
}

func upgrade_2_3_0__1(ctx context.Context, logger *zap.Logger, storage storageTY.Plugin, api *entitiesAPI.API) error {
	if err := migrateUsersEnabled(logger, storage); err != nil {
		return err
	}
	return migrateServiceAccountsEnabled(logger, storage)
}

func migrateUsersEnabled(logger *zap.Logger, storage storageTY.Plugin) error {
	filters := []storageTY.Filter{}
	pagination := &storageTY.Pagination{}
	recordLimit := int64(20)
	offset := int64(0)
	updated := 0
	for {
		pagination.Offset = offset
		pagination.Limit = recordLimit
		rows := make([]userOld_2_2_0, 0)
		result, err := storage.Find(types.EntityUser, &rows, filters, pagination)
		if err != nil {
			logger.Error("error on getting users", zap.Error(err))
			return err
		}
		data, ok := result.Data.(*[]userOld_2_2_0)
		if !ok {
			logger.Error("received invalid type", zap.String("actualType", fmt.Sprintf("%T", result.Data)))
			return errors.New("received invalid type")
		}

		for _, row := range *data {
			user := userTY.User{
				ID:         row.ID,
				Username:   row.Username,
				Email:      row.Email,
				Password:   row.Password,
				FullName:   row.FullName,
				Enabled:    !row.Disabled,
				Policies:   row.Policies,
				Labels:     row.Labels,
				CreatedOn:  row.CreatedOn,
				ModifiedOn: row.ModifiedOn,
			}
			if err := storage.Upsert(types.EntityUser, &user, []storageTY.Filter{{Key: types.KeyID, Value: user.ID}}); err != nil {
				logger.Error("error on migrating user enabled", zap.String("id", user.ID), zap.Error(err))
				return err
			}
			updated++
		}

		offset += recordLimit
		if result.Count < offset {
			break
		}
	}
	if updated > 0 {
		logger.Info("migrated users to enabled", zap.Int("count", updated))
	}
	return nil
}

func migrateServiceAccountsEnabled(logger *zap.Logger, storage storageTY.Plugin) error {
	filters := []storageTY.Filter{}
	pagination := &storageTY.Pagination{}
	recordLimit := int64(20)
	offset := int64(0)
	updated := 0
	for {
		pagination.Offset = offset
		pagination.Limit = recordLimit
		rows := make([]serviceAccountOld_2_2_0, 0)
		result, err := storage.Find(types.EntityServiceAccount, &rows, filters, pagination)
		if err != nil {
			logger.Error("error on getting service accounts", zap.Error(err))
			return err
		}
		data, ok := result.Data.(*[]serviceAccountOld_2_2_0)
		if !ok {
			logger.Error("received invalid type", zap.String("actualType", fmt.Sprintf("%T", result.Data)))
			return errors.New("received invalid type")
		}

		for _, row := range *data {
			account := svcAccountTY.ServiceAccount{
				ID:          row.ID,
				UserID:      row.UserID,
				Username:    row.Username,
				Name:        row.Name,
				Description: row.Description,
				Token:       row.Token,
				Enabled:     !row.Disabled,
				NeverExpire: row.NeverExpire,
				ExpiresOn:   row.ExpiresOn,
				Statements:  row.Statements,
				Labels:      row.Labels,
				CreatedOn:   row.CreatedOn,
			}
			if err := storage.Upsert(types.EntityServiceAccount, &account, []storageTY.Filter{{Key: types.KeyID, Value: account.ID}}); err != nil {
				logger.Error("error on migrating service account enabled", zap.String("id", account.ID), zap.Error(err))
				return err
			}
			updated++
		}

		offset += recordLimit
		if result.Count < offset {
			break
		}
	}
	if updated > 0 {
		logger.Info("migrated service accounts to enabled", zap.Int("count", updated))
	}
	return nil
}

func updateRestoreApiMap_2_3_0(ctx context.Context, logger *zap.Logger, storage storageTY.Plugin, apiMap map[string]backupTY.Backup) (map[string]backupTY.Backup, error) {
	apiMap[types.EntityUser] = &userApiOld_2_2_0{
		ctx:     ctx,
		logger:  logger,
		storage: storage,
	}
	apiMap[types.EntityServiceAccount] = &serviceAccountApiOld_2_2_0{
		ctx:     ctx,
		logger:  logger,
		storage: storage,
	}
	return apiMap, nil
}

type userApiOld_2_2_0 struct {
	ctx     context.Context
	logger  *zap.Logger
	storage storageTY.Plugin
}

func (u *userApiOld_2_2_0) Import(data interface{}) error {
	input, ok := data.(userOld_2_2_0)
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

func (u *userApiOld_2_2_0) GetEntityInterface() interface{} {
	return userOld_2_2_0{}
}

func (u *userApiOld_2_2_0) List(filters []storageTY.Filter, pagination *storageTY.Pagination) (*storageTY.Result, error) {
	return nil, errors.New("this method not implemented")
}

type serviceAccountApiOld_2_2_0 struct {
	ctx     context.Context
	logger  *zap.Logger
	storage storageTY.Plugin
}

func (s *serviceAccountApiOld_2_2_0) Import(data interface{}) error {
	input, ok := data.(serviceAccountOld_2_2_0)
	if !ok {
		return fmt.Errorf("invalid type:%T", data)
	}
	if input.ID == "" {
		input.ID = utils.RandUUID()
	}
	filters := []storageTY.Filter{
		{Key: types.KeyID, Value: input.ID},
	}
	return s.storage.Upsert(types.EntityServiceAccount, &input, filters)
}

func (s *serviceAccountApiOld_2_2_0) GetEntityInterface() interface{} {
	return serviceAccountOld_2_2_0{}
}

func (s *serviceAccountApiOld_2_2_0) List(filters []storageTY.Filter, pagination *storageTY.Pagination) (*storageTY.Result, error) {
	return nil, errors.New("this method not implemented")
}
