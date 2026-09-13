package api

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/mycontroller-org/server/v2/pkg/json"
	"github.com/mycontroller-org/server/v2/pkg/types"
	svcAccountTY "github.com/mycontroller-org/server/v2/pkg/types/service_account"
	storageTY "github.com/mycontroller-org/server/v2/plugin/database/storage/types"
)

func (c *Client) CreateServiceAccount(account *svcAccountTY.ServiceAccount) (*svcAccountTY.CreateAccountResponse, error) {
	res, err := c.executeJson(API_SERVICE_ACCOUNT_CREATE, http.MethodPost, nil, nil, account, http.StatusOK)
	if err != nil {
		return nil, err
	}
	created := &svcAccountTY.CreateAccountResponse{}
	if err := json.Unmarshal(res.Body, created); err != nil {
		return nil, err
	}
	return created, nil
}

func (c *Client) UpdateServiceAccount(account *svcAccountTY.ServiceAccount) error {
	_, err := c.executeJson(API_SERVICE_ACCOUNT_UPDATE, http.MethodPost, nil, nil, account, http.StatusOK)
	return err
}

func (c *Client) FindServiceAccount(id, name, userRef string) (*svcAccountTY.ServiceAccount, error) {
	item := &svcAccountTY.ServiceAccount{}
	found, err := c.findResource(API_SERVICE_ACCOUNT_LIST, idFilters(id), item)
	if err != nil {
		return nil, err
	}
	if found {
		return item, nil
	}
	if name == "" {
		return nil, nil
	}
	filters := []storageTY.Filter{equalFilter(types.KeyName, name)}
	if userID := c.resolveUserRef(userRef); userID != "" {
		filters = append(filters, equalFilter(types.KeyUserID, userID))
	}
	item = &svcAccountTY.ServiceAccount{}
	found, err = c.findResource(API_SERVICE_ACCOUNT_LIST, filters, item)
	if err != nil || !found {
		return nil, err
	}
	return item, nil
}

func (c *Client) resolveUserRef(userRef string) string {
	userRef = strings.TrimSpace(userRef)
	if userRef == "" {
		return ""
	}
	user, err := c.FindUser(userRef, userRef)
	if err != nil || user == nil {
		return userRef
	}
	return user.ID
}

func (c *Client) ResolveServiceAccountIDs(selectors []string) ([]string, error) {
	ids := make([]string, 0, len(selectors))
	missing := make([]string, 0)
	for _, selector := range selectors {
		item, err := c.FindServiceAccount(selector, selector, "")
		if err != nil {
			return ids, err
		}
		if item == nil {
			missing = append(missing, selector)
			continue
		}
		ids = append(ids, item.ID)
	}
	if len(missing) > 0 {
		return ids, fmt.Errorf("service-account(s) not present: %s", strings.Join(missing, ", "))
	}
	return ids, nil
}
