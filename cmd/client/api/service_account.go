package api

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/mycontroller-org/server/v2/pkg/json"
	"github.com/mycontroller-org/server/v2/pkg/types"
	svcAccountTY "github.com/mycontroller-org/server/v2/pkg/types/service_account"
	"github.com/mycontroller-org/server/v2/pkg/utils"
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
	if id != "" {
		item := &svcAccountTY.ServiceAccount{}
		found, err := c.findResource(API_SERVICE_ACCOUNT_LIST, idFilters(id), item)
		if err != nil {
			return nil, err
		}
		if found {
			if userRef != "" && !serviceAccountMatchesUser(item, c.resolveUserRef(userRef), userRef) {
				return nil, nil
			}
			return item, nil
		}
	}
	if name == "" {
		return nil, nil
	}
	items, err := c.FindServiceAccounts(name, userRef)
	if err != nil {
		return nil, err
	}
	switch len(items) {
	case 0:
		return nil, nil
	case 1:
		return &items[0], nil
	default:
		users := make([]string, 0, len(items))
		for _, item := range items {
			label := item.Username
			if label == "" {
				label = item.UserID
			}
			users = append(users, label)
		}
		return nil, fmt.Errorf("multiple service accounts named %s (users: %s); specify --user", name, strings.Join(users, ", "))
	}
}

func (c *Client) FindServiceAccounts(name, userRef string) ([]svcAccountTY.ServiceAccount, error) {
	filters := []storageTY.Filter{equalFilter(types.KeyName, name)}
	if userID := c.resolveUserRef(userRef); userID != "" {
		filters = append(filters, equalFilter(types.KeyUserID, userID))
	}
	raw, err := c.findResources(API_SERVICE_ACCOUNT_LIST, filters, 1000)
	if err != nil {
		return nil, err
	}
	items := make([]svcAccountTY.ServiceAccount, 0, len(raw))
	for _, data := range raw {
		item := svcAccountTY.ServiceAccount{}
		if err := utils.MapToStruct(utils.TagNameJSON, data, &item); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, nil
}

func serviceAccountMatchesUser(item *svcAccountTY.ServiceAccount, userID, userRef string) bool {
	if item == nil {
		return false
	}
	if userID != "" && (item.UserID == userID || strings.EqualFold(item.Username, userRef)) {
		return true
	}
	return strings.EqualFold(item.Username, userRef) || item.UserID == userRef
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

func (c *Client) ResolveServiceAccountIDs(selectors []string, userRef string) ([]string, error) {
	ids := make([]string, 0, len(selectors))
	missing := make([]string, 0)
	for _, selector := range selectors {
		item, err := c.FindServiceAccount(selector, selector, userRef)
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
