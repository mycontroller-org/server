package service_token

import (
	"context"
	"errors"
	"fmt"
	"time"

	policyAPI "github.com/mycontroller-org/server/v2/pkg/api/policy"
	types "github.com/mycontroller-org/server/v2/pkg/types"
	svcTokenTY "github.com/mycontroller-org/server/v2/pkg/types/service_token"
	"github.com/mycontroller-org/server/v2/pkg/utils"
	"github.com/mycontroller-org/server/v2/pkg/utils/hashed"
	storageTY "github.com/mycontroller-org/server/v2/plugin/database/storage/types"
	"go.uber.org/zap"
)

type ServiceTokenAPI struct {
	ctx     context.Context
	logger  *zap.Logger
	storage storageTY.Plugin
}

func New(ctx context.Context, logger *zap.Logger, storage storageTY.Plugin) *ServiceTokenAPI {
	return &ServiceTokenAPI{
		ctx:     ctx,
		logger:  logger.Named("service_token_api"),
		storage: storage,
	}
}

func (st *ServiceTokenAPI) notifyCache(token *svcTokenTY.ServiceToken) {
	if token == nil {
		return
	}
	policyAPI.New(st.ctx, st.logger, st.storage).NotifyTokenUpdated(token)
}

// List by filter and pagination
func (st *ServiceTokenAPI) List(filters []storageTY.Filter, pagination *storageTY.Pagination) (*storageTY.Result, error) {
	result := make([]svcTokenTY.ServiceToken, 0)
	return st.storage.Find(types.EntityServiceToken, &result, filters, pagination)
}

// Get returns a item
func (st *ServiceTokenAPI) Get(filters []storageTY.Filter) (svcTokenTY.ServiceToken, error) {
	result := svcTokenTY.ServiceToken{}
	err := st.storage.FindOne(types.EntityServiceToken, &result, filters)
	return result, err
}

// GetByID returns a item
func (st *ServiceTokenAPI) GetByID(ID string) (svcTokenTY.ServiceToken, error) {
	result := svcTokenTY.ServiceToken{}
	filters := []storageTY.Filter{
		{Key: types.KeyID, Value: ID},
	}
	err := st.storage.FindOne(types.EntityServiceToken, &result, filters)
	return result, err
}

// GetByUserID returns a item
func (st *ServiceTokenAPI) GetByUserID(userID string) (svcTokenTY.ServiceToken, error) {
	result := svcTokenTY.ServiceToken{}
	filters := []storageTY.Filter{
		{Key: types.KeyUserID, Value: userID},
	}
	err := st.storage.FindOne(types.EntityServiceToken, &result, filters)
	return result, err
}

// GetByTokenID returns a item
func (st *ServiceTokenAPI) GetByTokenID(tokenID string) (svcTokenTY.ServiceToken, error) {
	result := svcTokenTY.ServiceToken{}
	filters := []storageTY.Filter{
		{Key: types.KeyTokenID, Value: tokenID},
	}
	err := st.storage.FindOne(types.EntityServiceToken, &result, filters)
	return result, err
}

// Save config into disk
func (st *ServiceTokenAPI) Save(token *svcTokenTY.ServiceToken) error {
	if token.ID == "" {
		token.ID = utils.RandUUID()
	} else { // get the existing entity and update token and other fields
		oldToken, err := st.GetByID(token.ID)
		if err != nil {
			return fmt.Errorf("unable to get token with id:%s, error:%s", token.ID, err.Error())
		}
		// user tie-up is immutable
		token.UserID = oldToken.UserID
		token.Token = oldToken.Token
	}
	if token.UserID == "" {
		return errors.New("user id can not be empty")
	}
	if token.Actions == nil {
		token.Actions = []string{}
	}
	if token.Resources == nil {
		token.Resources = []string{}
	}

	filters := []storageTY.Filter{
		{Key: types.KeyID, Value: token.ID},
	}

	if err := st.storage.Upsert(types.EntityServiceToken, token, filters); err != nil {
		return err
	}
	st.notifyCache(token)
	return nil
}

// Delete items
func (st *ServiceTokenAPI) Delete(IDs []string) (int64, error) {
	// load token ids for cache invalidation
	pac := policyAPI.New(st.ctx, st.logger, st.storage)
	for _, id := range IDs {
		if t, err := st.GetByID(id); err == nil {
			pac.NotifyTokenDeleted(t.ID, t.Token.ID)
		}
	}
	filters := []storageTY.Filter{{Key: types.KeyID, Operator: storageTY.OperatorIn, Value: IDs}}
	return st.storage.Delete(types.EntityServiceToken, filters)
}

// creates new token
func (st *ServiceTokenAPI) Create(newToken *svcTokenTY.ServiceToken) (*svcTokenTY.CreateTokenResponse, error) {
	if newToken.UserID == "" {
		return nil, errors.New("user id can not be empty")
	}

	// remove token id, will be generated
	newToken.ID = ""

	// generate new token
	generatedToken := svcTokenTY.GetNewToken()
	hashedToken, err := hashed.GenerateHash(generatedToken.Token)
	if err != nil {
		return nil, fmt.Errorf("error on generating hash:%s", err.Error())
	}

	newToken.Token = svcTokenTY.Token{ID: generatedToken.ID, Token: hashedToken}
	newToken.CreatedOn = time.Now()
	if newToken.Actions == nil {
		newToken.Actions = []string{}
	}
	if newToken.Resources == nil {
		newToken.Resources = []string{}
	}

	// Save would try to reload by ID - for create set ID first then upsert without old-token branch
	if newToken.ID == "" {
		newToken.ID = utils.RandUUID()
	}
	filters := []storageTY.Filter{{Key: types.KeyID, Value: newToken.ID}}
	if err := st.storage.Upsert(types.EntityServiceToken, newToken, filters); err != nil {
		return nil, fmt.Errorf("error on saving token:%s", err.Error())
	}
	st.notifyCache(newToken)

	// returns generated token
	return &svcTokenTY.CreateTokenResponse{ID: newToken.ID, Token: generatedToken.GetTokenWithID()}, nil
}

func (st *ServiceTokenAPI) Import(data interface{}) error {
	input, ok := data.(svcTokenTY.ServiceToken)
	if !ok {
		return fmt.Errorf("invalid type:%T", data)
	}
	if input.ID == "" {
		input.ID = utils.RandUUID()
	}

	filters := []storageTY.Filter{
		{Key: types.KeyID, Value: input.ID},
	}
	if err := st.storage.Upsert(types.EntityServiceToken, &input, filters); err != nil {
		return err
	}
	st.notifyCache(&input)
	return nil
}

func (st *ServiceTokenAPI) GetEntityInterface() interface{} {
	return svcTokenTY.ServiceToken{}
}
