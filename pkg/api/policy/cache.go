package policy

import (
	"sync"

	policyTY "github.com/mycontroller-org/server/v2/pkg/types/policy"
	svcTokenTY "github.com/mycontroller-org/server/v2/pkg/types/service_token"
	userTY "github.com/mycontroller-org/server/v2/pkg/types/user"
)

// Cache holds users, policies, and service tokens in memory for fast auth checks.
// Call Invalidate* after any write so the next request reloads from storage.
type Cache struct {
	mu sync.RWMutex

	users    map[string]*userTY.User             // by user id
	policies map[string]*policyTY.Policy         // by policy id
	tokens   map[string]*svcTokenTY.ServiceToken // by token.Token.ID (token id used in JWT)

	// loaders - set by API
	loadUser   func(id string) (*userTY.User, error)
	loadPolicy func(id string) (*policyTY.Policy, error)
	loadToken  func(tokenID string) (*svcTokenTY.ServiceToken, error)
	loadAllPol func() ([]policyTY.Policy, error)
}

func newCache() *Cache {
	return &Cache{
		users:    make(map[string]*userTY.User),
		policies: make(map[string]*policyTY.Policy),
		tokens:   make(map[string]*svcTokenTY.ServiceToken),
	}
}

func (c *Cache) setLoaders(
	loadUser func(id string) (*userTY.User, error),
	loadPolicy func(id string) (*policyTY.Policy, error),
	loadToken func(tokenID string) (*svcTokenTY.ServiceToken, error),
	loadAllPol func() ([]policyTY.Policy, error),
) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.loadUser = loadUser
	c.loadPolicy = loadPolicy
	c.loadToken = loadToken
	c.loadAllPol = loadAllPol
}

func (c *Cache) userLoader() func(id string) (*userTY.User, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.loadUser
}

func (c *Cache) policyLoader() func(id string) (*policyTY.Policy, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.loadPolicy
}

func (c *Cache) tokenLoader() func(tokenID string) (*svcTokenTY.ServiceToken, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.loadToken
}

// GetUser returns a cached copy or loads from storage.
func (c *Cache) GetUser(id string) (*userTY.User, error) {
	c.mu.RLock()
	if u, ok := c.users[id]; ok {
		cp := *u
		c.mu.RUnlock()
		return &cp, nil
	}
	c.mu.RUnlock()

	load := c.userLoader()
	if load == nil {
		return nil, errCacheNotReady
	}
	u, err := load(id)
	if err != nil {
		return nil, err
	}
	c.mu.Lock()
	c.users[id] = u
	c.mu.Unlock()
	cp := *u
	return &cp, nil
}

// GetPolicy returns a cached policy or loads it.
func (c *Cache) GetPolicy(id string) (*policyTY.Policy, error) {
	c.mu.RLock()
	if p, ok := c.policies[id]; ok {
		cp := *p
		c.mu.RUnlock()
		return &cp, nil
	}
	c.mu.RUnlock()

	load := c.policyLoader()
	if load == nil {
		return nil, errCacheNotReady
	}
	p, err := load(id)
	if err != nil {
		return nil, err
	}
	c.mu.Lock()
	c.policies[id] = p
	c.mu.Unlock()
	cp := *p
	return &cp, nil
}

// GetToken returns a cached service token by raw token id (Token.ID), or loads it.
func (c *Cache) GetToken(tokenID string) (*svcTokenTY.ServiceToken, error) {
	c.mu.RLock()
	if t, ok := c.tokens[tokenID]; ok {
		cp := *t
		c.mu.RUnlock()
		return &cp, nil
	}
	c.mu.RUnlock()

	load := c.tokenLoader()
	if load == nil {
		return nil, errCacheNotReady
	}
	t, err := load(tokenID)
	if err != nil {
		return nil, err
	}
	c.mu.Lock()
	c.tokens[tokenID] = t
	c.mu.Unlock()
	cp := *t
	return &cp, nil
}

// PutUser updates the cache after a write.
func (c *Cache) PutUser(u *userTY.User) {
	if u == nil || u.ID == "" {
		return
	}
	cp := *u
	c.mu.Lock()
	c.users[u.ID] = &cp
	c.mu.Unlock()
}

// PutPolicy updates the cache after a write.
func (c *Cache) PutPolicy(p *policyTY.Policy) {
	if p == nil || p.ID == "" {
		return
	}
	cp := *p
	c.mu.Lock()
	c.policies[p.ID] = &cp
	c.mu.Unlock()
}

// PutToken updates the cache after a write (keyed by Token.ID).
func (c *Cache) PutToken(t *svcTokenTY.ServiceToken) {
	if t == nil {
		return
	}
	cp := *t
	c.mu.Lock()
	if t.Token.ID != "" {
		c.tokens[t.Token.ID] = &cp
	}
	c.mu.Unlock()
}

// InvalidateUser drops a user from cache (call on delete).
func (c *Cache) InvalidateUser(id string) {
	c.mu.Lock()
	delete(c.users, id)
	c.mu.Unlock()
}

// InvalidatePolicy drops a policy from cache.
func (c *Cache) InvalidatePolicy(id string) {
	c.mu.Lock()
	delete(c.policies, id)
	c.mu.Unlock()
}

// InvalidateToken drops a service token from cache by Token.ID.
func (c *Cache) InvalidateToken(tokenID string) {
	c.mu.Lock()
	delete(c.tokens, tokenID)
	c.mu.Unlock()
}

// InvalidateTokenByEntityID removes any cached token matching entity id.
func (c *Cache) InvalidateTokenByEntityID(entityID string) {
	c.mu.Lock()
	for k, t := range c.tokens {
		if t.ID == entityID {
			delete(c.tokens, k)
		}
	}
	c.mu.Unlock()
}

// WarmPolicies loads all policies into cache (startup).
func (c *Cache) WarmPolicies() error {
	c.mu.RLock()
	load := c.loadAllPol
	c.mu.RUnlock()
	if load == nil {
		return nil
	}
	list, err := load()
	if err != nil {
		return err
	}
	c.mu.Lock()
	for i := range list {
		p := list[i]
		cp := p
		c.policies[p.ID] = &cp
	}
	c.mu.Unlock()
	return nil
}
