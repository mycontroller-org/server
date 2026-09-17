package service_account

import (
	"context"
	"errors"
	"fmt"
	"time"

	policyAPI "github.com/mycontroller-org/server/v2/pkg/api/policy"
	types "github.com/mycontroller-org/server/v2/pkg/types"
	dateTimeTY "github.com/mycontroller-org/server/v2/pkg/types/cusom_datetime"
	policyTY "github.com/mycontroller-org/server/v2/pkg/types/policy"
	svcAccountTY "github.com/mycontroller-org/server/v2/pkg/types/service_account"
	"github.com/mycontroller-org/server/v2/pkg/utils"
	"github.com/mycontroller-org/server/v2/pkg/utils/hashed"
	storageTY "github.com/mycontroller-org/server/v2/plugin/database/storage/types"
	"go.uber.org/zap"
)

type ServiceAccountAPI struct {
	ctx     context.Context
	logger  *zap.Logger
	storage storageTY.Plugin
}

func New(ctx context.Context, logger *zap.Logger, storage storageTY.Plugin) *ServiceAccountAPI {
	return &ServiceAccountAPI{
		ctx:     ctx,
		logger:  logger.Named("service_account_api"),
		storage: storage,
	}
}

func (st *ServiceAccountAPI) notifyCache(token *svcAccountTY.ServiceAccount) {
	if token == nil {
		return
	}
	policyAPI.New(st.ctx, st.logger, st.storage).NotifyTokenUpdated(token)
}

// List by filter and pagination
func (st *ServiceAccountAPI) List(filters []storageTY.Filter, pagination *storageTY.Pagination) (*storageTY.Result, error) {
	result := make([]svcAccountTY.ServiceAccount, 0)
	return st.storage.Find(types.EntityServiceAccount, &result, filters, pagination)
}

// Get returns a item
func (st *ServiceAccountAPI) Get(filters []storageTY.Filter) (svcAccountTY.ServiceAccount, error) {
	result := svcAccountTY.ServiceAccount{}
	err := st.storage.FindOne(types.EntityServiceAccount, &result, filters)
	return result, err
}

// GetByID returns a item
func (st *ServiceAccountAPI) GetByID(ID string) (svcAccountTY.ServiceAccount, error) {
	result := svcAccountTY.ServiceAccount{}
	filters := []storageTY.Filter{
		{Key: types.KeyID, Value: ID},
	}
	err := st.storage.FindOne(types.EntityServiceAccount, &result, filters)
	return result, err
}

// GetByUserID returns a item
func (st *ServiceAccountAPI) GetByUserID(userID string) (svcAccountTY.ServiceAccount, error) {
	result := svcAccountTY.ServiceAccount{}
	filters := []storageTY.Filter{
		{Key: types.KeyUserID, Value: userID},
	}
	err := st.storage.FindOne(types.EntityServiceAccount, &result, filters)
	return result, err
}

// GetByTokenID returns a item
func (st *ServiceAccountAPI) GetByTokenID(tokenID string) (svcAccountTY.ServiceAccount, error) {
	result := svcAccountTY.ServiceAccount{}
	filters := []storageTY.Filter{
		{Key: types.KeyTokenID, Value: tokenID},
	}
	err := st.storage.FindOne(types.EntityServiceAccount, &result, filters)
	return result, err
}

// Save config into disk
func (st *ServiceAccountAPI) Save(token *svcAccountTY.ServiceAccount) error {
	if token.ID == "" {
		token.ID = utils.RandUUID()
	} else { // get the existing entity and update token and other fields
		oldToken, err := st.GetByID(token.ID)
		if err != nil {
			return fmt.Errorf("unable to get service account with id:%s, error:%s", token.ID, err.Error())
		}
		// user tie-up is immutable
		token.UserID = oldToken.UserID
		token.Token = oldToken.Token
	}
	if token.UserID == "" {
		return errors.New("user id can not be empty")
	}
	if token.Statements == nil {
		token.Statements = []policyTY.Statement{}
	}
	if token.NeverExpire {
		token.ExpiresOn = dateTimeTY.CustomDate{}
	}

	filters := []storageTY.Filter{
		{Key: types.KeyID, Value: token.ID},
	}

	if err := st.storage.Upsert(types.EntityServiceAccount, token, filters); err != nil {
		return err
	}
	st.notifyCache(token)
	return nil
}

