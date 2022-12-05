package service_token

import (
	"errors"
	"fmt"
	"time"

	"github.com/mycontroller-org/server/v2/pkg/service/configuration"
	"github.com/mycontroller-org/server/v2/pkg/store"
	types "github.com/mycontroller-org/server/v2/pkg/types"
	svcTokenTY "github.com/mycontroller-org/server/v2/pkg/types/service_token"
	"github.com/mycontroller-org/server/v2/pkg/utils"
	"github.com/mycontroller-org/server/v2/pkg/utils/hashed"
	storageTY "github.com/mycontroller-org/server/v2/plugin/database/storage/types"
)

// List by filter and pagination
func List(filters []storageTY.Filter, pagination *storageTY.Pagination) (*storageTY.Result, error) {
	result := make([]svcTokenTY.ServiceToken, 0)
	return store.STORAGE.Find(types.EntityServiceToken, &result, filters, pagination)
}

// Get returns a item
func Get(filters []storageTY.Filter) (svcTokenTY.ServiceToken, error) {
	result := svcTokenTY.ServiceToken{}
	err := store.STORAGE.FindOne(types.EntityServiceToken, &result, filters)
	return result, err
}

// GetByID returns a item
func GetByID(ID string) (svcTokenTY.ServiceToken, error) {
	result := svcTokenTY.ServiceToken{}
	filters := []storageTY.Filter{
		{Key: types.KeyID, Value: ID},
	}
	err := store.STORAGE.FindOne(types.EntityServiceToken, &result, filters)
	return result, err
}

// GetByUserID returns a item
func GetByUserID(userID string) (svcTokenTY.ServiceToken, error) {
	result := svcTokenTY.ServiceToken{}
	filters := []storageTY.Filter{
		{Key: types.KeyUserID, Value: userID},
	}
	err := store.STORAGE.FindOne(types.EntityServiceToken, &result, filters)
	return result, err
}

// GetByToken returns a item
func GetByToken(token string) (svcTokenTY.ServiceToken, error) {
	result := svcTokenTY.ServiceToken{}
	filters := []storageTY.Filter{
		{Key: types.KeyToken, Value: token},
	}
	err := store.STORAGE.FindOne(types.EntityServiceToken, &result, filters)
	return result, err
}

// Save config into disk
func Save(token *svcTokenTY.ServiceToken) error {
	if token.ID == "" {
		token.ID = utils.RandUUID()
	} else if !configuration.PauseModifiedOnUpdate.IsSet() { // get the existing entity and update token and other fields
		oldToken, err := GetByID(token.ID)
		if err != nil {
			return fmt.Errorf("unable to get token with id:%s, error:%s", token.ID, err.Error())
		}
		token.UserID = oldToken.UserID
		token.Token = oldToken.Token
	}

	filters := []storageTY.Filter{
		{Key: types.KeyID, Value: token.ID},
	}

	return store.STORAGE.Upsert(types.EntityServiceToken, token, filters)
}

// Delete items
func Delete(IDs []string) (int64, error) {
	filters := []storageTY.Filter{{Key: types.KeyID, Operator: storageTY.OperatorIn, Value: IDs}}
	return store.STORAGE.Delete(types.EntityServiceToken, filters)
}

// creates new token
func Create(newToken *svcTokenTY.ServiceToken) (*svcTokenTY.CreateTokenResponse, error) {
	if newToken.UserID == "" {
		return nil, errors.New("user id can not be empty")
	}

	// remove token id, will be generated
	newToken.ID = ""

	// generate new token
	generatedToken := utils.RandIDWithLength(21)
	hashedToken, err := hashed.GenerateHash(generatedToken)
	if err != nil {
		return nil, fmt.Errorf("error on generating hash:%s", err.Error())
	}

	newToken.Token = hashedToken
	newToken.CreatedAt = time.Now()

	err = Save(newToken)
	if err != nil {
		return nil, fmt.Errorf("error on saving token:%s", err.Error())
	}

	// returns generated token
	return &svcTokenTY.CreateTokenResponse{ID: newToken.ID, Token: generatedToken}, nil
}
