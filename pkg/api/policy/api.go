package policy

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	types "github.com/mycontroller-org/server/v2/pkg/types"
	policyTY "github.com/mycontroller-org/server/v2/pkg/types/policy"
	svcTokenTY "github.com/mycontroller-org/server/v2/pkg/types/service_token"
	userTY "github.com/mycontroller-org/server/v2/pkg/types/user"
	"github.com/mycontroller-org/server/v2/pkg/utils"
	storageTY "github.com/mycontroller-org/server/v2/plugin/database/storage/types"
	"go.uber.org/zap"
)

// singleton-ish cache shared across API instances for process lifetime
var (
	globalCache     *Cache
	globalCacheOnce sync.Once
)

type API struct {
	ctx     context.Context
	logger  *zap.Logger
	storage storageTY.Plugin
	cache   *Cache
}

func New(ctx context.Context, logger *zap.Logger, storage storageTY.Plugin) *API {
	a := &API{
		ctx:     ctx,
		logger:  logger.Named("policy_api"),
		storage: storage,
		cache:   nil,
	}

	// The cache is process wide and read by every request. Install the storage
	// loaders exactly once: rebinding them on each New() (this runs per request,
	// via entities.API.Policy()) would race with the readers.
	globalCacheOnce.Do(func() {
		globalCache = newCache()
		globalCache.setLoaders(
			func(id string) (*userTY.User, error) {
				u, err := a.loadUserFromStorage(id)
				if err != nil {
					return nil, err
				}
				return &u, nil
			},
			func(id string) (*policyTY.Policy, error) {
				p, err := a.loadPolicyFromStorage(id)
				if err != nil {
					return nil, err
				}
				return &p, nil
			},
			func(tokenID string) (*svcTokenTY.ServiceToken, error) {
				t, err := a.loadTokenFromStorage(tokenID)
				if err != nil {
					return nil, err
				}
				return &t, nil
			},
			func() ([]policyTY.Policy, error) {
				return a.listAllPolicies()
			},
		)
	})
	a.cache = globalCache

	return a
}

// Cache exposes the in-memory cache for invalidation from other packages.
func (a *API) Cache() *Cache {
	return a.cache
}

// EnsureBuiltInPolicies creates system policies if missing and warms cache.
// Built-in statements are reset on every start: they are code, not user data.
func (a *API) EnsureBuiltInPolicies() error {
	for _, p := range BuiltInPolicies() {
		cp := p
		cp.System = true
		if err := a.saveSystemPolicy(&cp); err != nil {
			return fmt.Errorf("ensure built-in policy %s: %w", p.ID, err)
		}
	}
	return a.cache.WarmPolicies()
}

// ReportUsersWithoutPolicies logs users that cannot access anything and warms the
// user cache. It deliberately does not grant anything: an empty policy list means
// "no access", and silently promoting such a user to admin on every restart would
// undo an administrator's decision. Pre-RBAC installs are handled once by the
// 2.2.0-1 upgrade (see AssignAdminToUsersWithoutPolicies).
func (a *API) ReportUsersWithoutPolicies() error {
	users, err := a.listAllUsers()
	if err != nil {
		return err
	}
	for i := range users {
		u := users[i]
		a.cache.PutUser(&u)
		if len(u.Policies) == 0 {
			a.logger.Warn("user has no policies attached and cannot access any api",
				zap.String("userId", u.ID), zap.String("username", u.Username))
		}
	}
	return nil
}

// AssignAdminToUsersWithoutPolicies grants the admin policy to users that have none.
// Only for the pre-RBAC migration, where every existing user was effectively an admin.
func (a *API) AssignAdminToUsersWithoutPolicies() error {
	users, err := a.listAllUsers()
	if err != nil {
		return err
	}
	for i := range users {
		u := users[i]
		if len(u.Policies) > 0 {
			a.cache.PutUser(&u)
			continue
		}
		u.Policies = []string{policyTY.PolicyAdmin}
		if err := a.storage.Upsert(types.EntityUser, &u, []storageTY.Filter{{Key: types.KeyID, Value: u.ID}}); err != nil {
			return err
		}
		a.cache.PutUser(&u)
		a.logger.Info("assigned admin policy to pre-rbac user",
			zap.String("userId", u.ID), zap.String("username", u.Username))
	}
	return nil
}

func (a *API) listAllUsers() ([]userTY.User, error) {
	result := make([]userTY.User, 0)
	if _, err := a.storage.Find(types.EntityUser, &result, nil, &storageTY.Pagination{Limit: -1}); err != nil {
		return nil, err
	}
	return result, nil
}

func (a *API) List(filters []storageTY.Filter, pagination *storageTY.Pagination) (*storageTY.Result, error) {
	result := make([]policyTY.Policy, 0)
	return a.storage.Find(types.EntityPolicy, &result, filters, pagination)
}

func (a *API) GetByID(id string) (policyTY.Policy, error) {
	return a.loadPolicyFromStorage(id)
}

func (a *API) loadPolicyFromStorage(id string) (policyTY.Policy, error) {
	result := policyTY.Policy{}
	err := a.storage.FindOne(types.EntityPolicy, &result, []storageTY.Filter{{Key: types.KeyID, Value: id}})
	return result, err
}