// Delete items
func (st *ServiceAccountAPI) Delete(IDs []string) (int64, error) {
	// load account token ids for cache invalidation
	pac := policyAPI.New(st.ctx, st.logger, st.storage)
	for _, id := range IDs {
		if t, err := st.GetByID(id); err == nil {
			pac.NotifyTokenDeleted(t.ID, t.Token.ID)
		}
	}
	filters := []storageTY.Filter{{Key: types.KeyID, Operator: storageTY.OperatorIn, Value: IDs}}
	return st.storage.Delete(types.EntityServiceAccount, filters)
}

// creates new service account
func (st *ServiceAccountAPI) Create(newToken *svcAccountTY.ServiceAccount) (*svcAccountTY.CreateAccountResponse, error) {
	if newToken.UserID == "" {
		return nil, errors.New("user id can not be empty")
	}

	// generate new token
	generatedToken := svcAccountTY.GetNewToken()
	hashedToken, err := hashed.GenerateHash(generatedToken.Token)
	if err != nil {
		return nil, fmt.Errorf("error on generating hash:%s", err.Error())
	}

	newToken.Token = svcAccountTY.Token{ID: generatedToken.ID, Token: hashedToken}
	newToken.CreatedOn = time.Now()
	if newToken.Statements == nil {
		newToken.Statements = []policyTY.Statement{}
	}
	if newToken.NeverExpire {
		newToken.ExpiresOn = dateTimeTY.CustomDate{}
	}

	// Keep a caller-supplied id only when it is unused (apply --replace deletes
	// first, then recreates with the same id). Never overwrite an existing row.
	if newToken.ID != "" {
		if _, err := st.GetByID(newToken.ID); err == nil {
			newToken.ID = ""
		}
	}
	if newToken.ID == "" {
		newToken.ID = utils.RandUUID()
	}
	filters := []storageTY.Filter{{Key: types.KeyID, Value: newToken.ID}}
	if err := st.storage.Upsert(types.EntityServiceAccount, newToken, filters); err != nil {
		return nil, fmt.Errorf("error on saving service account:%s", err.Error())
	}
	st.notifyCache(newToken)

	// returns generated token
	return &svcAccountTY.CreateAccountResponse{ID: newToken.ID, Token: generatedToken.GetTokenWithID()}, nil
}

func (st *ServiceAccountAPI) Import(data interface{}) error {
	input, ok := data.(svcAccountTY.ServiceAccount)
	if !ok {
		return fmt.Errorf("invalid type:%T", data)
	}
	if input.ID == "" {
		input.ID = utils.RandUUID()
	}

	filters := []storageTY.Filter{
		{Key: types.KeyID, Value: input.ID},
	}
	if err := st.storage.Upsert(types.EntityServiceAccount, &input, filters); err != nil {
		return err
	}
	st.notifyCache(&input)
	return nil
}

func (st *ServiceAccountAPI) GetEntityInterface() interface{} {
	return svcAccountTY.ServiceAccount{}
}

// Enable sets Enabled on the given accounts.
func (st *ServiceAccountAPI) Enable(ids []string) error {
	return st.setEnabled(ids, true)
}

// Disable clears Enabled on the given accounts.
func (st *ServiceAccountAPI) Disable(ids []string) error {
	return st.setEnabled(ids, false)
}

func (st *ServiceAccountAPI) setEnabled(ids []string, enabled bool) error {
	for _, id := range ids {
		if id == "" {
			continue
		}
		account, err := st.GetByID(id)
		if err != nil {
			return err
		}
		if account.Enabled == enabled {
			continue
		}
		account.Enabled = enabled
		if err := st.Save(&account); err != nil {
			return err
		}
	}
	return nil
}
