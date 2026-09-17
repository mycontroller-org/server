package api

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/mycontroller-org/server/v2/pkg/json"
	"github.com/mycontroller-org/server/v2/pkg/types"
	policyTY "github.com/mycontroller-org/server/v2/pkg/types/policy"
	userTY "github.com/mycontroller-org/server/v2/pkg/types/user"
	storageTY "github.com/mycontroller-org/server/v2/plugin/database/storage/types"
)

func (c *Client) GetProfile() (*userTY.User, error) {
	res, err := c.executeJson(API_USER_PROFILE, http.MethodGet, nil, nil, nil, http.StatusOK)
	if err != nil {
		return nil, err
	}
	item := &userTY.User{}
	if err := json.Unmarshal(res.Body, item); err != nil {
		return nil, err
	}
	return item, nil
}

func (c *Client) SaveUser(update *userTY.UserAdminUpdate) error {
	return c.saveResource(API_USER_LIST, update)
}

func (c *Client) SavePolicy(policy *policyTY.Policy) error {
	return c.saveResource(API_POLICY_LIST, policy)
}

func (c *Client) FindUser(id, username string) (*userTY.User, error) {
	item := &userTY.User{}
	found, err := c.findResource(API_USER_LIST, idFilters(id), item)
	if err != nil {
		return nil, err
	}
	if found {
		return item, nil
	}
	if username == "" {
		return nil, nil
	}
	item = &userTY.User{}
	found, err = c.findResource(API_USER_LIST, []storageTY.Filter{
		equalFilter(types.KeyUsername, username),
	}, item)
	if err != nil || !found {
		return nil, err
	}
	return item, nil
}

func (c *Client) FindPolicy(id string) (*policyTY.Policy, error) {
	item := &policyTY.Policy{}
	found, err := c.findResource(API_POLICY_LIST, idFilters(id), item)
	if err != nil || !found {
		return nil, err
	}
	return item, nil
}

func (c *Client) EnableUser(selectors ...string) error {
	return c.setUsersEnabled(selectors, true)
}

func (c *Client) DisableUser(selectors ...string) error {
	return c.setUsersEnabled(selectors, false)
}

func (c *Client) setUsersEnabled(selectors []string, enabled bool) error {
	var current *userTY.User
	if !enabled {
		profile, err := c.GetProfile()
		if err != nil {
			return err
		}
		current = profile
	}
	var firstErr error
	updated := 0
	for _, selector := range selectors {
		user, err := c.FindUser(selector, selector)
		if err != nil {
			if firstErr == nil {
				firstErr = err
			}
			continue
		}
		if user == nil {
			if firstErr == nil {
				firstErr = fmt.Errorf("user %s is not present", selector)
			} else {
				firstErr = fmt.Errorf("%w; user %s is not present", firstErr, selector)
			}
			continue
		}
		if !enabled && current != nil && (user.ID == current.ID || strings.EqualFold(user.Username, current.Username)) {
			err := fmt.Errorf("cannot disable the current user %s", user.Username)
			if firstErr == nil {
				firstErr = err
			} else {
				firstErr = fmt.Errorf("%w; %s", firstErr, err)
			}
			continue
		}
		flag := enabled
		if err := c.SaveUser(&userTY.UserAdminUpdate{
			ID:       user.ID,
			Username: user.Username,
			Email:    user.Email,
			FullName: user.FullName,
			Enabled:  &flag,
			Policies: user.Policies,
			Labels:   user.Labels,
		}); err != nil {
			if firstErr == nil {
				firstErr = err
			}
			continue
		}
		updated++
	}
	if updated == 0 && firstErr != nil {
		return firstErr
	}
	return firstErr
}

func (c *Client) ResolveUserIDs(selectors []string) ([]string, error) {
	ids := make([]string, 0, len(selectors))
	missing := make([]string, 0)
	for _, selector := range selectors {
		item, err := c.FindUser(selector, selector)
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
		return ids, fmt.Errorf("user(s) not present: %s", strings.Join(missing, ", "))
	}
	return ids, nil
}