func (a *API) loadUserFromStorage(id string) (userTY.User, error) {
	result := userTY.User{}
	err := a.storage.FindOne(types.EntityUser, &result, []storageTY.Filter{{Key: types.KeyID, Value: id}})
	return result, err
}

func (a *API) loadTokenFromStorage(tokenID string) (svcTokenTY.ServiceToken, error) {
	result := svcTokenTY.ServiceToken{}
	err := a.storage.FindOne(types.EntityServiceToken, &result, []storageTY.Filter{{Key: types.KeyTokenID, Value: tokenID}})
	return result, err
}

func (a *API) listAllPolicies() ([]policyTY.Policy, error) {
	result := make([]policyTY.Policy, 0)
	_, err := a.storage.Find(types.EntityPolicy, &result, nil, &storageTY.Pagination{Limit: -1})
	return result, err
}

// Save persists a user authored policy.
//
// System policies are code: their statements are rewritten on every start, so
// accepting edits here would silently discard them. The System flag itself is
// never taken from the request - a client cannot mark a custom policy undeletable
// or create a policy with a built-in id.
func (a *API) Save(policy *policyTY.Policy) error {
	if policy.ID == "" {
		policy.ID = utils.RandID()
	}
	existing, err := a.loadPolicyFromStorage(policy.ID)
	isExisting := err == nil && existing.ID != ""

	if IsBuiltInPolicyID(policy.ID) || (isExisting && existing.System) {
		return fmt.Errorf("%w: %s", ErrSystemPolicyImmutable, policy.ID)
	}
	if policy.System {
		return fmt.Errorf("%w: %s", ErrSystemFlagNotAllowed, policy.ID)
	}
	policy.System = false

	return a.upsert(policy)
}

// saveSystemPolicy persists a built-in policy. Internal use only.
func (a *API) saveSystemPolicy(policy *policyTY.Policy) error {
	policy.System = true
	return a.upsert(policy)
}

func (a *API) upsert(policy *policyTY.Policy) error {
	policy.ModifiedOn = time.Now()
	filters := []storageTY.Filter{{Key: types.KeyID, Value: policy.ID}}
	if err := a.storage.Upsert(types.EntityPolicy, policy, filters); err != nil {
		return err
	}
	a.cache.PutPolicy(policy)
	return nil
}

// Delete removes policies. System policies are protected, and so are policies still
// attached to a user - detaching them silently would leave that user with an
// unresolvable policy id, and a user whose list becomes empty loses all access.
func (a *API) Delete(IDs []string) (int64, error) {
	for _, id := range IDs {
		p, err := a.loadPolicyFromStorage(id)
		if err == nil && p.System {
			return 0, fmt.Errorf("cannot delete system policy: %s", id)
		}
	}

	users, err := a.listAllUsers()
	if err != nil {
		return 0, err
	}
	for _, id := range IDs {
		var attached []string
		for i := range users {
			if utils.ContainsString(users[i].Policies, id) {
				attached = append(attached, users[i].Username)
			}
		}
		if len(attached) > 0 {
			return 0, fmt.Errorf("policy %s is attached to user(s): %s", id, strings.Join(attached, ", "))
		}
	}

	filters := []storageTY.Filter{{Key: types.KeyID, Operator: storageTY.OperatorIn, Value: IDs}}
	n, err := a.storage.Delete(types.EntityPolicy, filters)
	if err != nil {
		return n, err
	}
	for _, id := range IDs {
		a.cache.InvalidatePolicy(id)
	}
	return n, nil
}

func (a *API) Import(data interface{}) error {
	input, ok := data.(policyTY.Policy)
	if !ok {
		return fmt.Errorf("invalid type:%T", data)
	}
	if IsBuiltInPolicyID(input.ID) {
		return nil
	}
	input.System = false
	return a.Save(&input)
}

func (a *API) GetEntityInterface() interface{} {
	return policyTY.Policy{}
}

// NotifyUserUpdated refreshes user cache after user write.
func (a *API) NotifyUserUpdated(user *userTY.User) {
	a.cache.PutUser(user)
}

// NotifyUserDeleted removes user from cache.
func (a *API) NotifyUserDeleted(id string) {
	a.cache.InvalidateUser(id)
}

// NotifyTokenUpdated refreshes token cache.
func (a *API) NotifyTokenUpdated(token *svcTokenTY.ServiceToken) {
	a.cache.PutToken(token)
}

// NotifyTokenDeleted invalidates token cache.
func (a *API) NotifyTokenDeleted(entityID, tokenID string) {
	if tokenID != "" {
		a.cache.InvalidateToken(tokenID)
	}
	if entityID != "" {
		a.cache.InvalidateTokenByEntityID(entityID)
	}
}

// ValidatePoliciesExist ensures all policy ids exist.
func (a *API) ValidatePoliciesExist(ids []string) error {
	for _, id := range ids {
		if _, err := a.cache.GetPolicy(id); err != nil {
			// try storage
			if _, err2 := a.loadPolicyFromStorage(id); err2 != nil {
				return fmt.Errorf("policy not found: %s", id)
			}
		}
	}
	return nil
}

var (
	ErrNotFound              = errors.New("not found")
	ErrSystemPolicyImmutable = errors.New("cannot modify system policy")
	ErrSystemFlagNotAllowed  = errors.New("system flag is not allowed on custom policies")
)
